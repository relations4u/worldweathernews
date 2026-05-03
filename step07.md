# Session 7 — OpenAPI-Schema und Type-Generation

**Phase**: C (CI/CD)
**Geschätzte Dauer**: 1-2 Stunden
**Vorbedingung**: Sessions 4 und 5 abgeschlossen, Backend und Frontend lauffähig.

## Ziel

Wir etablieren **OpenAPI 3.1** als Single Source of Truth für die API. Aus dem
Schema werden generiert:
- Go-Server-Stubs und -Models (oapi-codegen)
- TypeScript-Client-Types (openapi-typescript)

Das Backend implementiert die generierten Interfaces. Das Frontend nutzt die
generierten Types in seinem Fetch-Wrapper.

CI prüft via `git diff --exit-code`, dass die generierten Files aktuell sind.

Eine erste ADR dokumentiert die Entscheidung.

## Aufgaben

### 1. `packages/api-schema/openapi.yaml`

```yaml
openapi: 3.1.0
info:
  title: worldweathernews API
  version: 0.1.0
  description: |
    HTTP API für worldweathernews.com. Alle Endpunkte unter `/api/v1/`.
  contact:
    name: worldweathernews
    url: https://worldweathernews.com
  license:
    name: TBD

servers:
  - url: http://api.localhost
    description: Local development
  - url: https://api.worldweathernews.com
    description: Production

tags:
  - name: system
    description: System-Endpunkte (Health, Metrics)
  - name: locations
    description: Geographische Orte

paths:
  /api/v1/ping:
    get:
      operationId: ping
      tags: [system]
      summary: Connectivity check
      responses:
        '200':
          description: pong
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/PingResponse'

  /api/v1/locations:
    get:
      operationId: searchLocations
      tags: [locations]
      summary: Search locations by query string
      parameters:
        - name: q
          in: query
          required: true
          schema:
            type: string
            minLength: 2
            maxLength: 100
        - name: limit
          in: query
          required: false
          schema:
            type: integer
            minimum: 1
            maximum: 50
            default: 10
      responses:
        '200':
          description: List of matching locations
          content:
            application/json:
              schema:
                type: object
                properties:
                  results:
                    type: array
                    items:
                      $ref: '#/components/schemas/Location'
                required: [results]
        '400':
          $ref: '#/components/responses/BadRequest'

components:
  schemas:
    PingResponse:
      type: object
      properties:
        message:
          type: string
          example: pong
        traceId:
          type: string
      required: [message, traceId]

    Location:
      type: object
      properties:
        id:
          type: string
          format: uuid
        name:
          type: string
        countryCode:
          type: string
          minLength: 2
          maxLength: 2
          description: ISO 3166-1 alpha-2
        latitude:
          type: number
          format: double
          minimum: -90
          maximum: 90
        longitude:
          type: number
          format: double
          minimum: -180
          maximum: 180
        timezone:
          type: string
          description: IANA timezone, e.g. Europe/Berlin
      required: [id, name, countryCode, latitude, longitude]

    Problem:
      description: |
        Problem details object as per RFC 7807.
      type: object
      properties:
        type:
          type: string
          format: uri
          default: about:blank
        title:
          type: string
        status:
          type: integer
          minimum: 100
          maximum: 599
        detail:
          type: string
        instance:
          type: string
        traceId:
          type: string
      required: [title, status]

  responses:
    BadRequest:
      description: Bad request
      content:
        application/problem+json:
          schema:
            $ref: '#/components/schemas/Problem'
    NotFound:
      description: Resource not found
      content:
        application/problem+json:
          schema:
            $ref: '#/components/schemas/Problem'
    InternalServerError:
      description: Server error
      content:
        application/problem+json:
          schema:
            $ref: '#/components/schemas/Problem'
```

### 2. Validierung

`packages/api-schema/Makefile`:

