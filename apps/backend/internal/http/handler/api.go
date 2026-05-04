package handler

import (
	"context"

	chimw "github.com/go-chi/chi/v5/middleware"

	"github.com/relations4u/worldweathernews/apps/backend/internal/api"
)

// APIHandler implementiert das von oapi-codegen generierte api.StrictServerInterface.
// Hier landen alle Handler für die in openapi.yaml definierten Endpoints.
type APIHandler struct{}

// NewAPIHandler liefert einen neuen APIHandler. In künftigen Sessions kommen
// Storage-Dependencies (Pool, Redis-Client) als Felder dazu.
func NewAPIHandler() *APIHandler {
	return &APIHandler{}
}

// Ping spiegelt die Trace-ID der Request als JSON wider.
func (h *APIHandler) Ping(ctx context.Context, _ api.PingRequestObject) (api.PingResponseObject, error) {
	return api.Ping200JSONResponse{
		Message: "pong",
		TraceId: chimw.GetReqID(ctx),
	}, nil
}

// SearchLocations ist ein Stub bis Locations in der DB liegen — antwortet mit leerer Liste.
func (h *APIHandler) SearchLocations(_ context.Context, _ api.SearchLocationsRequestObject) (api.SearchLocationsResponseObject, error) {
	return api.SearchLocations200JSONResponse{
		Results: []api.Location{},
	}, nil
}
