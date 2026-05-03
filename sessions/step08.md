# Session 8 — GitHub Actions CI-Workflows

**Phase**: C (CI/CD)
**Geschätzte Dauer**: 2 Stunden
**Vorbedingung**: Sessions 4-7 abgeschlossen.

## Ziel

Bei jedem Push auf einen Branch und jedem Pull Request laufen die passenden
CI-Workflows automatisch. Sie testen, linten, bauen, scannen.

Concurrent läuft nur ein Run pro Branch (alte werden gecanceled).

Workflows sind über `paths`-Filter gezielt: Frontend-Änderung triggert nicht
Backend-CI.

## Vorbedingung GitHub-Repo

- Repository ist auf GitHub angelegt
- Optional: Branch-Protection-Rule für `main` aktiviert (manuell)

## Aufgaben

### 1. `.github/workflows/ci-backend.yml`

```yaml
name: CI Backend

on:
  pull_request:
    paths:
      - "apps/backend/**"
      - "packages/api-schema/**"
      - ".github/workflows/ci-backend.yml"
  push:
    branches: [main]
    paths:
      - "apps/backend/**"
      - "packages/api-schema/**"
      - ".github/workflows/ci-backend.yml"

concurrency:
  group: ci-backend-${{ github.ref }}
  cancel-in-progress: true

defaults:
  run:
    working-directory: apps/backend

jobs:
  lint:
    name: Lint
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version-file: apps/backend/go.mod
          cache-dependency-path: apps/backend/go.sum
      - uses: golangci/golangci-lint-action@v6
        with:
          version: latest
          working-directory: apps/backend
          args: --timeout=5m

  test:
    name: Test
    runs-on: ubuntu-latest
    services:
      postgres:
        image: postgis/postgis:16-3.4
        env:
          POSTGRES_USER: wwn
          POSTGRES_PASSWORD: wwn
          POSTGRES_DB: wwn_test
        ports:
          - 5432:5432
        options: >-
          --health-cmd "pg_isready -U wwn -d wwn_test"
          --health-interval 5s
          --health-timeout 5s
          --health-retries 10
      redis:
        image: redis:7-alpine
        ports:
          - 6379:6379
        options: >-
          --health-cmd "redis-cli ping"
          --health-interval 5s
          --health-timeout 5s
          --health-retries 5
    env:
      WWN_DATABASE_URL: postgres://wwn:wwn@localhost:5432/wwn_test?sslmode=disable
      WWN_REDIS_URL: redis://localhost:6379/0
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version-file: apps/backend/go.mod
          cache-dependency-path: apps/backend/go.sum
      - name: Run tests
        run: go test -race -coverprofile=coverage.out -covermode=atomic ./...
      - uses: actions/upload-artifact@v4
        with:
          name: backend-coverage
          path: apps/backend/coverage.out

  build:
    name: Build
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version-file: apps/backend/go.mod
          cache-dependency-path: apps/backend/go.sum
      - run: |
          CGO_ENABLED=0 go build \
            -trimpath \
            -ldflags="-s -w -X github.com/<org>/worldweathernews/apps/backend/internal/version.Commit=${{ github.sha }}" \
            -o /tmp/api ./cmd/api
      - uses: actions/upload-artifact@v4
        with:
          name: backend-binary
          path: /tmp/api
          retention-days: 7
```

### 2. `.github/workflows/ci-frontend.yml`

```yaml
name: CI Frontend

on:
  pull_request:
    paths:
      - "apps/frontend/**"
      - "packages/api-schema/**"
      - "packages/shared-types/**"
      - ".github/workflows/ci-frontend.yml"
  push:
    branches: [main]
    paths:
      - "apps/frontend/**"
      - "packages/api-schema/**"
      - "packages/shared-types/**"
      - ".github/workflows/ci-frontend.yml"

concurrency:
  group: ci-frontend-${{ github.ref }}
  cancel-in-progress: true

defaults:
  run:
    working-directory: apps/frontend

jobs:
  install:
    name: Install
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: pnpm/action-setup@v4
      - uses: actions/setup-node@v4
        with:
          node-version: 22
          cache: pnpm
          cache-dependency-path: apps/frontend/pnpm-lock.yaml
      - run: pnpm install --frozen-lockfile

  lint:
    name: Lint
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: pnpm/action-setup@v4
      - uses: actions/setup-node@v4
        with:
          node-version: 22
          cache: pnpm
          cache-dependency-path: apps/frontend/pnpm-lock.yaml
      - run: pnpm install --frozen-lockfile
      - run: pnpm lint

  check:
    name: svelte-check
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: pnpm/action-setup@v4
      - uses: actions/setup-node@v4
        with:
          node-version: 22
          cache: pnpm
          cache-dependency-path: apps/frontend/pnpm-lock.yaml
      - run: pnpm install --frozen-lockfile
      - run: pnpm check

  test:
    name: Test
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: pnpm/action-setup@v4
      - uses: actions/setup-node@v4
        with:
          node-version: 22
          cache: pnpm
          cache-dependency-path: apps/frontend/pnpm-lock.yaml
      - run: pnpm install --frozen-lockfile
      - run: pnpm test

  build:
    name: Build
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: pnpm/action-setup@v4
      - uses: actions/setup-node@v4
        with:
          node-version: 22
          cache: pnpm
          cache-dependency-path: apps/frontend/pnpm-lock.yaml
      - run: pnpm install --frozen-lockfile
      - run: pnpm build
        env:
          PUBLIC_API_BASE_URL: https://api.worldweathernews.com
```

