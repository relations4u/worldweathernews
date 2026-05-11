# Feature-Phase Status — Snapshot 7. Mai 2026

Maintainer: Heinz W. Richter <hwr@relations4u.de>
Letzte Aktivität: 7. Mai 2026

Dieses Dokument ist ein **Wieder-Einstiegs-Punkt** — wenn die Session
unterbrochen wird, hier reinschauen für den schnellen Kontext.

Hauptdokumente in derselben Folder:

- `feature-decisions.md` — Atomare Entscheidungen mit Begründungen
- `feature-roadmap.md` — Konkrete Umsetzungs-Schritte je Iteration

---

## Wo wir stehen

### ✅ Setup-Phase abgeschlossen (Sessions 1-12, formal durch)

Komplett dokumentiert in der gepflegten `CLAUDE.md` im Repo. App-Stack
v0.0.2 läuft live auf wwn-prod (10.100.100.21), monitoring-stack auf
wwn-mon (10.100.100.22). Alle Public Smokes grün:

- https://worldweathernews.com (Apex)
- https://www.worldweathernews.com (Redirect)
- https://research.worldweathernews.com (Forschungs-Frontend)
- https://api.research.worldweathernews.com (Backend-API)

### ✅ Feature-Phase Track 1 — Konzept vollständig entschieden

Frontend, statische Seiten, CMS-Setup. Alle 18 A-Punkte entschieden.
Konkrete Umsetzung steht in `feature-roadmap.md` als 5 Iterationen
(1.1 Compliance-Skelett, 1.1b Object Storage, 1.2 mdsvex+i18n,
1.3 Sveltia einbinden, 1.4 Blog, 1.5 Editor-Onboarding).

### 🔄 Track 1 Iteration 1.1 + 1.1b ist die nächste konkrete Arbeit

**Übergabe an Claude Code in der `handover/`-Folder dokumentiert.**

### ⏳ Track 2 — Wetterdaten

Konzept-Diskussion noch nicht begonnen. Strategie: läuft parallel zu
Claude-Code-Implementation von Track 1, sobald die das tut. Drei offene
Architektur-Fragen (B.1 Datenquelle-Reihenfolge, B.2 Wetterkarten-
Strategie, B.4 Lizenzen) — siehe `feature-decisions.md` Abschnitt B.

### ⏳ Track 3 — KI-Agenten

Komplett offen. Sechs Agent-Rollen vorgeschlagen, drei werden Phase 1.
LLM-Provider-Wahl ist die größte offene Architektur-Frage. Siehe
`feature-decisions.md` Abschnitt C.

---

## Architektur-Stack (entschieden)

```
Frontend:           SvelteKit + TypeScript + Tailwind
Markdown-Pipeline:  mdsvex (Markdown + Svelte-Components)
i18n:               Paraglide.js + @inlang/paraglide-sveltekit
                    (Compile-time, typisiert, ~250B Bundle)
                    Strategy A: zentrale messages/de.json + en.json
CMS:                Sveltia (Git-based, Markdown im Frontend-Repo)
OAuth-Proxy:        Cloudflare Worker (sveltia-cms-auth)
Object Storage:     Hetzner Object Storage Falkenstein (€6.49/Monat)
Compute:            Self-hosted Proxmox (mit Migration-Triggern A.18)
Mail:               ProtonMail
DNS:                Cloudflare Free + Joker.com (DynDNS-Anker gate.hw7.eu)

Sprachen:           DE + EN parallel von Phase 1 an
Editorial:          Co-Autor:innen direkt-Commit (PR via Sveltia)
                    KI-Agenten via Draft-Review (kommt später Phase 3)

Backend (existiert):  Go + Chi + sqlc + pgx + Postgres+TimescaleDB+PostGIS
Workers (existiert):  Python 3.12 + uv (für GRIB/Wetterdaten-Adapter)
Caddy:                eigenständiger Stack auf wwn-prod, network_mode host
```

---

## Aktive offene Punkte

### Track 1 Pre-Implementation Tasks (vor Claude-Code-Sessions)

- [ ] **GitHub OAuth-App** anlegen für Sveltia (Settings → Developer → OAuth Apps)
- [ ] **Cloudflare Worker Account** vorbereiten (Workers + Pages aktivieren)
- [ ] **Hetzner Cloud Account** vorbereiten — vorhanden ✅
- [ ] **Anwalt** klar machen für Impressum/Datenschutz-Review (oder eRecht24)

### Track 2 Konzept-Fragen (offene Architektur-Diskussion)

- [ ] B.1 — Open-Meteo als Hello World oder direkt DWD?
- [ ] B.2 — Wetterkarten: selbst rendern oder einbinden?
- [ ] B.4 — Daten-Lizenzen: konkrete Attribution-Anforderungen

### Track 3 Konzept-Fragen (offene Architektur-Diskussion)

- [ ] C.1 — Welche 3 Agent-Rollen Phase 1?
- [ ] C.3 — LLM-Provider: Cloud / EU-Cloud / Self-hosted?
- [ ] C.4 — DSGVO-Strategie für Agent-Inputs
- [ ] C.5 — Budget-Rahmen für LLM-Calls

---

## Wichtige Kontext-Hinweise für Wieder-Einstieg

### Was Claude (Konzept) tut vs. Claude Code (Implementation)

**Claude (in dieser Chat-Session):**

- Architektur-Entscheidungen durchsprechen
- Trade-offs analysieren mit Web-Recherche
- Tracking-Files pflegen (`feature-decisions.md`, `feature-roadmap.md`)
- Übergabe-Prompts an Claude Code formulieren
- Track 2 + 3 Konzept entwickeln

**Claude Code (auf wwn-dev):**