```makefile
.PHONY: validate gen gen-go gen-ts clean

validate:
	pnpm dlx @redocly/cli@latest lint openapi.yaml

gen-go:
	@mkdir -p ../../apps/backend/internal/api
	go run github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen@latest \
		-package api \
		-generate types,chi-server,strict-server,spec \
		-o ../../apps/backend/internal/api/api.gen.go \
		openapi.yaml

gen-ts:
	@mkdir -p ../../apps/frontend/src/lib/api
	pnpm dlx openapi-typescript@latest openapi.yaml \
		-o ../../apps/frontend/src/lib/api/types.gen.ts

gen: validate gen-go gen-ts

clean:
	rm -f ../../apps/backend/internal/api/api.gen.go
	rm -f ../../apps/frontend/src/lib/api/types.gen.ts
```

Hinweis: `pnpm dlx` ist langsam wegen Download bei jedem Aufruf. Wenn das stört:
einmal global installieren oder als Dev-Dep im Frontend.

### 3. Top-Level Makefile

```makefile
gen: ## Generierten Code aus OpenAPI aktualisieren
	$(MAKE) -C packages/api-schema gen

gen-check: ## Prüft ob generierter Code aktuell ist (für CI)
	$(MAKE) gen
	git diff --exit-code -- \
		apps/backend/internal/api/api.gen.go \
		apps/frontend/src/lib/api/types.gen.ts || \
		(echo "Generated code is out of date. Run 'make gen' and commit." && exit 1)
```

### 4. Backend integrieren

#### 4.1 Generierung initial laufen lassen

```bash
make gen
```

Das erzeugt `apps/backend/internal/api/api.gen.go` mit:
- Schema-Types (`PingResponse`, `Location`, `Problem`)
- Chi-Server-Interface
- Strict-Server-Wrapper

#### 4.2 Handler anpassen

Bisheriger `Ping`-Handler wird zur Implementation des generierten Interfaces:

```go
// apps/backend/internal/http/handler/api.go

package handler

import (
    "context"
    "github.com/go-chi/chi/v5/middleware"
    "github.com/<org>/worldweathernews/apps/backend/internal/api"
)

type APIHandler struct {
    // dependencies, e.g. db pool, redis client
}

func NewAPIHandler(/*deps*/) *APIHandler {
    return &APIHandler{}
}

// Ping implements api.StrictServerInterface
func (h *APIHandler) Ping(ctx context.Context, request api.PingRequestObject) (api.PingResponseObject, error) {
    traceID := middleware.GetReqID(ctx)
    return api.Ping200JSONResponse{
        Message: "pong",
        TraceId: traceID,
    }, nil
}

// SearchLocations implements api.StrictServerInterface
func (h *APIHandler) SearchLocations(ctx context.Context, request api.SearchLocationsRequestObject) (api.SearchLocationsResponseObject, error) {
    // Stub: leere Liste, echte Implementation in Feature-Session
    return api.SearchLocations200JSONResponse{
        Results: []api.Location{},
    }, nil
}
```

#### 4.3 Router umbauen

```go
// apps/backend/internal/http/router.go

import (
    "github.com/<org>/worldweathernews/apps/backend/internal/api"
    "github.com/<org>/worldweathernews/apps/backend/internal/http/handler"
)

func NewRouter(cfg *config.Config, deps Deps) http.Handler {
    r := chi.NewRouter()

    // ... bestehende Middleware ...

    // System endpoints (nicht in OpenAPI, bewusst)
    r.Get("/health", handler.Health(deps))
    r.Get("/ready", handler.Ready(deps))
    if cfg.MetricsEnabled {
        r.Handle("/metrics", promhttp.Handler())
    }

    // OpenAPI-generierte Routen
    apiHandler := handler.NewAPIHandler( /* deps */ )
    strictHandler := api.NewStrictHandler(apiHandler, nil)
    api.HandlerFromMux(strictHandler, r)

    return r
}
```