### 3. `.github/workflows/ci-pyworkers.yml`

```yaml
name: CI PyWorkers

on:
  pull_request:
    paths:
      - "apps/pyworkers/**"
      - ".github/workflows/ci-pyworkers.yml"
  push:
    branches: [main]
    paths:
      - "apps/pyworkers/**"
      - ".github/workflows/ci-pyworkers.yml"

concurrency:
  group: ci-pyworkers-${{ github.ref }}
  cancel-in-progress: true

defaults:
  run:
    working-directory: apps/pyworkers

jobs:
  lint:
    name: Lint & Format
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: astral-sh/setup-uv@v3
        with:
          enable-cache: true
      - run: uv sync --frozen
      - run: uv run ruff check
      - run: uv run ruff format --check

  typecheck:
    name: mypy
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: astral-sh/setup-uv@v3
        with:
          enable-cache: true
      - run: uv sync --frozen
      - run: uv run mypy pyworkers

  test:
    name: pytest
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: astral-sh/setup-uv@v3
        with:
          enable-cache: true
      - run: uv sync --frozen
      - run: uv run pytest --cov=pyworkers --cov-report=xml
      - uses: actions/upload-artifact@v4
        with:
          name: pyworkers-coverage
          path: apps/pyworkers/coverage.xml
```

### 4. `.github/workflows/ci-shared.yml`

```yaml
name: CI Shared

on:
  pull_request:
  push:
    branches: [main]

concurrency:
  group: ci-shared-${{ github.ref }}
  cancel-in-progress: true

jobs:
  openapi-lint:
    name: OpenAPI Lint
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: pnpm/action-setup@v4
      - uses: actions/setup-node@v4
        with:
          node-version: 22
      - run: pnpm dlx @redocly/cli@latest lint packages/api-schema/openapi.yaml

  check-generated:
    name: Check Generated Code
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version-file: apps/backend/go.mod
          cache-dependency-path: apps/backend/go.sum
      - uses: pnpm/action-setup@v4
      - uses: actions/setup-node@v4
        with:
          node-version: 22
      - name: Run check
        run: bash scripts/check-generated.sh

  commitlint:
    name: Commit Messages
    runs-on: ubuntu-latest
    if: github.event_name == 'pull_request'
    steps:
      - uses: actions/checkout@v4
        with:
          fetch-depth: 0
      - uses: wagoid/commitlint-github-action@v6
        with:
          configFile: .commitlintrc.yaml

  yaml-lint:
    name: YAML Lint
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - run: pip install yamllint
      - run: yamllint -c .yamllint .

  markdown-links:
    name: Markdown Link Check
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: lycheeverse/lychee-action@v2
        with:
          fail: true
          args: --no-progress --exclude-mail '**/*.md'
```

### 5. `.commitlintrc.yaml`

```yaml
extends:
  - "@commitlint/config-conventional"
rules:
  type-enum:
    - 2
    - always
    - [feat, fix, chore, docs, refactor, test, perf, ci, style, build, revert]
  scope-enum:
    - 1
    - always
    - [backend, frontend, pyworkers, infra, api, ci, deps, docs, root]
```

### 6. `.github/workflows/security-scan.yml`

```yaml
name: Security Scan

on:
  schedule:
    - cron: "0 6 * * 1" # Mo 06:00 UTC
  workflow_dispatch:
  push:
    branches: [main]
    paths:
      - "**/go.mod"
      - "**/go.sum"
      - "**/pnpm-lock.yaml"
      - "**/uv.lock"

permissions:
  contents: read
  security-events: write

jobs:
  trivy-fs:
    name: Trivy Filesystem Scan
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: aquasecurity/trivy-action@master
        with:
          scan-type: fs
          format: sarif
          output: trivy-fs-results.sarif
          severity: CRITICAL,HIGH
          ignore-unfixed: true
      - uses: github/codeql-action/upload-sarif@v3
        if: always()
        with:
          sarif_file: trivy-fs-results.sarif

  govulncheck:
    name: govulncheck
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version-file: apps/backend/go.mod
      - run: |
          go install golang.org/x/vuln/cmd/govulncheck@latest
          cd apps/backend && govulncheck ./...

  pnpm-audit:
    name: pnpm audit
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: pnpm/action-setup@v4
      - uses: actions/setup-node@v4
        with:
          node-version: 22
      - run: cd apps/frontend && pnpm install --frozen-lockfile
      - run: cd apps/frontend && pnpm audit --audit-level=high
        continue-on-error: true

  pip-audit:
    name: pip-audit
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: astral-sh/setup-uv@v3
      - run: |
          cd apps/pyworkers
          uv pip compile pyproject.toml -o /tmp/requirements.txt
          uv pip install pip-audit
          uv run pip-audit -r /tmp/requirements.txt
        continue-on-error: true
```

