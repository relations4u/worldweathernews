# Iteration 1.1b — Hetzner Object Storage einrichten

**Übergabe-Prompt für Claude Code auf wwn-dev**

---

## Verwendung

Diesen Prompt entweder:

- **Parallel zu Iteration 1.1** in einer eigenen Claude-Code-Session
- **Nach Iteration 1.1** als Folge-Iteration

Empfehlung: parallel, weil unabhängig vom Frontend-Skelett. Storage ist
fertig, bevor Sveltia-Bild-Upload in Iteration 1.3 nötig wird.

```
# Falls neue Session:
ssh hwr@10.100.100.113
cd ~/repos/worldweathernews
claude code
# → kompletten Inhalt unten als ersten Prompt einfügen
```

---

## Prompt für Claude Code (Copy-Paste ab hier)

---

Hallo Claude Code. Wir richten **Hetzner Object Storage** als
S3-kompatiblen Bucket für die worldweathernews.com-Plattform ein.
Lies bitte zuerst:

1. `CLAUDE.md` im Repo-Root — die zentralen Spielregeln, Tech-Stack,
   Conventions
2. Externe Tracking-Dokumente (Pfad vom Maintainer):
   - `feature-decisions.md` (insbesondere A.13 Storage-Provider)
   - `feature-roadmap.md` (insbesondere Iteration 1.1b)
3. `infra/secrets/` — wie SOPS-Secrets im Repo organisiert sind

Sobald du diese gelesen hast, melde dich kurz mit einer Zusammenfassung,
damit ich sicher bin dass du den Kontext hast.

## Feature-Phase-Modus

**Wichtige Abweichung von der Setup-Phase:** In der Setup-Phase galt
"Maintainer committet selbst." In der Feature-Phase gilt **"Claude Code
committet nach expliziter Freigabe."**

Workflow:

1. Branch anlegen (`feat/iteration-1-1b-object-storage`)
2. Implementation in mehreren Commits auf dem Branch
3. **Vor jedem Commit: mich um Freigabe fragen**
4. Bei "OK" oder "commit" oder "merge": committen oder mergen
5. Bei "warte" oder "nochmal" oder "anders": warten und nachbessern
6. Push zu GitHub: erst nach explizitem "push" oder "PR aufmachen"

Kein eigenständiger Commit ohne Freigabe.

## Was diese Iteration liefert

Wir richten einen **S3-kompatiblen Bucket bei Hetzner Object Storage**
ein, der später für Bilder und Media-Assets der Plattform genutzt
wird. Plus die Domain-Konfiguration `media.worldweathernews.com` und
SOPS-Secrets im Repo.

Diese Iteration berührt **kein Application-Code** — es geht rein um
Infrastructure-Setup (Hetzner Console, Cloudflare DNS, SOPS-Secrets,
Test-Uploads). Application-Integration (Sveltia-Bild-Upload) kommt
in Iteration 1.3.

## Konzept-Hintergrund

**Warum Hetzner Object Storage:**

- Entscheidung in `feature-decisions.md` A.13 begründet
- S3-kompatibel (Sveltia kann später direkt drauf zugreifen)
- DSGVO-trivial: deutscher Anbieter, deutsche Server (Falkenstein)
- Pricing: €6.49/Monat ab 1. April 2026 (vorher €4.99)
- Inkludiert 1 TB Storage + 1 TB Egress pro Monat

**Architektur-Klarstellung:**

- Bucket-Endpoint: `https://fsn1.your-objectstorage.com`
- Public-Auslieferung: `https://media.worldweathernews.com` — **Reverse-Proxy
  über Caddy auf wwn-prod**, der den Host-Header auf den Hetzner-Bucket-Endpoint
  rewritet (Hetzner S3 routet nach Host-Header; Client-Host würde 400 BadRequest
  geben).
- Markdown referenziert Bilder via absolute URLs
- Sveltia (später) uploaded direkt via Pre-Signed-URLs

**Hinweis zu späterer Umstellung auf Cloudflare:** Die ursprüngliche Idee
war ein Cloudflare-Worker als Edge-Proxy (mit Cache-API). Wurde verworfen,
weil das Cloudflare-Dashboard im aktuellen Account die Workers-relevanten
Menüs ausgegraut zeigt (Subscription-Status unklar). Wenn das später
geklärt ist, ist die Umstellung mehrstufig denkbar:

1. Cloudflare-Worker als Cache vor dem Hetzner-Bucket (Caddy fällt weg).
2. Migration des Buckets zu Cloudflare R2 (würde A.13 in
   `feature-decisions.md` revidieren — DSGVO-Frage neu prüfen).

Beide Varianten plus die offene Subscription-Frage sind in
`docs/backlog.md` unter „CDN/Edge-Cache vor `media.worldweathernews.com`"
und „Cloudflare-Workers-Subscription-Status klären" festgehalten.

## Iterations-Plan

### Schritt 1 — Branch + Plan + Zugriff prüfen

1. Branch anlegen: `feat/iteration-1-1b-object-storage`
2. Verifikation: bist du auf `wwn-dev`? Bist du im richtigen Repo-Root?
   Maintainer-Identität korrekt?
3. Klär mit mir: welcher Hetzner-Account wird benutzt? Du brauchst
   keine direkten Hetzner-Credentials, du **weist mich an**, was zu
   tun ist (Bucket erstellen, Credentials abrufen) und ich liefere dir
   die Informationen.
4. Sobald alles OK: kurzen Plan zeigen wie du Schritte 2-7 angehen
   willst, dann Freigabe vom Maintainer abwarten

### Schritt 2 — Hetzner Bucket erstellen (Maintainer-Aktion!)

**Hier handelt der Maintainer in der Hetzner Cloud Console.** Du leitest
mich durch die Schritte:

1. Hetzner Cloud Console öffnen: https://console.hetzner.cloud/
2. Projekt wählen oder erstellen (Vorschlag: "worldweathernews")
3. Linkes Menü: "Object Storage" → "Add Bucket"
4. Bucket-Name: `media-worldweathernews-prod`
5. Location: **Falkenstein (FSN1)** (DSGVO-konform, deutscher Server)
6. Visibility: zunächst "Private" (Public-Read kommt später per Policy)
7. Bucket erstellen → Bestätigung warten

**Nicht** S3-Credentials direkt erstellen — das machen wir gleich
strukturiert in Schritt 3.

Maintainer bestätigt Erstellung mit Bucket-URL (z. B.
`https://media-worldweathernews-prod.fsn1.your-objectstorage.com`).
Du dokumentierst diese URL für die nächsten Schritte.

### Schritt 3 — S3-Credentials erstellen und in SOPS speichern

1. **Maintainer in Hetzner Console**:
   - Object Storage → S3 Credentials
   - "Generate New Credentials" → Name: `wwn-prod-cms-upload`
   - Access Key und Secret Key sicher kopieren (werden nur einmal angezeigt!)

2. **Du erstellst die SOPS-File-Struktur**:
   - `infra/secrets/production/media-storage.sops.env` als neue Datei
   - Inhalt-Schema:
     ```env
     S3_ENDPOINT=https://fsn1.your-objectstorage.com
     S3_REGION=fsn1
     S3_ACCESS_KEY=<wird vom Maintainer eingefügt>
     S3_SECRET_KEY=<wird vom Maintainer eingefügt>
     S3_BUCKET=media-worldweathernews-prod
     S3_PUBLIC_URL=https://media.worldweathernews.com
     ```

3. **Maintainer befüllt die Werte und verschlüsselt**:
   - In Repo-Root: `sops --encrypt --in-place infra/secrets/production/media-storage.sops.env`
   - Pre-commit-Hook `forbid-unencrypted-secrets` muss greifen wenn
     versehentlich Plaintext committed wird
   - Verifikation: `sops --decrypt infra/secrets/production/media-storage.sops.env`
     zeigt Plaintext zurück

