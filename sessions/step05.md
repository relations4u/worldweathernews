# Session 5 — SvelteKit-Frontend-Skelett

**Phase**: B (Services)
**Geschätzte Dauer**: 2 Stunden
**Vorbedingung**: Session 4 abgeschlossen, Backend antwortet auf `/api/v1/ping`.

## Ziel

Das Frontend ist ein SvelteKit-Projekt mit TypeScript, Tailwind und shadcn-svelte.
Es zeigt eine Hero-Page, die per Fetch das Backend kontaktiert und den Verbindungs-Status anzeigt.

Hot-Reload funktioniert im Container, Caddy routet `app.localhost` korrekt.

## Vor-Klärung

Falls noch nicht klar:

- **i18n-Library**: svelte-i18n vs. Paraglide (Inlang). Default-Empfehlung:
  TODO im Code lassen, Entscheidung in Session 12 oder erste Feature-Session.
  Jetzt nur DE und EN als statische Strings vorbereiten.

## Aufgaben

### 1. SvelteKit-Projekt initialisieren

In `apps/frontend/`. **Wichtig**: nicht-interaktiv aufsetzen.

Empfohlener Weg (kann sich ändern, prüfe aktuelle SvelteKit-Doku):

```bash
cd apps/frontend
pnpm dlx sv create . \
  --template minimal \
  --types ts \
  --no-add-ons \
  --install pnpm
```

Falls das CLI interaktiv hängt: **abbrechen, fragen**, manuell aufsetzen.

Alternative manuelle Aufsetzung:

- `package.json` mit den richtigen Dependencies
- `svelte.config.js` mit `@sveltejs/adapter-node`
- `vite.config.ts`
- `tsconfig.json` mit strict
- Basis-Verzeichnisstruktur `src/routes/` etc.

### 2. Adapter wechseln

Auf `@sveltejs/adapter-node` (Self-Hosting im Container):

```bash
pnpm remove @sveltejs/adapter-auto
pnpm add -D @sveltejs/adapter-node
```

`svelte.config.js`:

```js
import adapter from "@sveltejs/adapter-node";
import { vitePreprocess } from "@sveltejs/vite-plugin-svelte";

export default {
  preprocess: vitePreprocess(),
  kit: {
    adapter: adapter({
      out: "build",
      precompress: true,
      envPrefix: "WWN_FRONTEND_",
    }),
  },
};
```

### 3. Tailwind hinzufügen

```bash
pnpm dlx svelte-add@latest tailwindcss
```

oder manuell:

```bash
pnpm add -D tailwindcss postcss autoprefixer
pnpm dlx tailwindcss init -p
```

`tailwind.config.js` mit den richtigen content-Pfaden für SvelteKit.

`src/app.css` mit Tailwind-Imports und Layer-Setup.

### 4. shadcn-svelte initialisieren

```bash
pnpm dlx shadcn-svelte@latest init
```

Optionen:

- TypeScript: yes
- Style: default
- Base color: neutral (oder slate, frag mich)
- Tailwind CSS: ja
- Components-Alias: `$lib/components/ui`
- Utils-Alias: `$lib/utils`

Erste paar Komponenten installieren, die wir gleich brauchen:

```bash
pnpm dlx shadcn-svelte@latest add button card badge
```

### 5. Linting + Formatting

```bash
pnpm add -D eslint eslint-plugin-svelte @typescript-eslint/parser \
  @typescript-eslint/eslint-plugin prettier prettier-plugin-svelte \
  prettier-plugin-tailwindcss eslint-config-prettier
```

ESLint-Config (`eslint.config.js` für ESLint v9 flat-config):

- Svelte-Plugin
- TypeScript-Plugin
- Prettier-Compat

Prettier-Config (`.prettierrc.json`):

```json
{
  "useTabs": false,
  "tabWidth": 2,
  "singleQuote": true,
  "trailingComma": "none",
  "printWidth": 100,
  "plugins": ["prettier-plugin-svelte", "prettier-plugin-tailwindcss"]
}
```

### 6. Vitest hinzufügen

```bash
pnpm add -D vitest @vitest/ui jsdom @testing-library/svelte @testing-library/jest-dom
```