### 7. `.github/dependabot.yml`

```yaml
version: 2
updates:
  - package-ecosystem: gomod
    directory: /apps/backend
    schedule:
      interval: weekly
      day: monday
    commit-message:
      prefix: chore
      include: scope
    open-pull-requests-limit: 5

  - package-ecosystem: npm
    directory: /apps/frontend
    schedule:
      interval: weekly
      day: monday
    commit-message:
      prefix: chore
      include: scope
    open-pull-requests-limit: 5
    groups:
      svelte:
        patterns:
          - "svelte*"
          - "@sveltejs/*"

  - package-ecosystem: pip
    directory: /apps/pyworkers
    schedule:
      interval: weekly
      day: monday
    commit-message:
      prefix: chore
      include: scope
    open-pull-requests-limit: 5

  - package-ecosystem: github-actions
    directory: /
    schedule:
      interval: weekly
      day: monday
    commit-message:
      prefix: ci
      include: scope

  - package-ecosystem: docker
    directory: /apps/backend
    schedule: { interval: weekly }
  - package-ecosystem: docker
    directory: /apps/frontend
    schedule: { interval: weekly }
  - package-ecosystem: docker
    directory: /apps/pyworkers
    schedule: { interval: weekly }
```

### 8. Branch-Protection-Dokumentation

`docs/deployment.md` (Anfang) erweitern:

```markdown
## Branch Protection

In den GitHub-Repo-Settings für `main` einstellen:

- Require pull request before merging
- Require status checks to pass:
  - CI Backend / lint, test, build
  - CI Frontend / lint, check, test, build
  - CI PyWorkers / lint, typecheck, test
  - CI Shared / openapi-lint, check-generated, commitlint
- Require branches to be up to date before merging
- Require linear history (optional, empfohlen für saubere Historie)
- Require signed commits (optional, empfohlen)
- Do not allow bypass for administrators (für Disziplin)
- Restrict force pushes
```

### 9. README-Badges

Top-Level `README.md` Badges einfügen:

```markdown
![CI Backend](https://github.com/<org>/worldweathernews/actions/workflows/ci-backend.yml/badge.svg)
![CI Frontend](https://github.com/<org>/worldweathernews/actions/workflows/ci-frontend.yml/badge.svg)
![CI PyWorkers](https://github.com/<org>/worldweathernews/actions/workflows/ci-pyworkers.yml/badge.svg)
```

## Vorgehen (verbindlich)

1. Plan zeigen
2. Freigabe abwarten
3. Workflows schreiben
4. Lokal nicht direkt testbar — Vorgehensweise:
   a) Branch erzeugen `chore/ci-setup`
   b) Files committen (ich committe selbst)
   c) Push, PR aufmachen
   d) Live Workflows beobachten
   e) Wenn rot: Fixes als weitere Commits in derselben PR
   f) Wenn alle grün: PR merge

5. Du machst die Fixes via Code-Edits, ich machst die Pushes.
6. Keine eigenständigen Pushes von Claude Code.

## Erfolgs-Kriterien

- [ ] Alle vier CI-Workflows existieren und sind syntaktisch valid
- [ ] Push auf einen Branch triggert die richtigen Workflows
- [ ] Lokal: `yamllint .github/workflows/` grün
- [ ] In GitHub Actions: alle Workflows laufen für eine Test-PR mindestens einmal
      durch und sind grün
- [ ] Concurrency-Gruppen funktionieren (alte Runs werden gecanceled)
- [ ] Path-Filter funktionieren (nur relevante Workflows triggern)
- [ ] Caching funktioniert (zweiter Run ist deutlich schneller)
- [ ] Dependabot-Config valide (GitHub zeigt im Insights-Tab)

## Mögliche Stolpersteine

- **GitHub-Actions-Versionen**: ändern sich. Aktuell empfohlen sind v4/v5/v6 wie
  oben angegeben. Bei Veröffentlichung der Workflow-Datei prüfen.
- **postgis-Image im CI**: starten dauert 15-30s, mit Healthcheck warten.
  TimescaleDB im CI ist optional und kann übersprungen werden für Standard-Tests.
- **commitlint und merge-commits**: bei Squash-Merges greifen die Conventional-
  Commits-Regeln, bei Merge-Commits manchmal nicht. Workflow ist für PR-Mode
  konfiguriert, das ist OK.
- **lychee-link-checker**: kann externe URLs als rate-limited reporten.
  `--accept` und Timeouts setzen falls nötig.
- **trivy SARIF-Upload**: braucht `security-events: write` Permission und
  Advanced Security in privaten Repos. In Public-Repos kostenlos.

## Was diese Session NICHT tut

- Kein Container-Build im CI (kommt Session 9)
- Kein Deploy aus CI (kommt Session 9 + 11)
- Kein Performance-Testing
- Kein E2E-Testing (kommt mit Playwright in Feature-Sessions)

## Suggested Commit-Message

```
ci: add github actions for backend, frontend, pyworkers, and shared checks
```
