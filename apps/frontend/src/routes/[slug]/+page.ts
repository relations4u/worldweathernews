import { error } from '@sveltejs/kit';
import type { PageLoad } from './$types';
import { getLocale } from '$lib/paraglide/runtime';

// Markdown-Inhalte für lokalisierte Pages.
// Per locale-Tag → Verzeichnis: 'de-de' → 'de', 'en' → 'en'.
// import.meta.glob mit eager: true bündelt alle Pages zur Build-Zeit;
// Vite tree-shaked unbenutzte Locale-Bundles raus, sobald wir Code-Splitting
// pro Locale brauchen.
const modules = import.meta.glob('/src/content/pages/*/*.md', { eager: true });

const localeToFolder: Record<string, string> = {
	'de-de': 'de',
	en: 'en'
};

export type PageMetadata = {
	title: string;
	slug: string;
	lang: string;
	lead?: string;
	updated_at?: string;
	status?: 'draft' | 'published';
};

type MdsvexModule = {
	default: import('svelte').Component;
	metadata: PageMetadata;
};

export const load: PageLoad = ({ params }) => {
	const locale = getLocale();
	const folder = localeToFolder[locale] ?? 'de';
	const path = `/src/content/pages/${folder}/${params.slug}.md`;
	const mod = (modules as Record<string, MdsvexModule>)[path];

	if (!mod || mod.metadata?.status !== 'published') {
		error(404, `Seite nicht gefunden: ${params.slug} (${locale})`);
	}

	return {
		metadata: mod.metadata,
		Content: mod.default
	};
};
