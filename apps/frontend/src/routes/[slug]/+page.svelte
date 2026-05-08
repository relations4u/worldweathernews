<script lang="ts">
	import { page } from '$app/state';
	import * as m from '$lib/paraglide/messages';
	import { localizeHref } from '$lib/paraglide/runtime';

	let { data } = $props();
	const Content = $derived(data.Content);
	const metadata = $derived(data.metadata);
	const availableLocales = $derived(data.availableLocales);

	// Kanonischer (de-localisierter) Pfad — Basis für hreflang-Generierung.
	const canonicalPath = $derived(`/${page.params.slug}`);
</script>

<svelte:head>
	<title>{metadata.title} — worldweathernews</title>
	{#if metadata.lead}
		<meta name="description" content={metadata.lead} />
		<meta property="og:description" content={metadata.lead} />
	{/if}
	<meta property="og:title" content={metadata.title} />
	<meta property="og:type" content="article" />
	{#each availableLocales as tag (tag)}
		<link
			rel="alternate"
			hreflang={tag}
			href={page.url.origin + localizeHref(canonicalPath, { locale: tag as 'de-de' | 'en' })}
		/>
	{/each}
	{#if availableLocales.includes('de-de')}
		<link
			rel="alternate"
			hreflang="x-default"
			href={page.url.origin + localizeHref(canonicalPath, { locale: 'de-de' })}
		/>
	{/if}
</svelte:head>

<article class="mx-auto max-w-3xl px-4 py-8 text-slate-900">
	<header>
		<h1 class="text-3xl font-bold tracking-tight">{metadata.title}</h1>
		{#if metadata.lead}
			<p class="mt-3 text-lg text-slate-600">{metadata.lead}</p>
		{/if}
	</header>

	<div
		class="mt-8 space-y-4 text-slate-700 [&_a]:text-slate-900 [&_a]:underline [&_a]:underline-offset-2 hover:[&_a]:no-underline [&_h2]:mt-8 [&_h2]:text-xl [&_h2]:font-semibold [&_h2]:text-slate-900 [&_h3]:mt-6 [&_h3]:text-lg [&_h3]:font-semibold [&_h3]:text-slate-900 [&_ol]:ml-6 [&_ol]:list-decimal [&_p]:leading-relaxed [&_ul]:ml-6 [&_ul]:list-disc [&_ul]:space-y-1"
	>
		<Content />
	</div>

	{#if metadata.updated_at}
		<footer class="mt-12 border-t border-slate-200 pt-4 text-xs text-slate-500">
			{m.page_last_updated()}: {metadata.updated_at}
		</footer>
	{/if}
</article>
