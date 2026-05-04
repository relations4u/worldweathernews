# wwn-frontend

SvelteKit 2 + Svelte 5 (Runes) + TypeScript + TailwindCSS + shadcn-svelte
(als manuell verwaltete Komponenten unter `src/lib/components/ui/`).

## Lokal starten

Über den Compose-Stack (empfohlen):

```bash
make dev    # vom Repo-Root
# Browser: http://app.localhost
```

Direkt ohne Container:

```bash
cd apps/frontend
pnpm install
pnpm dev      # http://localhost:5173
```

`PUBLIC_API_BASE_URL` muss zeigen auf das Backend (Default `http://api.localhost`).

## Befehle

| Kommando       | Was                                       |
| -------------- | ----------------------------------------- |
| `pnpm dev`     | Vite-Dev-Server mit HMR                   |
| `pnpm build`   | Produktions-Build (`build/`)              |
| `pnpm preview` | Preview des Builds                        |
| `pnpm test`    | Vitest (Unit-Tests in `src/**/*.test.ts`) |
| `pnpm check`   | `svelte-check` Type-Check                 |
| `pnpm lint`    | ESLint + `prettier --check`               |
| `pnpm format`  | Prettier-Auto-Format                      |

## ENV-Variablen

| Variable              | Default                    | Wirkung                     |
| --------------------- | -------------------------- | --------------------------- |
| `PUBLIC_API_BASE_URL` | `http://api.localhost`     | Backend-URL für `apiFetch`  |
| `WWN_FRONTEND_*`      | _(siehe svelte.config.js)_ | adapter-node `envPrefix`    |
| `HOST`, `PORT`        | `0.0.0.0`, `3000` (prod)   | adapter-node Listen-Adresse |

## Neue shadcn-svelte-Komponente hinzufügen

shadcn-svelte funktioniert über **Code-Copying**, nicht über npm-Install. Für
diesen Stub haben wir `Badge` manuell unter `src/lib/components/ui/badge/`
abgelegt. Weitere Komponenten:

```bash
# automatisch (interaktiv, kann hängen):
pnpm dlx shadcn-svelte@latest add button card

# oder manuell aus https://shadcn-svelte.com/docs/components kopieren,
# dabei `cn()` aus `$lib/utils` verwenden und `tailwind-variants` für
# Variants — siehe Badge als Vorlage.
```

## API-Types

`src/lib/api/types.gen.ts` wird aus
[`packages/api-schema/openapi.yaml`](../../packages/api-schema/openapi.yaml)
generiert (`openapi-typescript`, siehe
[ADR-0001](../../docs/adr/0001-openapi-as-source-of-truth.md)). Schema-
Änderungen → vom Repo-Root `make gen`. `client.ts` re-exportiert die
relevanten Typen (`PingResponse`, `Location`, `Problem`).

Die generierte Datei ist als `linguist-generated=true` markiert und aus
ESLint/Prettier ausgeschlossen — nicht von Hand editieren.

## TODOs

- **i18n-Library** (svelte-i18n vs. Paraglide vs. Inlang) — Entscheidung in
  einer späteren Session. Bis dahin Übersetzungs-JSONs unter
  `src/lib/i18n/{de,en}.json`, in Komponenten direkt englische Strings.
- **Theme-Persist** (light/dark) — bisher nur Tailwind `darkMode: ['class']`
  vorbereitet, kein Persist-Mechanismus.