#### 4.4 Tests anpassen

Bestehende Ping-Tests so umstellen, dass sie via HTTP gegen den fertigen Router
laufen und das JSON-Schema prüfen.

### 5. Frontend integrieren

#### 5.1 Generated Types einbinden

`apps/frontend/src/lib/api/client.ts` umbauen:

```ts
import type { paths, components } from './types.gen';
import { PUBLIC_API_BASE_URL } from '$env/static/public';

export type PingResponse = components['schemas']['PingResponse'];
export type Location = components['schemas']['Location'];
export type Problem = components['schemas']['Problem'];

export class ApiError extends Error {
    constructor(
        public status: number,
        public problem: Partial<Problem>
    ) {
        super(`API ${status}: ${problem.title ?? 'unknown'}`);
    }
}

async function request<T>(path: string, init: RequestInit = {}): Promise<T> {
    const res = await fetch(`${PUBLIC_API_BASE_URL}${path}`, {
        ...init,
        headers: { Accept: 'application/json', ...init.headers }
    });

    if (!res.ok) {
        let problem: Partial<Problem> = { title: res.statusText, status: res.status };
        try {
            problem = await res.json();
        } catch { /* ignore */ }
        throw new ApiError(res.status, problem);
    }

    return res.json() as Promise<T>;
}

export function ping(): Promise<PingResponse> {
    return request<PingResponse>('/api/v1/ping');
}

export function searchLocations(q: string, limit = 10): Promise<{ results: Location[] }> {
    const params = new URLSearchParams({ q, limit: String(limit) });
    return request(`/api/v1/locations?${params}`);
}
```

#### 5.2 Smoke-Test im UI

`+page.svelte`-Anpassung: zeigt zusätzlich einen Test der location-search
(Eingabe + Button) — leer ist OK, der Endpoint gibt ja `[]` zurück. Damit
hat man einen End-to-End-Test des Flows.

Optional, fragmich ob das den Scope sprengt — defaults: ja, klein machen,
zwei Zeilen mehr `+page.svelte`.

### 6. CI-Vorbereitung: Check-Script

`scripts/check-generated.sh`:

```bash
#!/usr/bin/env bash
set -euo pipefail

echo "==> Regenerating OpenAPI artifacts..."
make gen

echo "==> Checking for diff..."
if ! git diff --exit-code \
    apps/backend/internal/api/api.gen.go \
    apps/frontend/src/lib/api/types.gen.ts; then
    echo
    echo "✗ Generated files are out of date."
    echo "  Run 'make gen' locally and commit the changes."
    exit 1
fi

echo "✓ Generated files are up to date."
```

Ausführbar machen.

### 7. Linguist-Annotations

`.gitattributes` ergänzen:
```
apps/backend/internal/api/api.gen.go linguist-generated=true
apps/frontend/src/lib/api/types.gen.ts linguist-generated=true
```

Und im File-Header (oapi-codegen macht das schon, openapi-typescript auch — verifiziere).

### 8. ADR

`docs/adr/0001-openapi-as-source-of-truth.md` im MADR-Format:

