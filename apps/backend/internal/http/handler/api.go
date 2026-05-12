package handler

import (
	"context"
	"errors"
	"fmt"

	chimw "github.com/go-chi/chi/v5/middleware"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/relations4u/worldweathernews/apps/backend/internal/api"
	"github.com/relations4u/worldweathernews/apps/backend/internal/storage/db"
)

// Attribution-Texte pro Datenquelle (Lizenz-Pflicht, siehe B.4 in
// feature-decisions.md). Werden im Response je nach gewählter Quelle
// ausgespielt; das Frontend zeigt zusätzlich pro Card ein Badge.
const (
	sourceOpenMeteo = "open-meteo"
	sourceDwd       = "dwd"

	openMeteoAttribution = "Daten von Open-Meteo.com, CC BY 4.0"
	dwdAttribution       = "Datenbasis: Deutscher Wetterdienst, eigene Bearbeitung"
)

// combinedAttribution wird auf List-Endpoints zurückgegeben, wo mehrere
// Quellen über die Locations hinweg vertreten sind. Frontend kann die
// per-Source-Attribution feiner pro Card rendern.
const combinedAttribution = dwdAttribution + " · " + openMeteoAttribution

// APIHandler implementiert das von oapi-codegen generierte
// api.StrictServerInterface.
type APIHandler struct {
	queries *db.Queries
}

// NewAPIHandler liefert einen neuen APIHandler. Pool darf nil sein
// (dev-Fallback wenn die DB nicht erreichbar ist); in dem Fall sind die
// Location-Endpoints leere Stubs.
func NewAPIHandler(pool *pgxpool.Pool) *APIHandler {
	h := &APIHandler{}
	if pool != nil {
		h.queries = db.New(pool)
	}
	return h
}

// Ping spiegelt die Trace-ID der Request als JSON wider.
func (h *APIHandler) Ping(ctx context.Context, _ api.PingRequestObject) (api.PingResponseObject, error) {
	return api.Ping200JSONResponse{
		Message: "pong",
		TraceId: chimw.GetReqID(ctx),
	}, nil
}

// ListLocations gibt alle aktiven Locations zurück, inklusive
// altitudeM, dwdStationId und availableSources pro Location.
func (h *APIHandler) ListLocations(
	ctx context.Context, _ api.ListLocationsRequestObject,
) (api.ListLocationsResponseObject, error) {
	if h.queries == nil {
		return api.ListLocations200JSONResponse{
			Results:     []api.Location{},
			Attribution: combinedAttribution,
		}, nil
	}

	rows, err := h.queries.ListActiveLocations(ctx)
	if err != nil {
		return nil, fmt.Errorf("list active locations: %w", err)
	}

	results := make([]api.Location, len(rows))
	for i, r := range rows {
		sources := append([]string{}, r.AvailableSources...)
		results[i] = api.Location{
			Id:               r.ID,
			Slug:             r.Slug,
			Name:             r.Name,
			Country:          r.Country,
			Latitude:         r.Latitude,
			Longitude:        r.Longitude,
			Timezone:         r.Timezone,
			Source:           r.Source,
			AltitudeM:        toIntPtr(r.AltitudeM),
			DwdStationId:     toStringPtr(r.DwdStationID),
			AvailableSources: &sources,
		}
	}
	return api.ListLocations200JSONResponse{
		Results:     results,
		Attribution: combinedAttribution,
	}, nil
}

