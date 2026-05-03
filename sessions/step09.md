# Session 9 — Release-Workflow und Container-Registry

**Phase**: C (CI/CD)
**Geschätzte Dauer**: 1-2 Stunden
**Vorbedingung**: Session 8 abgeschlossen, alle CI-Workflows grün.

## Ziel

Ein Tag im Format `v*` (z.B. `v0.1.0`) löst eine Release-Pipeline aus:

- Container-Images für alle drei Services werden gebaut (multi-arch)
- Mit cosign signiert (keyless via Sigstore)
- SBOM mit Syft erzeugt
- Trivy-Scan auf finalem Image
- Push nach `ghcr.io/<org>/wwn-{backend,frontend,pyworkers}`
- GitHub Release mit Auto-Generated Notes (git-cliff)

Lokal: `scripts/release.sh` als Helper für saubere Tagging.

## Vor-Klärung

- **Multi-Arch**: amd64 fix, arm64 abhängig vom Hosting. Wenn auf Hetzner CCX
  bleiben (amd64): arm64 weglassen. Wenn ARM-Server geplant: rein. Default-
  Empfehlung: beide bauen, ist im Buildx-Setup gratis.
- **GitHub-Repo Visibility**: ghcr.io-Images sind initial private, müssen
  manuell auf "Internal" oder "Public" gesetzt werden. Dokumentieren.

## Aufgaben

### 1. `cliff.toml` für git-cliff

```toml
[changelog]
header = """
# Changelog\n
"""
body = """
{% if version %}\
    ## [{{ version | trim_start_matches(pat="v") }}] - {{ timestamp | date(format="%Y-%m-%d") }}
{% else %}\
    ## [unreleased]
{% endif %}\
{% for group, commits in commits | group_by(attribute="group") %}
    ### {{ group | upper_first }}
    {% for commit in commits %}
        - {% if commit.scope %}**{{ commit.scope }}**: {% endif %}{{ commit.message | upper_first }}\
    {% endfor %}
{% endfor %}
"""
trim = true

[git]
conventional_commits = true
filter_unconventional = false
commit_parsers = [
    { message = "^feat", group = "Features" },
    { message = "^fix", group = "Bug Fixes" },
    { message = "^perf", group = "Performance" },
    { message = "^refactor", group = "Refactoring" },
    { message = "^docs", group = "Documentation" },
    { message = "^chore\\(deps\\)", group = "Dependencies" },
    { message = "^chore", skip = true },
    { message = "^ci", skip = true },
    { message = "^test", skip = true },
    { message = "^style", skip = true },
]
filter_commits = false
tag_pattern = "v[0-9]*"
```

### 2. `.github/workflows/release.yml`

````yaml
name: Release

on:
  push:
    tags:
      - "v*"

permissions:
  contents: write
  packages: write
  id-token: write # für cosign keyless
  attestations: write

env:
  REGISTRY: ghcr.io
  IMAGE_NAMESPACE: ${{ github.repository_owner }}