```markdown
# 1. OpenAPI als Single Source of Truth für die HTTP-API

Date: 2026-XX-XX
Status: Accepted

## Context

Wir bauen eine HTTP-API in Go mit einem TypeScript-Frontend, das die API konsumiert.
Ohne formales Schema entsteht Drift zwischen Server und Client: Felder werden im Backend
umbenannt, das Frontend bricht; Validierung wird mehrfach implementiert; Doku ist
veraltet.

## Decision

Wir nutzen OpenAPI 3.1 als Single Source of Truth in `packages/api-schema/openapi.yaml`.
- Go-Server-Stubs werden mit `oapi-codegen` generiert (`internal/api/api.gen.go`)
- TypeScript-Types werden mit `openapi-typescript` generiert (`src/lib/api/types.gen.ts`)
- Der `make gen`-Workflow regeneriert beide
- CI prüft via `git diff --exit-code` dass generierte Files aktuell sind
- Schema-Änderungen erfordern Regenerierung und Test

## Consequences

**Positiv**:
- Keine Type-Drift zwischen Server und Client
- API-Doku immer aktuell (Schema ist die Doku)
- Externer Konsum der API später trivial (Schema kann veröffentlicht werden)
- B2B-API-Pläne sind ohne Mehrarbeit machbar

**Negativ**:
- Zusätzlicher Build-Schritt (`make gen`)
- Lernkurve für OpenAPI 3.1
- oapi-codegen-Aktualisierungen können generierte Types ändern → Migrations-Aufwand
- "Strict-Server"-Pattern zwingt zu spezifischem Style im Backend

## Alternatives Considered

- **gRPC + Protobuf**: leistungsfähiger, aber Browser-Support per gRPC-Web zusätzlich
  komplex; Frontend-DX schlechter.
- **GraphQL**: passt nicht zur überwiegend ressourcen-orientierten API; Tooling
  schwerer in Self-Hosting.
- **Hand-geschriebene Types beidseitig**: führt erfahrungsgemäß zu Drift.
```

### 9. README-Updates

- `packages/api-schema/README.md`: kurz erklären was hier passiert
- Top-Level `README.md`: Hinweis auf `make gen` und ADR-0001
- `apps/backend/README.md`: Hinweis dass Handler das generated Interface implementieren
- `apps/frontend/README.md`: Hinweis dass `client.ts` generated Types nutzt

## Vorgehen (verbindlich)

1. Plan zeigen
2. Freigabe abwarten
3. OpenAPI-Schema schreiben, validieren mit `redocly lint`
4. Generators einrichten und `make gen` einmal laufen lassen
5. Backend umbauen, Tests grün halten
6. Frontend umbauen, weiter "Backend connected" anzeigen
7. End-to-End-Test: Frontend → Backend mit den neuen Types
8. CI-Script `check-generated.sh` ausführen, sollte grün sein
9. ADR und READMEs schreiben
10. Nicht committen

## Erfolgs-Kriterien

- [ ] `make gen` läuft sauber (Validation + Generation)
- [ ] `apps/backend/internal/api/api.gen.go` existiert, ist gültiger Go-Code
- [ ] `apps/frontend/src/lib/api/types.gen.ts` existiert, ist gültiges TypeScript
- [ ] Backend-Tests grün, Lint grün
- [ ] Frontend-Check grün, Lint grün
- [ ] App im Browser zeigt weiter "Backend connected"
- [ ] `scripts/check-generated.sh` exit 0
- [ ] ADR-0001 vorhanden und konsistent
- [ ] Linguist markiert generierte Files (GitHub zeigt sie nicht in Diff-Stats)

## Mögliche Stolpersteine

- **oapi-codegen v2 vs v1**: API hat sich geändert. Aktuell v2.
- **Strict-Server-Pattern**: zwingt Funktions-Signaturen, deutlich anders als
  freie Handler. Doku lesen.
- **OpenAPI 3.1 vs 3.0**: oapi-codegen unterstützt 3.0 vollständig, 3.1
  zunehmend. Falls Probleme: auf 3.0.3 zurückfallen.
- **redocly lint warnings**: viele Warnings sind okay, errors müssen weg.
- **openapi-typescript-Output mit `paths`-Pfad-Keys**: ungewohnte Syntax,
  aber typsicher. Frontend muss damit umgehen.

## Was diese Session NICHT tut

- Keine echten Locations in der DB
- Keine Validation-Library im Backend (oapi-codegen kann das, aber wir
  bleiben minimal)
- Kein Authentifizierungs-Schema
- Kein API-Versioning v2

## Suggested Commit-Message

```
feat(api): add OpenAPI schema with Go and TypeScript code generation
```
