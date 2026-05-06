// i18n-Library-Wahl (svelte-i18n vs. Paraglide vs. Inlang) ist in
// docs/backlog.md vorgemerkt; entschieden wird in der ersten
// Feature-Session, die User-facing Strings einführt. Bis dahin nur
// die Übersetzungs-Bundles als statische Imports — Komponenten
// verwenden vorerst englische Strings direkt.

import de from './de.json';
import en from './en.json';

export type Locale = 'de' | 'en';
export const messages = { de, en } as const;
