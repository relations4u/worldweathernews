# Feature-Phase Track 1 — Status

Pflege diese Datei am Ende jeder Iteration. Format analog zu
`sessions/STATUS.md` der Setup-Phase: Status + Datum + PR/Tag-Refs + Notizen.

Wenn ein Punkt offen ist, kurz festhalten was als nächstes ansteht und
woran man erkennt, dass es Zeit ist.

Status-Legende: ✅ Done · 🟡 In Progress · ⏳ Geplant · ❌ Blocked · ⏭ Skipped

Stand: 2026-05-11

---

## Live auf Production (wwn-prod)

- **App-Stack v0.0.4** läuft (backend, frontend, pyworkers, cms-auth, postgres, redis)
- **Caddy** als eigenständiger Stack mit Let's-Encrypt-Certs für:
  - `worldweathernews.com` (Apex → Frontend)
  - `www.worldweathernews.com` (301 → Apex)
  - `research.worldweathernews.com` (Forschungs-Frontend)
  - `api.research.worldweathernews.com` (Backend-API)
  - `cms-auth.worldweathernews.com` (Self-hosted OAuth-Proxy für Sveltia)
  - `media.worldweathernews.com` (Hetzner-Object-Storage-Read-Only-Proxy)
- **Monitoring-Stack** auf wwn-mon: Prometheus, Loki, Tempo, Grafana

Letzte Smoketests grün (curl mit GET, 200 jeweils): Apex, www-Redirect,
research, api/v1/ping, cms-auth Banner + /auth-Redirect zu GitHub.

---

## Track 1 — Frontend, Inhalte, CMS

### Iteration 1.1 — Hardcoded-Skelett mit Compliance-Pages

Status: ✅ Done
Datum: 2026-05-07
PR: #45 (Squash 80fc6ec)
Notizen: SvelteKit-Skelett + Tailwind/shadcn-Setup + Compliance-Pflicht-Seiten
(`/impressum`, `/datenschutz`, `/barrierefreiheit`, `/quellen-attribution`).
Cookie-Banner als Permanent-sticky-Komponente eingebaut.

### Iteration 1.1b — Hetzner Object Storage als Media-Bucket

Status: ✅ Done
Datum: 2026-05-08
PR: #44 (Squash 55af7e3)
Notizen: `media.worldweathernews.com` live, Hetzner Object Storage in
Falkenstein (FSN1), Caddy-Proxy mit Host-Rewrite (Hetzner routet
host-basiert, sonst 400). GET/HEAD-only-Proxy; Schreibzugriffe gehen
später über Pre-Signed-URLs direkt zum Bucket-Endpoint, nicht durch Caddy.
Doku: `docs/media-storage.md`.

### Iteration 1.2 — mdsvex-Pipeline + Paraglide i18n (DE/EN)

Status: ✅ Done
Datum: 2026-05-08
PRs: #46 (Squash 6730bbd), Follow-Up #32f571d (standalone pnpm-lock.yaml-Sync)
Notizen: Markdown-Pages mit Svelte-Components inline (Live-Charts,
DataSourceCards). Paraglide.js entschieden gegen svelte-i18n/sveltekit-i18n
(siehe A.17 in `feature-decisions.md`). Compile-time-Generierung typisierter
Messages aus `messages/{de-de,en}.json`. URL-Pattern `/de/…` und `/en/…` mit
Default `/de/`. Locale-Switcher im Header.

Falle: Standalone `apps/frontend/pnpm-lock.yaml` wurde übersehen — der
Docker-Build-Context zeigt auf `apps/frontend/`, nicht auf den Workspace-
Root. Behoben mit Folge-Commit `32f571d` (Regenerieren via temp-hide
von `pnpm-workspace.yaml`).

### Iteration 1.3a — Sveltia-Text-Edit + OAuth-Worker-Skelett

