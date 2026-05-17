import type { PageLoad } from './$types';
import { SAT_INDEX_URL, type SatIndex } from '$lib/config/satellite';

// CSR-only, analog /wetter (B.6 / S1): die index.json liegt auf
// media.worldweathernews.com (eigener Origin, server-seitig befüllt vom
// pyworkers-EUMETSAT-Worker). Kein SSR, damit der Frontend-Container den
// Public-Hostname nicht selbst auflösen muss.
export const ssr = false;

const EMPTY: SatIndex = {
	layer: 'ir108',
	source: 'eumetsat',
	attribution: '© EUMETSAT',
	bbox: { lonMin: 0, latMin: 0, lonMax: 0, latMax: 0 },
	frames: []
};

export const load: PageLoad = async ({ fetch }) => {
	try {
		const res = await fetch(SAT_INDEX_URL);
		if (!res.ok) return { index: EMPTY };
		const index = (await res.json()) as SatIndex;
		// Defensiv: Frames aufsteigend nach Zeit (Worker sortiert schon,
		// aber der Slider verlässt sich darauf).
		index.frames.sort((a, b) => a.time.localeCompare(b.time));
		return { index };
	} catch {
		return { index: EMPTY };
	}
};
