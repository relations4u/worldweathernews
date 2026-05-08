import type { Reroute } from '@sveltejs/kit';
import { deLocalizeUrl } from '$lib/paraglide/runtime';

// Mappt lokalisierte URLs (z. B. /en/methodik) zurück auf die kanonischen
// SvelteKit-Routen, damit File-Based-Routing weiter wie bisher funktioniert.
export const reroute: Reroute = (request) => deLocalizeUrl(request.url).pathname;
