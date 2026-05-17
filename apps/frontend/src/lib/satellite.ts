// Pure Helfer für die Satelliten-Ansicht — geteilt/testbar, ohne
// Svelte-/MapLibre-Abhängigkeit (Muster wie lib/wind.ts aus 2.3).

import type { SatIndex } from '$lib/config/satellite';

type Corner = [number, number];

/**
 * Ecken-Koordinaten für die MapLibre-Image-Source aus der Geo-BBOX:
 * Reihenfolge top-left, top-right, bottom-right, bottom-left
 * (so erwartet MapLibre `image`-Source `coordinates`).
 */
export function bboxCorners(bbox: SatIndex['bbox']): [Corner, Corner, Corner, Corner] {
	const { lonMin, latMin, lonMax, latMax } = bbox;
	return [
		[lonMin, latMax],
		[lonMax, latMax],
		[lonMax, latMin],
		[lonMin, latMin]
	];
}

/**
 * Anzuzeigender Frame-Index: `selected` (Slider/Loop) oder — wenn
 * `null` — der neueste (letzte). Immer in [0, frameCount-1] geklemmt;
 * bei 0 Frames → 0.
 */
export function resolveFrameIndex(selected: number | null, frameCount: number): number {
	if (frameCount <= 0) return 0;
	const last = frameCount - 1;
	const idx = selected ?? last;
	return Math.min(last, Math.max(0, idx));
}
