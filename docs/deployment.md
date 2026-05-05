# Deployment

<!-- TODO: Vollständige Deployment-Anleitung kommt in Session 11 (Ansible). -->

## Container-Registry: ghcr.io

Alle drei Service-Images werden über die [Release-Pipeline](../.github/workflows/release.yml)
nach `ghcr.io/relations4u/wwn-{backend,frontend,pyworkers}` gepusht. Trigger ist
ein Tag im Format `v*` auf `main`.

### Erstmaliges Setup nach erstem Release

1. Im GitHub-Profil → Packages → `wwn-backend`, `wwn-frontend`, `wwn-pyworkers`
2. Pro Package: Settings → Manage Actions Access → mindestens dieses Repository
   freigeben (oder "All repositories")
3. Visibility nach Wunsch setzen: Private (default), Internal oder Public
4. Auf dem Server für den Pull: PAT mit `read:packages` Scope erzeugen, dann

   ```bash
   echo "$GITHUB_TOKEN" | docker login ghcr.io -u <github-user> --password-stdin
   ```

   Ablage in `~/.docker/config.json` reicht für `docker compose pull`.

### Image-Pull lokal testen

```bash
docker pull ghcr.io/relations4u/wwn-backend:0.1.0
```

### Cosign-Signaturen verifizieren

Alle Images sind keyless via Sigstore signiert. Verifikation:

```bash
cosign verify ghcr.io/relations4u/wwn-backend:0.1.0 \
  --certificate-identity-regexp "^https://github.com/relations4u/worldweathernews/" \
  --certificate-oidc-issuer "https://token.actions.githubusercontent.com"
```

## Release auslösen

```bash
make release      # interaktiv: bump auswählen, signed Tag erzeugen, pushen
```

Der Workflow läuft 5–10 Minuten. Watch:
<https://github.com/relations4u/worldweathernews/actions/workflows/release.yml>

## Production-Compose

`infra/compose/compose.prod.yml` referenziert die ghcr.io-Images über
`${IMAGE_NAMESPACE}` und `${VERSION}`. Diese Variablen werden vom Ansible-Deploy
zur Laufzeit gesetzt. Manuell-Test:

```bash
IMAGE_NAMESPACE=relations4u VERSION=0.1.0 \
  POSTGRES_USER=... POSTGRES_PASSWORD=... POSTGRES_DB=... \
  WWN_DATABASE_URL=... WWN_REDIS_URL=... \
  WWN_PY_DATABASE_URL=... WWN_PY_REDIS_URL=... \
  docker compose -f infra/compose/compose.prod.yml up -d
```

Vollständige Variablenliste und Ansible-Integration: kommt in Session 11.