Status: ✅ Done (am 11. Mai 2026 abgelöst durch Self-Hosting, siehe unten)
Datum: 2026-05-08
PRs: #47 (Squash 3e0b474), #48 (Squash ace7c7f, base_url-Wiring)
Notizen: Sveltia-CMS als Loader unter `/admin/`, Backend GitHub via
postMessage-OAuth-Handshake. Image-Upload bewusst aus (`media_folder: ""`)
bis 1.3b. Editorial-Workflow (jeder Edit → PR) per `publish_mode:
editorial_workflow`. OAuth-Worker als Cloudflare-Worker `wwn-cms-auth`
im Account `hwr-06e` deployed (Maintainer-Hausaufgabe).

### Self-hosted OAuth-Proxy (Supersede A.4)

Status: ✅ Done
Datum: 2026-05-11
PRs: #58 (Squash 008b283, neuer Go-Service + Deploy), #59 (Squash 99a124b,
CF-Worker-Source aus Repo entfernt nach erfolgreichem Cutover)
Tag: v0.0.4 (live auf wwn-prod)
Notizen: Cloudflare-Worker abgelöst durch `apps/cms-auth/` — Go-Service
(Chi-Router + slog, distroless Image, HEALTHCHECK via Binary-Subcommand,
~170 LOC Logik 1:1 portiert vom CF-Worker). Container im App-Compose-
Stack auf wwn-prod, Bind `127.0.0.1:8090`, Caddy proxied unter
`cms-auth.worldweathernews.com`. Sveltia-Login real getestet, Token-
Handshake durch.

Begründung: Maintainer-Prinzip „Cloudflare-Abhängigkeit minimieren" —
DNS läuft bereits über CF, jeder zusätzliche kritische Pfad dort hebt
das Migrations-Risiko. CLAUDE.md-Entscheidungs-Eintrag unter
„Beantwortete Entscheidungen ab 2026-05-11", A.4 in
`feature-decisions.md` mit Supersede-Note markiert.

Falle gelernt (gh-Token): GitHub-API verweigert teilweise `mergePullRequest`
auf Workflow-Files ohne `workflow`-Scope auf dem PAT. Auf Retry hat es
trotzdem geklappt (Race oder lazy-loaded Scope-Check). Wenn das wieder
auftritt: `gh auth refresh -s workflow` oder Web-UI-Merge.

### Iteration 1.3b — Image-Pipeline

Status: ⏳ Geplant
Notizen: Pre-Signed-URL-S3-Upload, WebP-Konvertierung + responsive Sizes
(320/640/1280/1920) + EXIF-Strip. Per Memory-Regel
**self-hosted Container, kein CF-Worker** (analog zu cms-auth). Voraussetzung:
mindestens eine bildbedürftige Page in Sicht (Blog 1.4) — keine Eile.
Code-Stellen: zweiter Service `apps/cms-media-upload/` analog zu cms-auth,
plus `apps/frontend/static/admin/config.yml` (`media_folder`-Switch auf
neuen Pipeline-Endpoint).

### Iteration 1.4 — Blog-Collection

Status: ⏳ Geplant
Notizen: Sveltia-Collection für Blog-Artikel mit Tags/Kategorien, Listing-Page,
Single-Article-Layout. Voraussetzung: 1.3b für Bild-bedürftige Artikel —
sonst nur Text-Blog, was als Übergang OK ist.

### Iteration 1.5 — Editor-Onboarding-Doku

Status: ⏳ Geplant
Notizen: Component-Referenz, Edit-Workflow, Slug-Konventionen — alles
in `docs/cms.md` bereits angelegt, aber noch nicht auf reale Co-Autor:innen
getestet. Wartet auf erste externe Editor:innen.

---

## Track 2 — Wetterdaten

Status: ⏳ Konzept offen
Notizen: Keine Architektur-Entscheidungen getroffen. Drei offene B-Punkte
in `feature-decisions.md`: B.1 (Datenquelle-Reihenfolge — Open-Meteo vs.
DWD), B.2 (Wetterkarten — selbst rendern vs. einbinden), B.4 (Daten-
Lizenzen + Attribution).

