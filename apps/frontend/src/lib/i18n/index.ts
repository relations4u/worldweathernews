// TODO: Library-Wahl (svelte-i18n vs. Paraglide vs. Inlang) wird in einer
// späteren Session entschieden. Bis dahin nur die Übersetzungs-Bundles
// als statische Imports — Komponenten verwenden vorerst englische Strings
// direkt.

import de from './de.json';
import en from './en.json';

export type Locale = 'de' | 'en';
export const messages = { de, en } as const;
