// Satelliten-Frame-Index (Iteration 2.4, Pfad A).
//
// Bewusst genau EINE Stelle. Der pyworkers-EUMETSAT-Worker schreibt
// `index.json` + Frames server-seitig in den A.13-Bucket; das Frontend
// lädt ausschließlich über das eigene `media.worldweathernews.com`
// (kein Drittanbieter-Client-Pfad, A.19-konform). Die Frame-URLs in
// der index.json sind bereits absolut (vom Worker gebaut) — hier nur
// der Einstiegspunkt. Wechsel des Layers/Hosts = Ein-Zeilen-Change.
export const SAT_INDEX_URL = 'https://media.worldweathernews.com/sat/ir108/index.json';

export type SatFrame = {
	time: string;
	url: string;
};

export type SatIndex = {
	layer: string;
	source: string;
	attribution: string;
	bbox: {
		lonMin: number;
		latMin: number;
		lonMax: number;
		latMax: number;
	};
	frames: SatFrame[];
};
