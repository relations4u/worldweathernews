import type { RequestHandler } from './$types';
import { locales, localizeHref } from '$lib/paraglide/runtime';

// Statische Routes mit fester (DE-only) Verfügbarkeit.
// Methodik wird separat behandelt, weil bilingual.
const deOnlyRoutes = [
	'/',
	'/about',
	'/impressum',
	'/datenschutz',
	'/barrierefreiheit',
	'/kontakt',
	'/quellen-attribution',
	'/cookie-einstellungen'
];

// Markdown-basierte Pages mit potenziell mehreren Locales.
// import.meta.glob bündelt zur Build-Zeit, daher kostenlos zur Laufzeit.
type MdsvexMetadata = {
	slug: string;
	status?: string;
};
type MdsvexModule = {
	metadata: MdsvexMetadata;
};
const pages = import.meta.glob<MdsvexModule>('/src/content/pages/*/*.md', {
	eager: true
});

function listMarkdownEntries(): Array<{ path: string; localesAvailable: string[] }> {
	// Map slug → set of locales where it's published.
	const map = new Map<string, Set<string>>();
	for (const [filePath, mod] of Object.entries(pages)) {
		if (mod.metadata?.status !== 'published') continue;
		// /src/content/pages/<folder>/<slug>.md
		const match = filePath.match(/\/src\/content\/pages\/([^/]+)\/([^/]+)\.md$/);
		if (!match) continue;
		const folder = match[1];
		const slug = mod.metadata.slug ?? match[2];
		const tag = folder === 'de' ? 'de-de' : folder;
		if (!map.has(slug)) map.set(slug, new Set());
		map.get(slug)!.add(tag);
	}
	return Array.from(map.entries()).map(([slug, set]) => ({
		path: `/${slug}`,
		localesAvailable: Array.from(set)
	}));
}

function xmlEscape(s: string): string {
	return s
		.replace(/&/g, '&amp;')
		.replace(/</g, '&lt;')
		.replace(/>/g, '&gt;')
		.replace(/"/g, '&quot;');
}

function urlEntry(loc: string, alternates: Array<{ href: string; hreflang: string }>): string {
	const links = alternates
		.map((a) => `\n    <xhtml:link rel="alternate" hreflang="${a.hreflang}" href="${a.href}" />`)
		.join('');
	return `  <url>\n    <loc>${xmlEscape(loc)}</loc>${links}\n  </url>`;
}

export const GET: RequestHandler = ({ url }) => {
	const origin = url.origin;
	const lines: string[] = [];

	// DE-only-Routes
	for (const path of deOnlyRoutes) {
		const loc = origin + path;
		lines.push(urlEntry(loc, []));
	}

	// Markdown-Routes (potenziell bilingual)
	for (const entry of listMarkdownEntries()) {
		const alternates: Array<{ href: string; hreflang: string }> = [];
		for (const tag of entry.localesAvailable) {
			alternates.push({
				href: origin + localizeHref(entry.path, { locale: tag as (typeof locales)[number] }),
				hreflang: tag
			});
		}
		// x-default = DE-DE-Variante (sofern vorhanden)
		if (entry.localesAvailable.includes('de-de')) {
			alternates.push({
				href: origin + localizeHref(entry.path, { locale: 'de-de' }),
				hreflang: 'x-default'
			});
		}
		// Eine URL-Entry pro Locale; jede listet alle Alternates
		for (const tag of entry.localesAvailable) {
			lines.push(
				urlEntry(
					origin + localizeHref(entry.path, { locale: tag as (typeof locales)[number] }),
					alternates
				)
			);
		}
	}

	const xml =
		`<?xml version="1.0" encoding="UTF-8"?>\n` +
		`<urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9"\n` +
		`        xmlns:xhtml="http://www.w3.org/1999/xhtml">\n` +
		lines.join('\n') +
		`\n</urlset>\n`;

	return new Response(xml, {
		headers: {
			'Content-Type': 'application/xml; charset=utf-8',
			'Cache-Control': 'public, max-age=3600'
		}
	});
};