`vite.config.ts` erweitern für Test-Setup:

```ts
import { sveltekit } from "@sveltejs/kit/vite";
import { defineConfig } from "vitest/config";

export default defineConfig({
  plugins: [sveltekit()],
  test: {
    include: ["src/**/*.{test,spec}.{js,ts}"],
    environment: "jsdom",
    setupFiles: ["./src/test-setup.ts"],
  },
  server: {
    host: "0.0.0.0",
    port: 5173,
    strictPort: true,
    watch: { usePolling: true }, // wichtig für Container-Filesystem
  },
});
```

### 7. Verzeichnis- und Datei-Struktur

```
apps/frontend/
├── src/
│   ├── lib/
│   │   ├── api/
│   │   │   ├── client.ts        # Fetch-Wrapper
│   │   │   └── types.ts         # Wird in Session 7 generiert
│   │   ├── components/
│   │   │   └── ui/              # shadcn-svelte
│   │   ├── i18n/
│   │   │   ├── de.json
│   │   │   ├── en.json
│   │   │   └── index.ts         # TODO für Session N
│   │   └── utils.ts
│   ├── routes/
│   │   ├── +layout.svelte
│   │   ├── +layout.ts
│   │   └── +page.svelte
│   ├── app.css
│   ├── app.d.ts
│   ├── app.html
│   └── test-setup.ts
├── static/
│   └── favicon.png              # Platzhalter, ggf. einfaches SVG
├── tests/                       # Falls nötig für E2E später
├── .dockerignore
├── .gitignore
├── Dockerfile
├── eslint.config.js
├── package.json
├── pnpm-lock.yaml
├── postcss.config.js
├── svelte.config.js
├── tailwind.config.js
├── tsconfig.json
├── vite.config.ts
├── README.md
└── Makefile
```

### 8. `src/lib/api/client.ts`

Dünner Fetch-Wrapper:

```ts
import { PUBLIC_API_BASE_URL } from "$env/static/public";

export class ApiError extends Error {
  constructor(
    public status: number,
    public detail: string,
    public traceId?: string,
  ) {
    super(`API ${status}: ${detail}`);
  }
}

interface ApiOptions extends RequestInit {
  timeout?: number;
}

export async function apiFetch<T>(
  path: string,
  options: ApiOptions = {},
): Promise<T> {
  const { timeout = 10_000, ...rest } = options;
  const controller = new AbortController();
  const timer = setTimeout(() => controller.abort(), timeout);

  try {
    const response = await fetch(`${PUBLIC_API_BASE_URL}${path}`, {
      ...rest,
      signal: controller.signal,
      headers: {
        Accept: "application/json",
        "Content-Type": "application/json",
        ...rest.headers,
      },
    });

    if (!response.ok) {
      let detail = response.statusText;
      try {
        const body = await response.json();
        detail = body.detail || body.title || detail;
      } catch {
        /* nicht-JSON-Body, ignorieren */
      }
      throw new ApiError(response.status, detail);
    }

    return (await response.json()) as T;
  } finally {
    clearTimeout(timer);
  }
}

export async function ping(): Promise<{ message: string; traceId: string }> {
  return apiFetch("/api/v1/ping");
}
```

In Session 7 wird `types.ts` generiert und `client.ts` typsicher umgestellt.

### 9. `src/routes/+layout.svelte`

Basis-Layout mit Header (Logo-Placeholder als Text), Footer mit Year und einem
schlichten Theme-Switcher (light/dark via Tailwind class strategy).

```svelte
<script lang="ts">
    import '../app.css';
    let { children } = $props();
</script>

<div class="min-h-screen flex flex-col bg-background text-foreground">
    <header class="border-b">
        <div class="container mx-auto px-4 py-4 flex justify-between items-center">
            <a href="/" class="font-bold text-xl">worldweathernews</a>
            <nav class="text-sm text-muted-foreground">
                <span>WIP</span>
            </nav>
        </div>
    </header>

    <main class="flex-1 container mx-auto px-4 py-8">
        {@render children()}
    </main>

    <footer class="border-t text-sm text-muted-foreground">
        <div class="container mx-auto px-4 py-4">
            © {new Date().getFullYear()} worldweathernews — under construction
        </div>
    </footer>
</div>
```