- Tatsächliche Code-Implementation der Iterationen
- Tests, Linter, Builds
- PR-Erstellung (im Forschungs-Modus auch direkt-Commit nach Freigabe)
- Konkrete Bug-Fixes und Refinements

### Forschungs-Modus für Claude Code

Maintainer-Entscheidung 2026-05-07:

> „Da Forschung soll Claude Code im Moment auch committen — nach meiner Freigabe."

Workflow:

1. Claude Code arbeitet auf Feature-Branch
2. Maintainer reviewed im PR-Modus (visuell oder lokal)
3. Maintainer gibt explizit Freigabe ("OK zum committen", "merge", o.ä.)
4. Claude Code committet/merged direkt

Das ist eine Abweichung von der Setup-Phase-Regel (Maintainer committet
selbst). Begründung: Forschungs-Geschwindigkeit, weniger Reibung, klare
Maintainer-Kontrolle bleibt durch explizite Freigabe.

### Hetzner-Pricing-Realität

Ab 1. April 2026 sind alle Hetzner-Preise gestiegen. Object Storage
€4.99 → **€6.49/Monat**. Self-Hosting bleibt günstigste Option für
Phase 1 und 2.

### Hardware-Migration-Trigger (A.18)

Wechsel auf dedicated Server (Hetzner AX52 empfohlen, ~€75/Monat) wird
durch konkrete Bedingungen ausgelöst, nicht spekulativ. Re-Evaluation
alle 3 Monate. Aktuell: kein Trigger erfüllt, Self-Hosting bleibt.

---

## Letzte Diskussions-Punkte (chronologisch zur Session)

```
Mai 6  - CLAUDE.md mehrere Drift-Korrekturen (Pflege-Pass)
Mai 6  - GHCR-Token-Diagnose (Fine-grained PATs unterstützen
         Packages NICHT, Classic PAT nötig)
Mai 7  - Feature-Phase begonnen
Mai 7  - Tranche 1: Sveltia-CMS-Wahl
Mai 7  - Tranche 2: Sechs A-Punkte entschieden, Compliance-
         Pages-Liste vollständig, A.13/A.14/A.15 neu
Mai 7  - Tranche 3: Markdown-Pipeline mdsvex, Sprachen DE+EN,
         Editorial Workflow, Hosting-Strategie GCP verworfen
Mai 7  - Tranche 4: Hetzner/IONOS-Recherche für dedicated Server,
         Pricing-Update, A.18 Migration-Trigger neu
Mai 7  - Tranche 5: Paraglide-i18n entschieden, Track 1
         vollständig spezifiziert
Mai 7  - Status-Save plus Übergabe an Claude Code für Iteration 1.1
```

---

## Was als nächstes ansteht

**Maintainer-Vorgaben aus letzter Diskussion (7. Mai 2026):**

1. **Status speichern** ✅ (dieses Dokument)
2. **Übergabe-Prompts an Claude Code erstellen** ✅
   (siehe `handover/`-Folder)
3. **Hetzner Account vorhanden** ✅
4. **Sobald Coding parallel läuft**: Track 2 Diskussion fortsetzen
5. **Forschungs-Modus aktiv**: Claude Code committet nach expliziter
   Maintainer-Freigabe

**Reihenfolge nach diesem Status-Save:**

```
Sofort:
  1. Maintainer: Pre-Implementation Tasks abarbeiten
     - Hetzner Cloud Console öffnen, Object Storage aktivieren
     - aws-cli auf Mac/wwn-dev installieren oder testen
     - Bitwarden/1Password bereit für Credential-Backup
     - Anwalt-Termin oder eRecht24-Account für Compliance-Pages

  2. Claude Code parallel auf wwn-dev starten:
     - Session A: prompt-iteration-1-1.md (Compliance-Skelett)
     - Session B: prompt-iteration-1-1b.md (Object Storage)
     Beide laufen unabhängig, können in zwei Terminals parallel sein

  3. Claude (diese Konzept-Session): Track 2 Diskussion starten
     - sobald Coding-Sessions auf wwn-dev laufen
     - Domänen-Architektur, Worker-Pattern, Datenquellen-Reihenfolge

Mittelfristig:
  4. Track 1 Iterationen 1.2 - 1.5 (mdsvex, Sveltia, Blog, Onboarding)
     mit eigenen Übergabe-Prompts aus späteren Konzept-Sessions

  5. Track 3 Diskussion (KI-Agenten) starten, sobald Track 1 in
     Production-State ist und Track 2 die wichtigsten Daten-Adapter hat
```

**Track 2 Einstiegs-Optionen für nächste Konzept-Diskussion:**

```
Option A — Domänen-Architektur zuerst
  Daten-Modell für Stations, Observations, Forecasts.
  Beeinflusst alle Datenquellen-Adapter.
  Empfohlen wenn wir die Architektur sauber wollen.

Option B — Direkt Open-Meteo als Hello World
  Hands-on, schnelles Erfolgserlebnis.
  Empfohlen wenn wir lernen wollen, wie Wetterdaten in der Praxis
  reinkommen, bevor wir das große Modell durchdenken.

Option C — Pause bis Iteration 1.1 + 1.1b durch sind
  Maintainer fokussiert sich auf Frontend-Setup, dann Track 2
  in eigener Session weiter.
```

---

## Files in dieser Folder

```
status-snapshot.md         ← du liest ihn gerade
feature-decisions.md       ← alle Entscheidungen mit Begründung
feature-roadmap.md         ← Iteration-Schritte mit Akzeptanzkriterien
handover/
  README.md                ← Übergabe-Strategie an Claude Code
  prompt-iteration-1-1.md  ← Übergabe-Prompt Iteration 1.1
  prompt-iteration-1-1b.md ← Übergabe-Prompt Iteration 1.1b
```
