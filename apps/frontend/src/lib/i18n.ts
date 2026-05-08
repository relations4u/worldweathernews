// Paraglide-v2-Konfiguration für die Plattform.
//
// Default-Locale: de-de (kein Pfad-Prefix, an der Apex unter /).
// Zusätzliche Locale: en (mit /en/-Prefix).
//
// Die URL-Patterns sind direkt in `project.inlang/settings.json` über das
// urlStrategy-Plugin konfiguriert; dieses Modul re-exportiert nur die
// laufzeitseitig wichtigen Helfer für die Komponenten und Loader.
import {
	baseLocale,
	locales,
	getLocale,
	setLocale,
	localizeHref,
	deLocalizeHref,
	localizeUrl,
	deLocalizeUrl
} from '$lib/paraglide/runtime';

export type Locale = (typeof locales)[number];

export {
	baseLocale,
	locales,
	getLocale,
	setLocale,
	localizeHref,
	deLocalizeHref,
	localizeUrl,
	deLocalizeUrl
};
