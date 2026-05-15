// Tile-Quelle für die Stations-Map.
//
// Bewusst genau EINE Stelle: ein späterer Wechsel auf einen self-hosted
// OpenFreeMap-Stack oder MapTiler ist hier ein Ein-Zeilen-Change. Die
// Style-URL darf nirgends sonst hartkodiert werden.
//
// Entscheidung Iteration 2.3: T2 OpenFreeMap (Liberty-Style) — frei,
// kein Account/API-Key, keine Cookies, Vector-Tiles. Self-hosting-
// Spannung (A.19) bewusst akzeptiert: client-seitig, nicht
// backend-kritisch. Begründung: sessions/feature2/prompt-iteration-2-3.md.
export const MAP_STYLE_URL = 'https://tiles.openfreemap.org/styles/liberty';

// Anfangs-Viewport — grob auf die sechs DE-Stationen zentriert
// (Helgoland 7.9°O bis Brocken/Zugspitze, Helgoland 54.2°N bis
// Zugspitze 47.4°N). Der Zoom zeigt alle Marker ohne initiales Pan.
export const MAP_INITIAL_VIEW = {
	center: [10.0, 51.0] as [number, number],
	zoom: 5
};
