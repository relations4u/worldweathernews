# 2. Go als Backend-Sprache

Date: 2026-05-03
Status: Accepted

## Context

Das Backend muss in zehn Jahren noch wartbar sein, mit kleinen Teams
laufen, im Container schlank deployen, und Server-Aufgaben (HTTP-API,
Hintergrund-Jobs, Rate-Limits) idiomatisch ausdrücken. Wir haben
mehrere Sprachen ernsthaft erwogen.

## Decision

Wir nutzen **Go 1.25** als Backend-Sprache (Modul
`github.com/relations4u/worldweathernews/apps/backend`).

Begleitende Kernentscheidungen:

- HTTP-Framework: **Chi v5** — idiomatisch, nah an `net/http`, kein
  Magic, gut testbar.
- DB-Access: **sqlc + pgx/v5** — typsichere Queries ohne ORM-
  Overhead, Migrations sprachunabhängig per goose.
- Logging: `slog` aus stdlib, JSON-strukturiert.
- Tracing: OpenTelemetry SDK + `otelchi` + `otelhttp`.
- Config: Viper mit `WWN_`-ENV-Prefix.

Das Backend hat **kein** ORM. Komplexere Queries werden in `.sql`-
Dateien geschrieben und von sqlc zu typsicheren Funktionen verarbeitet.

## Consequences

**Positiv**:

- Statisch verlinkte Binaries, distroless-Final-Image ~26 MB
- Erstklassiges Concurrency-Modell (Goroutines + Channels) für die
  vielen externen Wetter-API-Calls und Cache-Layer
- Toolchain-Reife: gofmt + goimports + golangci-lint v2 sind solide
- pgx ist die schnellste Postgres-Anbindung im Ökosystem
- Codegen aus OpenAPI funktioniert produktionsreif via oapi-codegen

**Negativ**:

- Boilerplate für Error-Handling (return-Pattern statt Exception)
- Schwächeres ORM-Ökosystem als z. B. Python — selbstgewählt, da wir
  bewusst kein ORM wollen, aber wer es gewohnt ist, fehlt es
- Generics sind erst seit 1.18 da; einige Libraries nutzen sie noch
  nicht idiomatisch
- pgx v5.9.2 zwingt die ganze Toolchain auf Go ≥ 1.25 — Versions-
  Pinning-Drift war einmal CI-Stolperstein

## Alternatives Considered

- **Node.js / TypeScript** — Frontend-Konsistenz wäre ein Bonus, aber
  Speicher- und Cold-Start-Footprint sind höher, Concurrency-Modell
  ist schwächer (Single-Thread-Event-Loop), und stdlib für Web ist
  dünn (Express-Ecosystem-Drift, Versions-Patch-Heuristik).
- **Python (FastAPI)** — wir nutzen Python ohnehin für Workers, also
  hätte das Konsolidierung gebracht. Aber: Performance pro Request
  schlechter, Container deutlich größer, kein statisches Linking,
  pydantic-Type-Drift in größeren Codebases.
- **Rust** — Top-Performance, aber Compile-Times und Lernkurve für
  einen Solo-Maintainer disqualifizieren es in dieser Phase.
  Wachstumspfad bleibt offen, falls wir später Hot-Path-Komponenten
  ausbauen.
- **Java/Kotlin** — JVM-Footprint und Build-Latenz sprechen dagegen
  für Self-Hosting auf kleinen VMs.
