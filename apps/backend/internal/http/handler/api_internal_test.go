package handler

import (
	"testing"

	"github.com/jackc/pgx/v5/pgtype"

	"github.com/relations4u/worldweathernews/apps/backend/internal/api"
)

// resolveSource ist privat im Handler-Paket — Tests hier in derselben
// Package-Datei, weil die Funktion die Kernlogik der source-Auswahl trägt
// und ohne DB testbar ist.
func TestResolveSource(t *testing.T) {
	dwdParam := api.GetLocationDetailParamsSource(sourceDwd)
	omParam := api.GetLocationDetailParamsSource(sourceOpenMeteo)
	withStation := pgtype.Text{String: "10384", Valid: true}
	noStation := pgtype.Text{Valid: false}
	emptyStation := pgtype.Text{String: "", Valid: true}

	cases := []struct {
		name       string
		param      *api.GetLocationDetailParamsSource
		dwdStation pgtype.Text
		want       string
	}{
		{
			name:       "explicit dwd param wins regardless of station",
			param:      &dwdParam,
			dwdStation: noStation,
			want:       sourceDwd,
		},
		{
			name:       "explicit open-meteo param wins even when station is set",
			param:      &omParam,
			dwdStation: withStation,
			want:       sourceOpenMeteo,
		},
		{
			name:       "no param + station present defaults to dwd",
			param:      nil,
			dwdStation: withStation,
			want:       sourceDwd,
		},
		{
			name:       "no param + no station defaults to open-meteo",
			param:      nil,
			dwdStation: noStation,
			want:       sourceOpenMeteo,
		},
		{
			name:       "no param + empty-but-valid station behaves like no station",
			param:      nil,
			dwdStation: emptyStation,
			want:       sourceOpenMeteo,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := resolveSource(tc.param, tc.dwdStation)
			if got != tc.want {
				t.Errorf("resolveSource = %q, want %q", got, tc.want)
			}
		})
	}
}

func TestAttributionFor(t *testing.T) {
	if got := attributionFor(sourceDwd); got != dwdAttribution {
		t.Errorf("attributionFor(dwd) = %q, want DWD attribution", got)
	}
	if got := attributionFor(sourceOpenMeteo); got != openMeteoAttribution {
		t.Errorf("attributionFor(open-meteo) = %q, want OM attribution", got)
	}
	// Defensive: unbekannte Quelle fällt auf OM zurück (sicherer Default,
	// damit Frontend nie eine leere Attribution rendert).
	if got := attributionFor("unknown"); got != openMeteoAttribution {
		t.Errorf("attributionFor(unknown) = %q, want OM fallback", got)
	}
}
