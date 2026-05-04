# @wwn/api-schema

OpenAPI-3.1-Schema und Code-Generators für die HTTP-API von
worldweathernews.com. Architektur-Entscheidung: siehe
[ADR-0001](../../docs/adr/0001-openapi-as-source-of-truth.md).

## Files

- `openapi.yaml` — Single Source of Truth (Endpoints, Schemas, Responses)
- `redocly.yaml` — Lint-Konfiguration (deaktiviert nicht passende Default-Regeln)
- `Makefile` — Validierung und Code-Generierung
- `package.json` — pnpm-Workspace-Member, devDeps `@redocly/cli` + `openapi-typescript`

## Workflow

Vom Repo-Root:

```bash
make gen          # validate + Go-Stubs + TS-Types regenerieren
make gen-check    # CI: regeneriert + git diff --exit-code (hard fail bei Drift)
```

Aus diesem Verzeichnis:

```bash
make validate     # nur redocly lint
make gen-go       # nur Go-Server-Stubs
make gen-ts       # nur TS-Types
make clean        # generierte Files löschen
```

## Was wird generiert wohin?

| Input          | Output                                   | Tool                 |
| -------------- | ---------------------------------------- | -------------------- |
| `openapi.yaml` | `apps/backend/internal/api/api.gen.go`   | `oapi-codegen` v2    |
| `openapi.yaml` | `apps/frontend/src/lib/api/types.gen.ts` | `openapi-typescript` |

Beide Output-Files sind in `.gitattributes` als `linguist-generated=true`
markiert, in `.golangci.yml` und `eslint.config.js` aus den Linters
ausgeschlossen, und im Pre-commit-Hook für Prettier ignoriert.

## Schema-Änderung — Workflow

1. `openapi.yaml` editieren.
2. `make gen` aus dem Repo-Root.
3. Backend: ggf. neue Methoden im `StrictServerInterface` implementieren
   (`apps/backend/internal/http/handler/api.go`).
4. Frontend: `client.ts` ergänzen (analog zu `ping`, `searchLocations`).
5. Tests anpassen, committen.
