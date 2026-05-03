package handler

import (
	"net/http"

	chimw "github.com/go-chi/chi/v5/middleware"
)

type pingResponse struct {
	Message string `json:"message"`
	TraceID string `json:"traceId"`
}

// Ping ist ein simpler Demo-Endpoint, der die Trace-ID widerspiegelt.
func Ping() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, pingResponse{
			Message: "pong",
			TraceID: chimw.GetReqID(r.Context()),
		})
	}
}
