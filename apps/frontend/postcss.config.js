/*
 * Tailwind v4 wird über @tailwindcss/vite in vite.config.ts geladen
 * (Vite-natives Plugin, vermeidet den postcss-import-Konflikt mit
 * `@import "tailwindcss"`). Diese Datei enthält daher keine
 * Tailwind-Referenz mehr, bleibt aber bestehen, damit Vite kein
 * implizites PostCSS-Setup zieht.
 */
export default {
	plugins: {}
};
