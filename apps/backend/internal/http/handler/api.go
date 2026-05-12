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

// openMeteoAttribution ist der CC-BY-4.0-Pflichtsatz, den wir mit jedem
// Locations-Response ausliefern. Im Backend zentral gehalten, damit
// Frontend ihn nicht hardcoded haben muss (siehe docs/cms.md, B.4 in
// feature-decisions.md, Iteration 2.1).
const openMeteoAttribution = "Daten von Open-Meteo.com, CC BY 4.0"

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

// ListLocations gibt alle aktiven Locations zurück.
func (h *APIHandler) ListLocations(
	ctx context.Context, _ api.ListLocationsRequestObject,
) (api.ListLocationsResponseObject, error) {
	if h.queries == nil {
		return api.ListLocations200JSONResponse{
			Results:     []api.Location{},
			Attribution: openMeteoAttribution,
		}, nil
	}

	rows, err := h.queries.ListActiveLocations(ctx)
	if err != nil {
		return nil, fmt.Errorf("list active locations: %w", err)
	}

	results := make([]api.Location, len(rows))
	for i, r := range rows {
		results[i] = api.Location{
			Id:        r.ID,
			Slug:      r.Slug,
			Name:      r.Name,
			Country:   r.Country,
			Latitude:  r.Latitude,
			Longitude: r.Longitude,
			Timezone:  r.Timezone,
			Source:    r.Source,
		}
	}
	return api.ListLocations200JSONResponse{
		Results:     results,
		Attribution: openMeteoAttribution,
	}, nil
}

// GetLocationDetail liefert eine Location mit der letzten Beobachtung
// und dem 24-h-Forecast (jüngster Forecast-Run).
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

	// Letzte Beobachtung ist optional: wenn der Worker noch nichts geliefert
	// hat, geben wir current=null zurück (Pointer bleibt nil).
	var current *api.Observation
	obs, err := h.queries.GetLatestObservation(ctx, loc.ID)
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
		}
	case errors.Is(err, pgx.ErrNoRows):
		// kein Observation-Datensatz, current bleibt nil
	default:
		return nil, fmt.Errorf("get latest observation: %w", err)
	}

	forecastRows, err := h.queries.GetForecastNext24h(ctx, loc.ID)
	if err != nil {
		return nil, fmt.Errorf("get forecast next 24h: %w", err)
	}

	forecast := make([]api.ForecastEntry, len(forecastRows))
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

	return api.GetLocationDetail200JSONResponse{
		Location: api.Location{
			Id:        loc.ID,
			Slug:      loc.Slug,
			Name:      loc.Name,
			Country:   loc.Country,
			Latitude:  loc.Latitude,
			Longitude: loc.Longitude,
			Timezone:  loc.Timezone,
			Source:    loc.Source,
		},
		Current:     current,
		Forecast:    forecast,
		Attribution: openMeteoAttribution,
	}, nil
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
