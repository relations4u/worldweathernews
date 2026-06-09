# Media-Storage

Bilder, Logos und Page-Assets für worldweathernews.com liegen in einem
S3-kompatiblen Bucket bei **Hetzner Object Storage** (Region Falkenstein,
FSN1) und werden öffentlich über `https://media.worldweathernews.com`
ausgeliefert.

Provider-Entscheidung: siehe `feature-decisions.md` A.13 (DSGVO-trivial,
deutscher Anbieter, 1 TB Storage + 1 TB Egress für €6.49/Monat ab April 2026).

## Architektur

```
Browser
   │
   │  https://media.worldweathernews.com/<key>
   ▼
Cloudflare DNS (media → home → gate.hw7.eu, DNS-only)
   │
   ▼
Hardware-Firewall (NAT 80/443) → wwn-prod (10.100.100.70)
   │
   ▼
Caddy (network_mode: host)
   │  Host-Rewrite auf media-worldweathernews-prod.fsn1.your-objectstorage.com
   │  Nur GET/HEAD durchgereicht
   ▼
Hetzner Object Storage  (media-worldweathernews-prod, Falkenstein)
```

Schreibzugriffe (Sveltia-Bild-Upload, Iteration 1.3) laufen **nicht**
durch den Proxy, sondern direkt vom Browser per pre-signed URL gegen den
Hetzner-Endpoint. Caddy reicht ausschließlich GET/HEAD durch und
antwortet auf alles andere mit 405.

Warum der Caddy-Proxy nötig ist: Hetzner S3 routet Bucket-Requests nach
dem Host-Header. Der Client-Host `media.worldweathernews.com` ist kein
gültiger Bucket-Name → Hetzner antwortet 400 BadRequest. Caddy
rewritet den Host beim Forward auf den virtual-hosted Bucket-Endpoint.
Eine Cloudflare-Worker-Variante mit Edge-Cache ist im Backlog
festgehalten.

## Bucket-Layout

```
media-worldweathernews-prod/
├── site/      # Logo, Favicon, OG-Bilder, allgemeine Site-Assets
├── blog/      # Blog-Bilder, später per Slug strukturiert
├── pages/     # Page-Bilder (statische Inhaltsseiten)
├── team/      # Team-Fotos
└── sat/       # Satelliten-Frames + index.json (Iteration 2.4,
                #   server-seitig vom pyworkers-eumetsat-Job befüllt,
                #   read-only public für die /satellit-Route)
```

Public-Read ist via Bucket-Policy auf genau diese fünf Präfixe begrenzt
(`infra/object-storage/bucket-policy.json`). Andere Pfade
(`drafts/`, `internal/`, …) bleiben privat und sind nur mit Credentials
zugreifbar. Der `sat/`-Prefix wird **nicht** über CMS/Upload befüllt,
sondern allein vom `pyworkers`-EUMETSAT-Worker (Pfad A, Konzept-Session
2.4): rollierendes Fenster, ältere Frames werden vom Worker gelöscht.

CORS-Regeln (`infra/object-storage/cors.json`) erlauben GET/HEAD/PUT
von den drei Frontend-Origins (Apex, www, research) plus Vite-Dev-Server.
PUT ist für die spätere Direktupload-Variante via pre-signed URL
vorgesehen — aktuell ungenutzt.

## Manueller Upload (Maintainer-Workflow)

```bash
# Voraussetzung: Credentials aus dem SOPS-File geladen
export $(sops --decrypt infra/secrets/production/media-storage.env | xargs)

# Einzeldatei
aws s3 cp ./logo.svg \
  s3://media-worldweathernews-prod/site/logo.svg \
  --endpoint-url "$S3_ENDPOINT" \
  --content-type image/svg+xml

# Verzeichnis (z. B. Blog-Bilder eines Posts)
aws s3 cp ./post-2026-05-07/ \
  s3://media-worldweathernews-prod/blog/post-2026-05-07/ \
  --endpoint-url "$S3_ENDPOINT" \
  --recursive

# Verifikation: Public-Read funktioniert
curl -I "https://media.worldweathernews.com/site/logo.svg"
# Erwartet: HTTP/2 200
```

Content-Type explizit setzen — Hetzner errät nicht aus der Extension,
und ohne `Content-Type` hat der Browser Heuristik-Pech.

## Sveltia-Upload (später, Iteration 1.3)

Sveltia uploaded direkt vom Browser per pre-signed PUT-URL. Backend
generiert die URLs (Endpoint, Schlüssel, Header-Signatur), das Frontend
PUTet die Bytes. Der Caddy-Proxy ist dabei **nicht** im Pfad — Sveltia
ruft den Hetzner-Endpoint direkt, CORS-Regeln auf dem Bucket regeln das.

Details zum Backend-Endpoint und Sveltia-Konfiguration kommen mit
Iteration 1.3.

## Credentials und Secrets

S3-Credentials liegen verschlüsselt in
`infra/secrets/production/media-storage.env` (SOPS + age, siehe
`docs/secrets.md`):

```
S3_ENDPOINT=https://fsn1.your-objectstorage.com
S3_REGION=fsn1
S3_ACCESS_KEY=<vom Maintainer ausgestellt>
S3_SECRET_KEY=<vom Maintainer ausgestellt>
S3_BUCKET=media-worldweathernews-prod
S3_PUBLIC_URL=https://media.worldweathernews.com
```

Plaintext-Versionen sind durch den `forbid-unencrypted-secrets`-Pre-commit-
Hook blockiert. Backup der Credentials liegt zusätzlich im
Maintainer-Passwort-Manager.

## Cloudflare-Migration als Folge-Option

Der Backlog-Punkt „CDN/Edge-Cache vor `media.worldweathernews.com`"
beschreibt zwei mögliche Umstellungen, falls der Proxy-via-Heim-Anschluss
zum Engpass wird:

1. Cloudflare-Worker als Edge-Cache vor dem Hetzner-Bucket. Caddy fällt
   weg, DNS-Eintrag wechselt auf orange-cloud.
2. Migration des Buckets zu Cloudflare R2. A.13 müsste neu bewertet
   werden (DSGVO-Frage Cloudflare-als-US-Konzern), Worker bekäme dann
   die native R2-Binding-API statt fetch-Proxy.

Voraussetzung für beides: der Cloudflare-Workers-Subscription-Status
am Account muss geklärt werden — bei der Iteration-1.1b-Recherche
waren die relevanten Dashboard-Menüs ausgegraut.
