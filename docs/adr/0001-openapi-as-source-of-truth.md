# 1. OpenAPI als Single Source of Truth für die HTTP-API

Date: 2026-05-03
Status: Accepted

## Context

Wir bauen eine HTTP-API in Go mit einem TypeScript-Frontend, das die API
konsumiert. Ohne formales Schema entsteht Drift zwischen Server und Client:
Felder werden im Backend umbenannt, das Frontend bricht; Validierung wird
mehrfach implementiert; Doku ist veraltet.

## Decision

Wir nutzen **OpenAPI 3.1** als Single Source of Truth in
`packages/api-schema/openapi.yaml`.

- Go-Server-Stubs werden mit [`oapi-codegen`](https://github.com/oapi-codegen/oapi-codegen) v2
  generiert (`apps/backend/internal/api/api.gen.go`).
- TypeScript-Types werden mit
  [`openapi-typescript`](https://openapi-ts.dev) generiert
  (`apps/frontend/src/lib/api/types.gen.ts`).
- Der `make gen`-Workflow regeneriert beide.
- CI prüft via `scripts/check-generated.sh` (`git diff --exit-code`), dass die
  generierten Files aktuell sind.
- Schema-Änderungen erfordern Regenerierung und mindestens einen Test.

Das Backend implementiert das `StrictServerInterface` aus dem generierten Code.
Das Frontend referenziert die Schema-Types direkt aus `types.gen.ts`. System-
Endpoints (`/health`, `/ready`, `/metrics`) bleiben bewusst außerhalb des
OpenAPI-Schemas.

## Consequences

**Positiv**:

- Keine Type-Drift zwischen Server und Client
- API-Doku immer aktuell (Schema ist die Doku)
- Externer Konsum der API später trivial (Schema kann veröffentlicht werden)
- B2B-API-Pläne sind ohne Mehrarbeit machbar

**Negativ**:

- Zusätzlicher Build-Schritt (`make gen`)
- Lernkurve für OpenAPI 3.1
- `oapi-codegen`-Aktualisierungen können generierte Types ändern → Migrations-Aufwand
- "Strict-Server"-Pattern zwingt zu spezifischen Funktions-Signaturen im Backend

**Aktuelle Einschränkung**: `oapi-codegen` v2.4.1 markiert OpenAPI 3.1 als
"not yet fully supported". Die für uns genutzten Features (Schemas, Operations,
Parameters, Responses) funktionieren stabil. Bei Problemen ist der Fallback auf
3.0.3 ohne Schema-Änderungen möglich.

## Alternatives Considered

- **gRPC + Protobuf**: leistungsfähiger, aber Browser-Support per gRPC-Web
  zusätzlich komplex; Frontend-DX schlechter.
- **GraphQL**: passt nicht zur überwiegend ressourcen-orientierten API; Tooling
  schwerer in Self-Hosting.
- **Hand-geschriebene Types beidseitig**: führt erfahrungsgemäß zu Drift.