4. **Backup**: Access Key und Secret Key zusätzlich in
   Bitwarden/1Password des Maintainers ablegen (Notiz: "wwn Hetzner
   Object Storage")

### Schritt 4 — CORS-Konfiguration auf dem Bucket

Sveltia (kommt in Iteration 1.3) wird direkt via Browser auf den Bucket
zugreifen. Dafür muss CORS konfiguriert sein.

CORS-Regeln (über Hetzner-Console oder via aws-cli):

```json
{
  "CORSRules": [
    {
      "AllowedOrigins": [
        "https://research.worldweathernews.com",
        "https://worldweathernews.com",
        "https://www.worldweathernews.com",
        "http://localhost:5173"
      ],
      "AllowedMethods": ["GET", "HEAD", "PUT"],
      "AllowedHeaders": ["*"],
      "ExposeHeaders": ["ETag"],
      "MaxAgeSeconds": 3600
    }
  ]
}
```

Wenn die Hetzner-Console das nicht direkt unterstützt: via aws-cli.
Du dokumentierst die exakten Befehle, der Maintainer führt sie aus.

```bash
aws s3api put-bucket-cors \
  --bucket media-worldweathernews-prod \
  --cors-configuration file://cors.json \
  --endpoint-url https://fsn1.your-objectstorage.com
```

### Schritt 5 — Bucket-Struktur anlegen

Initial-Verzeichnisse anlegen mit Platzhalter-Files:

```
media/
├── blog/             (Blog-Bilder, später per Slug strukturiert)
├── pages/            (Page-Bilder)
├── team/             (Team-Fotos)
└── site/             (Logo, Favicon, OG-Bilder, allgemein)
```

Test-Upload via aws-cli (Maintainer führt aus, du dokumentierst die
Befehle):

```bash
# Logo als ersten Test-Upload (du erstellst eine Test-Datei test.txt)
echo "wwn object storage initial test" > /tmp/test.txt

aws s3 cp /tmp/test.txt \
  s3://media-worldweathernews-prod/site/test.txt \
  --endpoint-url https://fsn1.your-objectstorage.com
```

### Schritt 6 — Public-Read, DNS und Caddy-Proxy

Wir brauchen einen Subset der Bucket-Inhalte als Public-Read, damit
der Browser sie via `media.worldweathernews.com` abrufen kann. Die
Auslieferung läuft über den Caddy-Stack auf wwn-prod, der den
Host-Header beim Forward auf den Bucket-Endpoint rewritet.

1. **Bucket-Policy** (Public-Read für die vier Asset-Verzeichnisse) —
   liegt bereits in `infra/object-storage/bucket-policy.json`:

   ```json
   {
     "Version": "2012-10-17",
     "Statement": [
       {
         "Sid": "PublicReadMedia",
         "Effect": "Allow",
         "Principal": "*",
         "Action": "s3:GetObject",
         "Resource": [
           "arn:aws:s3:::media-worldweathernews-prod/site/*",
           "arn:aws:s3:::media-worldweathernews-prod/blog/*",
           "arn:aws:s3:::media-worldweathernews-prod/pages/*",
           "arn:aws:s3:::media-worldweathernews-prod/team/*"
         ]
       }
     ]
   }
   ```

   Anwendung über aws-cli (Maintainer):

   ```bash
   aws s3api put-bucket-policy \
     --bucket media-worldweathernews-prod \
     --policy file://infra/object-storage/bucket-policy.json \
     --endpoint-url https://fsn1.your-objectstorage.com
   ```

2. **Cloudflare DNS-Eintrag** (Maintainer in Cloudflare Dashboard) —
   analog zu `research.` und `api.research.`:
   - DNS → worldweathernews.com → "Add record"
   - Type: CNAME
   - Name: `media`
   - Target: `home.worldweathernews.com`
   - Proxy: **AUS** (graue Wolke, DNS-only)
   - TTL: Auto

   Verlauf: `media.worldweathernews.com → home.worldweathernews.com →
gate.hw7.eu → Public-IP → Firewall-NAT 80/443 → Caddy auf wwn-prod`.

3. **Caddy-Block deployen** — der Block für `media.worldweathernews.com`
   liegt bereits in `infra/caddy/prod/Caddyfile`. Auf wwn-prod
   ausrollen:

   ```bash
   bash infra/deploy/deploy-caddy.sh
   # Achtung Bind-Mount-Inode-Falle: Skript macht restart caddy,
   # nicht up -d (siehe CLAUDE.md, Häufige Fallen)
   ```

   Caddy holt automatisch ein Let's-Encrypt-Zertifikat per HTTP-01-
   Challenge (Voraussetzung: Port 80 ist von außen offen, ist es
   bereits für die anderen Hostnames).

4. **Test**: nach DNS-Propagation und Cert-Issuing (1–5 Min):

   ```bash
   curl -I https://media.worldweathernews.com/site/test.txt
   # Erwartet: HTTP/2 200, Content-Type: text/plain
   ```

   Bei 4xx/5xx: Diagnose-Reihenfolge ist im neuen Runbook-Szenario
   „media.worldweathernews.com nicht erreichbar" dokumentiert
   (kommt mit Schritt 7).

### Schritt 7 — Repo-Doku ergänzen

In `docs/` neue oder erweiterte Files:

1. **`docs/media-storage.md`** neu anlegen:
   - Was wird wo gespeichert
   - Bucket-Struktur
   - Wie uploaded man manuell (für Maintainer-Bedarf)
   - Wie kommt Sveltia später dran (Verweis auf Iteration 1.3)
   - SOPS-Workflow (Verweis auf docs/secrets.md)

2. **`docs/runbook.md`** erweitern:
   - Neues Szenario: "Object Storage nicht erreichbar"
   - Diagnose-Schritte (Bucket existiert? CORS? DNS? Cloudflare?)
   - Rollback: temporär Sveltia auf Maintenance-Mode

3. **`CLAUDE.md`** "Wo finde ich was"-Tabelle erweitern:
   - Neuer Eintrag: Media-Bucket-Configuration → docs/media-storage.md
   - Neuer Eintrag: Bucket-Credentials → infra/secrets/production/media-storage.sops.env

4. **`docs/backlog.md`** erweitern um:
   - "Bucket-Backup-Strategie" (was passiert bei Hetzner-Outage?)
   - "Lifecycle-Policies" (Cleanup alter Bucket-Inhalte)
   - "Image-Optimierung-Pipeline" (WebP, AVIF, Größen-Varianten)

## Akzeptanzkriterien

- [ ] Bucket `media-worldweathernews-prod` in Falkenstein erstellt
- [ ] S3-Credentials in SOPS verschlüsselt im Repo
- [ ] Plaintext-Versuch wird vom Pre-commit-Hook geblockt
- [ ] CORS für die 4 Origins konfiguriert
- [ ] Test-Upload via aws-cli erfolgreich
- [ ] DNS: `media.worldweathernews.com` löst über `home` auf wwn-prod auf
- [ ] Caddy-Block ausgerollt und HTTPS aktiv (Let's-Encrypt-Zertifikat OK)
- [ ] Public-Read für `/site/test.txt` funktioniert via Browser
- [ ] `docs/media-storage.md` neu, Runbook + CLAUDE.md erweitert
- [ ] Backlog-Punkte für spätere Optimierungen dokumentiert
- [ ] Linter grün, Pre-commit-Hooks grün
- [ ] PR-Erstellung erst nach finalem OK des Maintainers

## Was du **noch nicht** baust

- **Sveltia-Bild-Upload-Integration** → Iteration 1.3
- **Pre-Signed-URL-Generierung im Worker** → Iteration 1.3
- **Image-Optimierung (WebP, Responsive)** → später, im Backlog
- **Lifecycle-Policies** → später, im Backlog

## Pre-Implementation-Tasks für den Maintainer

Bevor du loslegst, kläre ob:

- [ ] Hetzner Cloud Account einsatzbereit ist (Maintainer-Bestätigung
      vorhanden — laut letzter Nachricht: ja)
- [ ] Maintainer hat Zugriff auf Cloudflare-Dashboard für DNS-Edits
- [ ] aws-cli auf Maintainer-Mac installiert oder auf wwn-dev
      verfügbar (für Test-Uploads)
- [ ] Bitwarden/1Password bereit für Credential-Backup

Falls eines davon offen ist: pause die Implementation, kläre mit
Maintainer.

## Wenn etwas unklar ist

Frag mich. Insbesondere:

- Bei Hetzner-Console-Schritten: ich (Maintainer) führe sie aus
- Bei DNS-Konfiguration: ich führe Cloudflare-Edits aus
- Bei Credentials: niemals in Plaintext committen, niemals an Claude
  Code direkt senden — immer via SOPS

Lass uns loslegen. Bestätige mir kurz, dass du die Dokumente gelesen
hast, und schlag den ersten Schritt vor.