Drei Einstiegs-Optionen aus `status-snapshot.md`:

- **A** — Domänen-Architektur zuerst (Stations, Observations, Forecasts).
  Empfohlen wenn wir die Architektur sauber wollen, bevor irgendein
  Adapter geschrieben wird.
- **B** — Direkt Open-Meteo als Hello World. Empfohlen wenn wir lernen
  wollen, wie Wetterdaten in der Praxis reinkommen, bevor wir das große
  Modell durchdenken.
- **C** — Pause bis Iteration 1.4 + 1.5 durch sind.

---

## Track 3 — KI-Agenten

Status: ⏳ Konzept offen
Notizen: Komplett unangetastet. Sechs Agent-Rollen vorgeschlagen (in
`feature-decisions.md` Abschnitt C), drei für Phase 1. Offene Punkte:
C.1 (welche 3 Phase-1-Agenten), C.3 (LLM-Provider — Cloud / EU-Cloud /
Self-hosted), C.4 (DSGVO-Strategie für Agent-Inputs), C.5 (Budget-Rahmen
für LLM-Calls).

Wartet darauf, dass Track 1 in Production-State ist und Track 2 die
wichtigsten Daten-Adapter hat.

---

## Querschnitt-Arbeit (zwischen den Iterationen)

### Dependabot-Triage 2026-05-11

Status: ✅ Done

Eine kombinierte Triage-Runde, 21 offene Dependabot-PRs auf 1 reduziert
(nur Tailwind v4 wurde manuell ersetzt und ist gemerged — siehe nächster
Punkt):