jobs:
  meta:
    name: Meta
    runs-on: ubuntu-latest
    outputs:
      version: ${{ steps.meta.outputs.version }}
      build_date: ${{ steps.meta.outputs.build_date }}
    steps:
      - id: meta
        run: |
          VERSION="${GITHUB_REF#refs/tags/v}"
          echo "version=$VERSION" >> "$GITHUB_OUTPUT"
          echo "build_date=$(date -u +%Y-%m-%dT%H:%M:%SZ)" >> "$GITHUB_OUTPUT"

  build:
    name: Build ${{ matrix.service }}
    runs-on: ubuntu-latest
    needs: meta
    strategy:
      fail-fast: false
      matrix:
        service:
          - { name: backend, context: apps/backend }
          - { name: frontend, context: apps/frontend }
          - { name: pyworkers, context: apps/pyworkers }
    steps:
      - uses: actions/checkout@v4

      - uses: docker/setup-qemu-action@v3
      - uses: docker/setup-buildx-action@v3

      - name: Log in to GHCR
        uses: docker/login-action@v3
        with:
          registry: ${{ env.REGISTRY }}
          username: ${{ github.actor }}
          password: ${{ secrets.GITHUB_TOKEN }}

      - name: Generate metadata
        id: docker_meta
        uses: docker/metadata-action@v5
        with:
          images: ${{ env.REGISTRY }}/${{ env.IMAGE_NAMESPACE }}/wwn-${{ matrix.service.name }}
          tags: |
            type=semver,pattern={{version}}
            type=semver,pattern={{major}}.{{minor}}
            type=semver,pattern={{major}}
            type=raw,value=latest,enable={{is_default_branch}}
            type=sha,format=short

      - name: Build and push
        id: build
        uses: docker/build-push-action@v6
        with:
          context: ${{ matrix.service.context }}
          platforms: linux/amd64,linux/arm64
          push: true
          tags: ${{ steps.docker_meta.outputs.tags }}
          labels: ${{ steps.docker_meta.outputs.labels }}
          cache-from: type=gha,scope=${{ matrix.service.name }}
          cache-to: type=gha,mode=max,scope=${{ matrix.service.name }}
          build-args: |
            VERSION=${{ needs.meta.outputs.version }}
            COMMIT=${{ github.sha }}
            BUILD_DATE=${{ needs.meta.outputs.build_date }}
          provenance: true
          sbom: true

      - name: Install cosign
        uses: sigstore/cosign-installer@v3

      - name: Sign image with cosign
        env:
          DIGEST: ${{ steps.build.outputs.digest }}
          TAGS: ${{ steps.docker_meta.outputs.tags }}
        run: |
          for tag in $TAGS; do
            cosign sign --yes "${tag}@${DIGEST}"
          done

      - name: Generate SBOM with Syft
        uses: anchore/sbom-action@v0
        with:
          image: ${{ env.REGISTRY }}/${{ env.IMAGE_NAMESPACE }}/wwn-${{ matrix.service.name }}@${{ steps.build.outputs.digest }}
          format: spdx-json
          output-file: sbom-${{ matrix.service.name }}.spdx.json

      - name: Trivy scan
        uses: aquasecurity/trivy-action@master
        with:
          image-ref: ${{ env.REGISTRY }}/${{ env.IMAGE_NAMESPACE }}/wwn-${{ matrix.service.name }}@${{ steps.build.outputs.digest }}
          format: sarif
          output: trivy-${{ matrix.service.name }}.sarif
          severity: CRITICAL,HIGH
          exit-code: 0
          ignore-unfixed: true

      - name: Upload Trivy SARIF
        uses: github/codeql-action/upload-sarif@v3
        if: always()
        with:
          sarif_file: trivy-${{ matrix.service.name }}.sarif
          category: trivy-${{ matrix.service.name }}

      - name: Upload SBOM artifact
        uses: actions/upload-artifact@v4
        with:
          name: sbom-${{ matrix.service.name }}
          path: sbom-${{ matrix.service.name }}.spdx.json

  release:
    name: GitHub Release
    runs-on: ubuntu-latest
    needs: [meta, build]
    steps:
      - uses: actions/checkout@v4
        with:
          fetch-depth: 0

      - name: Download SBOMs
        uses: actions/download-artifact@v4
        with:
          pattern: sbom-*
          path: sboms/
          merge-multiple: true

      - name: Generate Changelog
        uses: orhun/git-cliff-action@v4
        id: changelog
        with:
          config: cliff.toml
          args: --latest --strip header

      - name: Create GitHub Release
        uses: softprops/action-gh-release@v2
        with:
          tag_name: v${{ needs.meta.outputs.version }}
          name: v${{ needs.meta.outputs.version }}
          body: |
            ${{ steps.changelog.outputs.content }}

            ## Container Images

            ```
            ghcr.io/${{ env.IMAGE_NAMESPACE }}/wwn-backend:${{ needs.meta.outputs.version }}
            ghcr.io/${{ env.IMAGE_NAMESPACE }}/wwn-frontend:${{ needs.meta.outputs.version }}
            ghcr.io/${{ env.IMAGE_NAMESPACE }}/wwn-pyworkers:${{ needs.meta.outputs.version }}
            ```

            All images are signed with cosign. SBOMs are attached as release assets.
          files: |
            sboms/*.spdx.json
          generate_release_notes: false
          prerelease: ${{ contains(needs.meta.outputs.version, '-') }}
````

### 3. `scripts/release.sh`

```bash
#!/usr/bin/env bash
set -euo pipefail

# Helper für tagging und push.

if ! command -v gh >/dev/null; then
    echo "Error: gh CLI not installed."
    exit 1
fi

# Letzten Tag finden
LATEST_TAG="$(git describe --tags --abbrev=0 2>/dev/null || echo 'v0.0.0')"
LATEST_VERSION="${LATEST_TAG#v}"
echo "Latest tag: $LATEST_TAG"

# Komponenten zerlegen
IFS='.' read -r MAJOR MINOR PATCH <<< "$LATEST_VERSION"
PATCH="${PATCH%%-*}"   # Pre-release-Suffix entfernen

# Bump-Type abfragen
echo
echo "Select bump type:"
echo "  1) patch (v${MAJOR}.${MINOR}.$((PATCH + 1)))"
echo "  2) minor (v${MAJOR}.$((MINOR + 1)).0)"
echo "  3) major (v$((MAJOR + 1)).0.0)"
echo "  4) custom"
read -rp "> " choice

case "$choice" in
    1) NEW_VERSION="${MAJOR}.${MINOR}.$((PATCH + 1))" ;;
    2) NEW_VERSION="${MAJOR}.$((MINOR + 1)).0" ;;
    3) NEW_VERSION="$((MAJOR + 1)).0.0" ;;
    4)
        read -rp "Custom version (without 'v' prefix): " NEW_VERSION
        ;;
    *) echo "Invalid"; exit 1 ;;
esac

NEW_TAG="v${NEW_VERSION}"

# Zusammenfassung
echo
echo "About to create tag: $NEW_TAG"
echo "Branch: $(git branch --show-current)"
echo "Commit: $(git rev-parse --short HEAD)"
echo

# Working tree clean?
if [ -n "$(git status --porcelain)" ]; then
    echo "✗ Working tree not clean. Commit or stash changes first."
    exit 1
fi

# Auf main?
if [ "$(git branch --show-current)" != "main" ]; then
    read -rp "Not on main branch. Continue anyway? (y/N) " confirm
    [[ "$confirm" =~ ^[Yy]$ ]] || exit 1
fi

# Bestätigen
read -rp "Create and push $NEW_TAG? (y/N) " confirm
[[ "$confirm" =~ ^[Yy]$ ]] || exit 1

# Signed-Tag erzeugen
git tag -s "$NEW_TAG" -m "Release $NEW_TAG"
git push origin "$NEW_TAG"

echo
echo "✓ Tag $NEW_TAG pushed."
echo "  Watch the release pipeline:"
echo "  https://github.com/$(gh repo view --json nameWithOwner -q .nameWithOwner)/actions"
```

Ausführbar machen.

### 4. Compose Production-File

`infra/compose/compose.prod.yml`:

```yaml
# Production-Compose. Wird von Ansible deployed.
# Image-Tags werden via ENV oder vor-deploy-Substitution gesetzt.

services:
  postgres:
    image: timescale/timescaledb-ha:pg16
    container_name: wwn-postgres
    restart: unless-stopped
    environment:
      POSTGRES_USER: ${POSTGRES_USER}
      POSTGRES_PASSWORD: ${POSTGRES_PASSWORD}
      POSTGRES_DB: ${POSTGRES_DB}
    volumes:
      - postgres_data:/var/lib/postgresql/data
      - ./postgres-init:/docker-entrypoint-initdb.d:ro
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U $${POSTGRES_USER} -d $${POSTGRES_DB}"]
      interval: 10s
      timeout: 5s
      retries: 10
    deploy:
      resources:
        limits:
          memory: 2G
    logging:
      driver: json-file
      options:
        max-size: "50m"
        max-file: "5"

  redis:
    image: redis:7-alpine
    container_name: wwn-redis
    restart: unless-stopped
    command: redis-server --maxmemory 256mb --maxmemory-policy allkeys-lru
    healthcheck:
      test: ["CMD", "redis-cli", "ping"]
      interval: 10s
      timeout: 3s
      retries: 5
    deploy:
      resources:
        limits:
          memory: 512M

  backend:
    image: ghcr.io/${IMAGE_NAMESPACE}/wwn-backend:${VERSION}
    container_name: wwn-backend
    restart: unless-stopped
    environment:
      WWN_DATABASE_URL: ${WWN_DATABASE_URL}
      WWN_REDIS_URL: ${WWN_REDIS_URL}
      WWN_LOG_LEVEL: info
      WWN_LOG_FORMAT: json
      WWN_ENVIRONMENT: production
      WWN_HTTP_PORT: 8080
    depends_on:
      postgres: { condition: service_healthy }
      redis: { condition: service_healthy }
    deploy:
      resources:
        limits:
          memory: 512M
    logging:
      driver: json-file
      options:
        max-size: "50m"
        max-file: "5"

  frontend:
    image: ghcr.io/${IMAGE_NAMESPACE}/wwn-frontend:${VERSION}
    container_name: wwn-frontend
    restart: unless-stopped
    environment:
      PUBLIC_API_BASE_URL: https://api.worldweathernews.com
      NODE_ENV: production
      HOST: 0.0.0.0
      PORT: 3000
    deploy:
      resources:
        limits:
          memory: 256M

  pyworkers:
    image: ghcr.io/${IMAGE_NAMESPACE}/wwn-pyworkers:${VERSION}
    container_name: wwn-pyworkers
    restart: unless-stopped
    environment:
      WWN_PY_DATABASE_URL: ${WWN_PY_DATABASE_URL}
      WWN_PY_REDIS_URL: ${WWN_PY_REDIS_URL}
      WWN_PY_LOG_LEVEL: INFO
      WWN_PY_LOG_FORMAT: json
      WWN_PY_ENVIRONMENT: production
    depends_on:
      postgres: { condition: service_healthy }
      redis: { condition: service_healthy }
    deploy:
      resources:
        limits:
          memory: 512M

  caddy:
    image: caddy:2-alpine
    container_name: wwn-caddy
    restart: unless-stopped
    ports:
      - "80:80"
      - "443:443"
    volumes:
      - ./Caddyfile.prod:/etc/caddy/Caddyfile:ro
      - caddy_data:/data
      - caddy_config:/config
    depends_on:
      - frontend
      - backend

volumes:
  postgres_data:
  caddy_data:
  caddy_config:
```

`infra/caddy/Caddyfile.prod`:

```caddy
{
    email ops@worldweathernews.com
    log {
        output stdout
        format json
    }
}

worldweathernews.com, www.worldweathernews.com {
    encode zstd gzip
    reverse_proxy frontend:3000
}

api.worldweathernews.com {
    encode zstd gzip
    reverse_proxy backend:8080
    rate_limit {
        zone api 10r/s
    }
    @options method OPTIONS
    respond @options 204
}
```

(Rate-Limit-Modul ist bei Caddy nicht standardmäßig dabei. Wenn nicht enthalten:
weglassen oder via Build-Plugin ergänzen — TODO-Kommentar.)

### 5. ghcr.io-Setup-Dokumentation

`docs/deployment.md` Abschnitt:

````markdown
## Container-Registry: ghcr.io

Erstmaliges Setup nach erstem Release:

1. Im GitHub-Profil → Packages → wwn-backend, wwn-frontend, wwn-pyworkers
2. Settings → Manage Actions Access → "All repositories" oder explizit dieses
3. Visibility nach Wunsch: Private (default), Internal, Public
4. Bei Pull aus Server: PAT mit `read:packages` Scope erzeugen, im Server
   in `~/.docker/config.json` hinterlegen via `docker login ghcr.io`

Image-Pull lokal testen:

```bash
docker pull ghcr.io/<org>/wwn-backend:0.1.0
```
````

````

### 6. Top-Level Makefile

```makefile
release: ## Neuen Release-Tag erstellen (interaktiv)
	bash scripts/release.sh
````

### 7. README-Updates

- Top-Level: Hinweis auf `make release`
- Hinweis dass Container-Images auf ghcr.io liegen
- Versions-/Status-Badge für Latest-Release

## Vorgehen (verbindlich)

1. Plan zeigen
2. Freigabe abwarten
3. Workflow + Compose-Prod + Script schreiben
4. Yaml-Lint laufen lassen
5. Lokal: `bash scripts/release.sh` testen — bricht ab vor `git push`, frag
   wenn unsicher (Dry-Run-Mode wäre hier sinnvoll, optional implementieren)
6. Tatsächliches Tagging mache **ich** in einer separaten Action nach unserem Review
7. Wenn nach erstem Tag: Workflow live beobachten, Fehler iterativ fixen
8. Nicht committen

## Erfolgs-Kriterien

- [ ] `release.yml` ist syntaktisch valid (yamllint)
- [ ] cliff.toml generiert plausible Notes auf Test-Daten
- [ ] `scripts/release.sh` ist ausführbar, hat sinnvolle Prompts
- [ ] `compose.prod.yml` ist valide (`docker compose -f compose.prod.yml config`)
- [ ] Caddyfile.prod ist syntaktisch korrekt (`caddy validate --config Caddyfile.prod`)
- [ ] Nach erstem `v0.1.0`-Tag: alle drei Images auf ghcr.io vorhanden
- [ ] Alle Images sind cosign-signiert (verifizierbar mit `cosign verify`)
- [ ] SBOM und Trivy-Scan-Artefakte am Release angehängt
- [ ] GitHub-Release-Notes enthalten Changelog

## Mögliche Stolpersteine

- **cosign keyless**: braucht Sigstore-OIDC und korrekte permissions
  (`id-token: write`). In private Repos kostenlos, aber Public-Transparency-Log.
  Prüfe ob das ok ist (bedeutet: Image-Hashes sind im Public Log). Falls nicht:
  cosign mit Key-Pair als GitHub-Secret.
- **Multi-Arch und QEMU**: arm64-Builds dauern länger, ggf. 10+ Minuten.
- **GHA-Cache**: `cache-to: mode=max` kann groß werden. GHA-Limit beachten (10GB),
  notfalls auf `min` zurück.
- **Caddy rate_limit**: nicht im Standard-Build. Entweder mit `xcaddy` builden
  oder weglassen.
- **gpg signed tags**: `git tag -s` braucht eingerichtetes GPG-Signing oder ssh
  signing (modern). Falls nicht: `-a` statt `-s` als Fallback.

## Was diese Session NICHT tut

- Kein automatisches Deployment (kommt Session 11)
- Keine Staging-Promote-Pipeline
- Kein Image-Promotion (z.B. staging → prod tag)
- Keine Cosign-Verifikation auf dem Host (Pull-Side, später)

## Suggested Commit-Message

```
ci: add release workflow with multi-arch builds, cosign signing, sbom and trivy scanning
```
