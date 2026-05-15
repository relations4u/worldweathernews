// Wind-Helfer, geteilt zwischen WeatherCard und StationsMap.
//
// `windDirection` ist meteorologisch die Richtung, AUS der der Wind
// kommt (0° = Nord, 90° = Ost). Der Kompass-Text zeigt diese Herkunft;
// der Karten-Pfeil zeigt, WOHIN der Wind weht (Herkunft + 180°), damit
// die Pfeilspitze intuitiv in die Strömungsrichtung zeigt.

const COMPASS_DIRS = ['N', 'NO', 'O', 'SO', 'S', 'SW', 'W', 'NW'] as const;

/** Kompass-Kurzform der Windrichtung (DE/EN identisch, Standard-Notation). */
export function compass(deg: number | null | undefined): string {
	if (typeof deg !== 'number') return '—';
	return COMPASS_DIRS[Math.round(deg / 45) % 8];
}

/**
 * Rotation für den Karten-Wind-Pfeil in Grad, 0–360, oder `null` wenn
 * keine Richtung vorliegt. Pfeil zeigt wohin der Wind weht
 * (Herkunftsrichtung + 180°).
 */
export function windArrowRotationDeg(deg: number | null | undefined): number | null {
	if (typeof deg !== 'number') return null;
	return (deg + 180) % 360;
}