(Verwendung von Svelte 5 `$props()` und `{@render children()}`. Falls die installierte
SvelteKit-Version noch Svelte 4 nutzt: angepasste Syntax.)

### 10. `src/routes/+page.svelte`

Hero mit Backend-Status:

```svelte
<script lang="ts">
    import { onMount } from 'svelte';
    import { ping } from '$lib/api/client';
    import { Badge } from '$lib/components/ui/badge';

    let status = $state<'pending' | 'ok' | 'error'>('pending');
    let traceId = $state<string | null>(null);
    let errorMsg = $state<string | null>(null);

    onMount(async () => {
        try {
            const result = await ping();
            status = 'ok';
            traceId = result.traceId;
        } catch (err) {
            status = 'error';
            errorMsg = err instanceof Error ? err.message : 'unknown error';
        }
    });
</script>

<section class="py-16 text-center">
    <h1 class="text-5xl font-bold tracking-tight">worldweathernews</h1>
    <p class="mt-4 text-lg text-muted-foreground">
        A global community for weather and climate observations.
    </p>
    <p class="mt-2 text-sm text-muted-foreground">
        Coming soon. Currently building the platform.
    </p>
</section>

<section class="text-center text-sm">
    {#if status === 'pending'}
        <Badge variant="outline">Connecting to backend…</Badge>
    {:else if status === 'ok'}
        <Badge>Backend connected</Badge>
        {#if traceId}
            <span class="ml-2 text-muted-foreground">Trace: {traceId}</span>
        {/if}
    {:else}
        <Badge variant="destructive">Backend offline</Badge>
        {#if errorMsg}
            <span class="ml-2 text-muted-foreground">{errorMsg}</span>
        {/if}
    {/if}
</section>
```

### 11. ENV-Setup

`.env.example` in `apps/frontend/`:

```bash
PUBLIC_API_BASE_URL=http://api.localhost
```

`src/app.d.ts` Types für ENV ergänzen falls SvelteKit es nicht automatisch macht.

### 12. Dockerfile

```dockerfile
# syntax=docker/dockerfile:1.7

FROM node:22-alpine AS deps
WORKDIR /app
COPY package.json pnpm-lock.yaml ./
RUN corepack enable && pnpm install --frozen-lockfile

FROM node:22-alpine AS builder
ARG VERSION=dev
ARG COMMIT=unknown
ARG PUBLIC_API_BASE_URL
ENV PUBLIC_API_BASE_URL=$PUBLIC_API_BASE_URL
WORKDIR /app
COPY --from=deps /app/node_modules ./node_modules
COPY . .
RUN corepack enable && pnpm build

FROM node:22-alpine AS runner
LABEL org.opencontainers.image.source="https://github.com/<org>/worldweathernews"
LABEL org.opencontainers.image.title="wwn-frontend"

WORKDIR /app
RUN addgroup -S app && adduser -S -G app app

COPY --from=builder /app/build ./build
COPY --from=builder /app/package.json ./package.json
COPY --from=deps /app/node_modules ./node_modules

USER app
EXPOSE 3000
ENV NODE_ENV=production
ENV HOST=0.0.0.0
ENV PORT=3000
HEALTHCHECK --interval=10s --timeout=3s --start-period=5s --retries=3 \
    CMD wget -q --spider http://localhost:3000/ || exit 1
CMD ["node", "build"]
```

### 13. `.dockerignore`

```
node_modules
.svelte-kit
build
.env
.env.*
!.env.example
.git
.github
README.md
*.md
tests
```

### 14. Compose-Integration

`infra/compose/compose.dev.yml`:

```yaml
frontend:
  image: node:22-alpine
  container_name: wwn-frontend
  command: sh -c "corepack enable && pnpm install --frozen-lockfile && pnpm dev --host 0.0.0.0"
  working_dir: /app
  volumes:
    - ../../apps/frontend:/app
    - frontend_node_modules:/app/node_modules
  environment:
    PUBLIC_API_BASE_URL: http://api.localhost
    HOST: 0.0.0.0
    PORT: 5173
  expose:
    - "5173"
  depends_on:
    - backend
```