- **Toolchain-Major-Bumps (#3 golang 1.25→1.26, #8 python 3.12→3.14,
  #49 node 22→26)** — alle drei verstoßen gegen die Pin-Regel in CLAUDE.md.
  Geschlossen, plus Dependabot-Config gehärtet (PR #60), damit sie nicht
  zurückkommen: jede Docker-Ecosystem-Definition hat jetzt ein `ignore`-
  Pattern auf `version-update:semver-major|semver-minor` für die jeweilige
  Sprach-Base. Patch-Updates und non-toolchain-Base-Bumps (z. B. distroless)
  fliessen weiter. cms-auth als vierter Docker-Watcher hinzugefügt.
- **Frontend-Patches (#10 @typescript-eslint/parser, #11 typescript-eslint,
  #12 bits-ui, #14 postcss)** — alle scheitern in CI an Lockfile-Drift,
  weil Dependabot nur den Workspace-Lockfile regeneriert, nicht den
  standalone `apps/frontend/pnpm-lock.yaml`, den die Dockerfile-Stufe nutzt.
  Konsolidiert in **PR #62** mit beiden Lockfiles regeneriert via dem
  temp-hide-workspace-Verfahren aus 32f571d.
- **OTel-Bundle (Backend #54..#57 + Pyworkers #50..#52)** — 1.34 → 1.43
  bzw. ~=1.29 → >=1.29,<1.42. 5 von 7 gemerged, 2 (#56 otel, #57 otel/trace)
  hat Dependabot transitiv auto-geschlossen, weil sie nach dem Merge von
  #54 (otel/sdk) ohnehin via `go mod tidy` mit kamen.
- **Backend-Einzelner (#53 kin-openapi 0.137 → 0.138)** — sauber durch.
- **CI-Action-Major-Bumps (#2 upload-artifact, #4 checkout, #5 setup-node,
  #6 trivy-action, #7 pnpm/action-setup)** — alle gemerged. `gh pr merge`
  hat einmal mit „workflow scope missing" abgelehnt, beim Retry funktioniert.
- **Tailwind v4 (#13/#61)** — Dependabot proposed eine 1-Line-Bump, die
  CI sofort zerlegt hat. Manuell als richtiges Migrations-PR ersetzt (siehe
  nächster Punkt).

### Tailwind CSS v3 → v4 Migration

Status: ✅ Done
Datum: 2026-05-11
PR: #63 (Squash 6691358)
Tag: v0.0.4 deckt auch das ab
Notizen: `tailwindcss` 3.4.19 → 4.3.0, plus `@tailwindcss/vite` als
Vite-Plugin (PostCSS-Plugin-Pfad kollidiert mit Vite's eingebautem
postcss-import, das `@import "tailwindcss"` als relative Datei aufzulösen
versucht). `autoprefixer` raus — v4 macht Vendor-Prefixing über
Lightning CSS intern. Config zog von `tailwind.config.js` in einen
`@theme`-Block in `src/app.css`. Die shadcn-svelte-HSL-Variablen
(`--background`, `--primary`, …) sind via `--color-*` in den neuen
v4-Namespace gebrückt, damit Utility-Klassen wie `bg-primary` weiter
auflösen. `.dark`-Klassen-Strategie über `@custom-variant dark`
beibehalten. `tailwind.config.js` gelöscht.

Lokal verifiziert: `pnpm build`, `pnpm test` (5 passed), `pnpm check`
(0 errors), `pnpm lint`, plus Docker-Build + Container-Smoketest
(Apex-Page 200, 7771 Bytes HTML).

### Security-Scan-Workflow

Status: ❌ Failing (pre-existing, nicht durch Feature-Arbeit verursacht)
PR mit Backlog-Eintrag: #64 (Squash b2836cf)
Notizen: `.github/workflows/security-scan.yml` schlägt seit 11. Mai 2026
auf jedem Commit fehl, der Dependency-Manifeste oder Dockerfiles berührt.
Vier separate Befunde:

1. `Trivy upload-sarif` → 403 (privates Repo ohne GHAS — gleicher Pattern
   wie in `release.yml` lösbar: SARIF als Artifact hochladen statt zu
   code-scanning)
2. `pnpm audit --audit-level=high` Exit 1 (npm-Findings in `apps/frontend/`)
3. `pip-audit` Exit 1 (Python-Findings in `apps/pyworkers/`)
4. `govulncheck` Exit 3 (Backend — Call-Chains durch `template.Template.Execute`
   und `net.Resolver.*`, sehen nach False-Positives aus)

Detail-Triage in `docs/backlog.md` → Sicherheit. Zwischenzustand:
Workflow läuft, aber das rote Häkchen ist aktuell rauschend — Alarme
daraus sind wertlos, bis die vier Punkte einzeln triagiert sind oder
der Workflow temporär auf `continue-on-error: true` geht.

---

## Offene Punkte (nicht-PR)

- **Cosign-Verify im Deploy-Pull** (Backlog/Sicherheit) — Pipeline
  signiert beim Push, der Pull verifiziert nicht. Ansible-Task vor
  `docker compose pull` ergänzen.
- **Tracing-Sampler** für Production auf `ParentBased(TraceIDRatioBased(0.1))`
  umstellen — aktuell `AlwaysSample`. Backlog.
- **Backend-/Pyworkers-`/metrics`-Endpoint von wwn-mon erreichbar machen**
  — aktuell `127.0.0.1`-Bind, von wwn-mon nicht scrape-bar. Backlog.
- **node-exporter auf wwn-mon** ergänzen. Backlog.

Alle vier sind in `docs/backlog.md` mit Begründung und Code-Stellen
gepflegt.

---

## Refs

- Atomare Entscheidungen: `feature-decisions.md`
- Iterations-Schritte mit Akzeptanzkriterien: `feature-roadmap.md`
- Wieder-Einstiegs-Snapshot (älter, 7. Mai): `status-snapshot.md`
- Übergabe-Prompts an Claude Code: `prompt-iteration-1-1.md`,
  `prompt-iteration-1-1b.md`
- Setup-Phase-Status: `../STATUS.md`
- Setup-Phase-Doku: `../step01.md` .. `../step12.md`