// GetLocationDetail liefert eine Location mit der letzten Beobachtung
// und (für open-meteo) dem 24-h-Forecast. Mit dem Query-Param `source`
// kann zwischen den verfügbaren Quellen gewechselt werden; ohne Param
// gewinnt DWD, sofern eine Station hinterlegt ist.
func (h *APIHandler) GetLocationDetail(
	ctx context.Context, req api.GetLocationDetailRequestObject,
) (api.GetLocationDetailResponseObject, error) {
	if h.queries == nil {
		return notFoundResponse(ctx, req.Slug, "Database unavailable"), nil
	}

	loc, err := h.queries.GetLocationBySlug(ctx, req.Slug)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return notFoundResponse(ctx, req.Slug, "Unknown location slug"), nil
		}
		return nil, fmt.Errorf("get location by slug: %w", err)
	}

	source := resolveSource(req.Params.Source, loc.DwdStationID)

	// Letzte Beobachtung der gewählten Quelle: optional. Fehlt sie, bleibt
	// current=nil — die Location selbst wird trotzdem ausgeliefert.
	var current *api.Observation
	obs, err := h.queries.GetLatestObservationBySource(ctx, db.GetLatestObservationBySourceParams{
		LocationID: loc.ID,
		Source:     source,
	})
	switch {
	case err == nil:
		current = &api.Observation{
			ObservedAt:    obs.ObservedAt.Time,
			FetchedAt:     obs.FetchedAt.Time,
			Source:        obs.Source,
			Temperature:   toFloat32Ptr(obs.Temperature),
			Precipitation: toFloat32Ptr(obs.Precipitation),
			WindSpeed:     toFloat32Ptr(obs.WindSpeed),
			WindDirection: toIntPtr(obs.WindDirection),
			Pressure:      toFloat32Ptr(obs.Pressure),
			Humidity:      toFloat32Ptr(obs.Humidity),
		}
	case errors.Is(err, pgx.ErrNoRows):
		// kein Observation-Datensatz, current bleibt nil
	default:
		return nil, fmt.Errorf("get latest observation: %w", err)
	}

	// Forecast nur für Open-Meteo — DWD-Forecasts (MOSMIX-KML) kommen
	// erst in Iteration 2.2b. Für source=dwd liefern wir leeres Array.
	forecast := []api.ForecastEntry{}
	if source == sourceOpenMeteo {
		forecastRows, ferr := h.queries.GetForecastNext24h(ctx, loc.ID)
		if ferr != nil {
			return nil, fmt.Errorf("get forecast next 24h: %w", ferr)
		}
		forecast = make([]api.ForecastEntry, len(forecastRows))
		for i, fr := range forecastRows {
			forecast[i] = api.ForecastEntry{
				ForecastFor:   fr.ForecastFor.Time,
				RunAt:         fr.RunAt.Time,
				Temperature:   toFloat32Ptr(fr.Temperature),
				Precipitation: toFloat32Ptr(fr.Precipitation),
				WindSpeed:     toFloat32Ptr(fr.WindSpeed),
				WindDirection: toIntPtr(fr.WindDirection),
			}
		}
	}

	sources := append([]string{}, loc.AvailableSources...)
	return api.GetLocationDetail200JSONResponse{
		Location: api.Location{
			Id:               loc.ID,
			Slug:             loc.Slug,
			Name:             loc.Name,
			Country:          loc.Country,
			Latitude:         loc.Latitude,
			Longitude:        loc.Longitude,
			Timezone:         loc.Timezone,
			Source:           loc.Source,
			AltitudeM:        toIntPtr(loc.AltitudeM),
			DwdStationId:     toStringPtr(loc.DwdStationID),
			AvailableSources: &sources,
		},
		Current:     current,
		Forecast:    forecast,
		Attribution: attributionFor(source),
	}, nil
}

// resolveSource picks the effective source for a request. Wenn das Param
// gesetzt ist, gewinnt es. Sonst: DWD wenn die Location eine Station-ID
// hat, sonst Open-Meteo (Fallback für rein-OM-Locations).
func resolveSource(param *api.GetLocationDetailParamsSource, dwdStation pgtype.Text) string {
	if param != nil {
		return string(*param)
	}
	if dwdStation.Valid && dwdStation.String != "" {
		return sourceDwd
	}
	return sourceOpenMeteo
}

func attributionFor(source string) string {
	if source == sourceDwd {
		return dwdAttribution
	}
	return openMeteoAttribution
}

func notFoundResponse(ctx context.Context, slug, detail string) api.GetLocationDetail404ApplicationProblemPlusJSONResponse {
	traceID := chimw.GetReqID(ctx)
	detailMsg := fmt.Sprintf("%s: %q", detail, slug)
	return api.GetLocationDetail404ApplicationProblemPlusJSONResponse{
		NotFoundApplicationProblemPlusJSONResponse: api.NotFoundApplicationProblemPlusJSONResponse{
			Title:   "Not Found",
			Status:  404,
			Detail:  &detailMsg,
			TraceId: &traceID,
		},
	}
}

func toFloat32Ptr(v pgtype.Float8) *float32 {
	if !v.Valid {
		return nil
	}
	f := float32(v.Float64)
	return &f
}

func toIntPtr(v pgtype.Int4) *int {
	if !v.Valid {
		return nil
	}
	i := int(v.Int32)
	return &i
}

func toStringPtr(v pgtype.Text) *string {
	if !v.Valid {
		return nil
	}
	s := v.String
	return &s
}