`volumes`-Section:

```yaml
frontend_node_modules:
```

`Caddyfile`:

```caddy
app.localhost, localhost:80 {
    reverse_proxy frontend:5173 {
        # Vite HMR braucht WebSocket
        header_up Upgrade {>Upgrade}
        header_up Connection {>Connection}
    }
}
```

### 15. `apps/frontend/Makefile`

```makefile
.PHONY: install dev test lint check build clean

install:
	pnpm install

dev:
	pnpm dev --host 0.0.0.0

test:
	pnpm test

lint:
	pnpm lint

check:
	pnpm check

build:
	pnpm build

clean:
	rm -rf .svelte-kit build node_modules
```

Falls die package.json-Scripts noch nicht passen: anpassen.

### 16. Top-Level Makefile

```makefile
frontend-dev: ## Frontend hot-reload
	$(MAKE) -C apps/frontend dev

frontend-test: ## Frontend-Tests
	$(MAKE) -C apps/frontend test

frontend-lint: ## Frontend-Lint
	$(MAKE) -C apps/frontend lint
```

### 17. README

`apps/frontend/README.md`:

- Stack-Übersicht (SvelteKit + TS + Tailwind + shadcn-svelte)
- Lokale Befehle
- ENV-Variablen
- Wie eine neue shadcn-Komponente hinzugefügt wird
- TODO-Hinweise zu i18n und Theme

## Vorgehen (verbindlich)

1. Plan zeigen
2. Freigabe abwarten
3. Implementieren (Service-Setup → Tailwind → shadcn → Code → Docker → Compose)
4. **Wenn `sv create` interaktiv hängen will: ABBRECHEN, fragen.**
5. Verifizieren: `pnpm install && pnpm dev` lokal außerhalb von Docker — geht das?
6. Dann Compose: `docker compose up frontend backend postgres redis caddy -d`
7. Browser: http://app.localhost zeigen ("Backend connected")
8. `pnpm lint`, `pnpm check`, `pnpm test` ausführen, alle grün
9. Nicht committen

## Erfolgs-Kriterien

- [ ] `pnpm install` und `pnpm build` erfolgreich
- [ ] `pnpm test` läuft (mindestens ein Smoke-Test)
- [ ] `pnpm check` (svelte-check) grün
- [ ] `pnpm lint` grün
- [ ] http://app.localhost zeigt Hero + "Backend connected" Badge
- [ ] Hot-Reload funktioniert: Änderung an `+page.svelte` wird im Browser sofort sichtbar
- [ ] Build-Output unter `build/` lauffähig: `node build` startet einen Server
- [ ] Caddy routet WebSockets korrekt (Vite HMR funktioniert via app.localhost)

## Mögliche Stolpersteine

- **Svelte 4 vs. Svelte 5 Syntax**: Code oben nutzt Svelte-5-Runes (`$state`, `$props`).
  Falls install Svelte 4 liefert: Syntax anpassen oder gezielt Svelte 5 installieren.
- **HMR im Container**: braucht oft `usePolling: true` in Vite-Config oder explizites
  `chokidar`-Polling. Auf macOS/Windows mit Docker Desktop häufig Pflicht.
- **Caddy + WebSockets**: einfaches `reverse_proxy` reicht in modernen Caddy-Versionen,
  aber bei Problemen die `header_up`-Direktiven explizit setzen.
- **shadcn-svelte vs. Svelte 5**: prüfen ob die installierte Version Svelte 5 unterstützt.
  Aktuell läuft das gut, aber bei Mismatches: ältere shadcn-svelte-Version pinnen oder
  Komponenten manuell anpassen.
- **pnpm in Container**: `corepack enable` ist meist nötig, ältere Node-Images haben es
  noch nicht aktiviert.

## Was diese Session NICHT tut

- Keine Map-Komponente
- Keine echten Wetterdaten-Anzeige
- Kein i18n-Setup (nur vorbereiten)
- Keine Authentifizierung
- Kein Theming-Persist (nur Tailwind dark-mode-Vorbereitung)

## Suggested Commit-Message

```
feat(frontend): add SvelteKit skeleton with backend connectivity check
```
