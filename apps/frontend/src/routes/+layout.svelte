<script lang="ts">
	import '../app.css';
	import { resolve } from '$app/paths';
	import { page } from '$app/state';
	import * as m from '$lib/paraglide/messages';
	import { getLocale } from '$lib/paraglide/runtime';
	import logoUrl from '$lib/assets/logo.png';
	import ResearchBanner from '$lib/components/ResearchBanner.svelte';
	import CookieBanner from '$lib/components/CookieBanner.svelte';
	import LocaleSwitcher from '$lib/components/LocaleSwitcher.svelte';

	let { children } = $props();

	const ogLocale = $derived(getLocale() === 'en' ? 'en_US' : 'de_DE');
</script>

<svelte:head>
	<link rel="icon" type="image/png" sizes="32x32" href="/favicon-32x32.png" />
	<link rel="icon" type="image/png" sizes="16x16" href="/favicon-16x16.png" />
	<link rel="shortcut icon" href="/favicon.ico" />
	<link rel="apple-touch-icon" sizes="180x180" href="/apple-touch-icon.png" />
	<title>{m.site_title()} — {m.site_tagline()}</title>
	<meta name="description" content={m.site_tagline()} />
	<meta property="og:title" content="{m.site_title()} — {m.site_tagline()}" />
	<meta property="og:description" content={m.site_tagline()} />
	<meta property="og:locale" content={ogLocale} />
	<meta property="og:type" content="website" />
	<meta property="og:url" content={page.url.href} />
	<meta name="twitter:card" content="summary" />
</svelte:head>

<div class="flex min-h-screen flex-col bg-white text-slate-900">
	<a
		href="#main-content"
		class="sr-only focus:not-sr-only focus:absolute focus:top-2 focus:left-2 focus:z-50 focus:rounded-md focus:bg-slate-900 focus:px-3 focus:py-2 focus:text-sm focus:font-medium focus:text-white"
	>
		{m.nav_skip_to_main()}
	</a>

	<ResearchBanner />

	<header class="border-b border-slate-200">
		<div class="mx-auto flex max-w-6xl items-center justify-between gap-4 px-4 py-4">
			<a href={resolve('/')} class="flex items-center gap-3 text-xl font-semibold tracking-tight">
				<img src={logoUrl} alt="" class="h-10 w-10" width="40" height="40" />
				<span>{m.site_title()}</span>
			</a>

			<nav aria-label="Hauptnavigation" class="flex items-center gap-4 text-sm text-slate-600">
				<a href={resolve('/wetter')} class="hover:text-slate-900">{m.nav_weather()}</a>
				<a href={resolve('/satellit')} class="hover:text-slate-900">{m.nav_satellite()}</a>
				<a href={resolve('/about')} class="hover:text-slate-900">{m.nav_about()}</a>
				<a href={resolve('/kontakt')} class="hover:text-slate-900">{m.nav_contact()}</a>
				<LocaleSwitcher />
			</nav>
		</div>
	</header>

	<main id="main-content" class="mx-auto w-full max-w-6xl flex-1 px-4 py-8">
		{@render children()}
	</main>

	<footer class="border-t border-slate-200 bg-slate-50 text-sm text-slate-600">
		<div class="mx-auto max-w-6xl px-4 py-6">
			<nav aria-label="Pflicht-Links" class="flex flex-wrap gap-x-4 gap-y-2">
				<a href={resolve('/impressum')} class="hover:text-slate-900">{m.footer_imprint()}</a>
				<a href={resolve('/datenschutz')} class="hover:text-slate-900">{m.footer_privacy()}</a>
				<a href={resolve('/barrierefreiheit')} class="hover:text-slate-900"
					>{m.footer_accessibility()}</a
				>
				<a href={resolve('/cookie-einstellungen')} class="hover:text-slate-900"
					>{m.footer_cookies()}</a
				>
				<a href={resolve('/quellen-attribution')} class="hover:text-slate-900"
					>{m.footer_attribution()}</a
				>
				<a href={resolve('/kontakt')} class="hover:text-slate-900">{m.footer_contact()}</a>
			</nav>
			<p class="mt-4 text-xs text-slate-600">
				© {new Date().getFullYear()} worldweathernews — {m.footer_research_phase()}
			</p>
		</div>
	</footer>

	<CookieBanner />
</div>
