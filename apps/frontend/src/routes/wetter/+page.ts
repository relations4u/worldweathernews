import type { PageLoad } from './$types';
import { getLocationDetail, listLocations, type LocationDetail } from '$lib/api/client';

// Client-side rendering only: PUBLIC_API_BASE_URL ist auf den Public-Hostname
// gepinnt (api.research.worldweathernews.com), den der SSR-Frontend-Container
// nicht zwingend selbst auflöst. Für 2.1 bleibt die Wetter-Seite CSR-only.
// Ein SSR-Upgrade (separater INTERNAL-API-Hostname für die Server-Side-fetches)
// steht im docs/backlog.md.
export const ssr = false;

export const load: PageLoad = async ({ fetch }) => {
	const list = await listLocations({ fetch });
	const details: LocationDetail[] = await Promise.all(
		list.results.map((loc) => getLocationDetail(loc.slug, { fetch }))
	);
	return {
		details,
		attribution: list.attribution
	};
};
