<script lang="ts">
	import '../app.css';
	import { resolve } from '$app/paths';
	import favicon from '$lib/assets/favicon.svg';
	import ResearchBanner from '$lib/components/ResearchBanner.svelte';
	import CookieBanner from '$lib/components/CookieBanner.svelte';

	let { children } = $props();
</script>

<svelte:head>
	<link rel="icon" href={favicon} />
	<title>worldweathernews — Wetter, Klima, Community</title>
	<meta
		name="description"
		content="worldweathernews.com aggregiert Wetter- und Klimadaten aus nationalen Wetterdiensten weltweit für eine Community von Beobachtern und Wetter-Interessierten."
	/>
</svelte:head>

<div class="flex min-h-screen flex-col bg-white text-slate-900">
	<a
		href="#main-content"
		class="sr-only focus:not-sr-only focus:absolute focus:left-2 focus:top-2 focus:z-50 focus:rounded-md focus:bg-slate-900 focus:px-3 focus:py-2 focus:text-sm focus:font-medium focus:text-white"
	>
		Zum Hauptinhalt springen
	</a>

	<ResearchBanner />

	<header class="border-b border-slate-200">
		<div class="mx-auto flex max-w-6xl items-center justify-between gap-4 px-4 py-4">
			<a href={resolve('/')} class="text-xl font-semibold tracking-tight">worldweathernews</a>

			<nav aria-label="Hauptnavigation" class="flex items-center gap-4 text-sm text-slate-600">
				<a href={resolve('/about')} class="hover:text-slate-900">Über</a>
				<a href={resolve('/kontakt')} class="hover:text-slate-900">Kontakt</a>
				<!-- Locale-Switcher-Slot — wird in Iteration 1.2 mit Paraglide gefüllt. -->
				<div data-slot="locale-switcher" aria-hidden="true"></div>
			</nav>
		</div>
	</header>

	<main id="main-content" class="mx-auto w-full max-w-6xl flex-1 px-4 py-8">
		{@render children()}
	</main>

	<footer class="border-t border-slate-200 bg-slate-50 text-sm text-slate-600">
		<div class="mx-auto max-w-6xl px-4 py-6">
			<nav aria-label="Pflicht-Links" class="flex flex-wrap gap-x-4 gap-y-2">
				<a href={resolve('/impressum')} class="hover:text-slate-900">Impressum</a>
				<a href={resolve('/datenschutz')} class="hover:text-slate-900">Datenschutz</a>
				<a href={resolve('/barrierefreiheit')} class="hover:text-slate-900">Barrierefreiheit</a>
				<a href={resolve('/cookie-einstellungen')} class="hover:text-slate-900"
					>Cookie-Einstellungen</a
				>
				<a href={resolve('/quellen-attribution')} class="hover:text-slate-900"
					>Quellen-Attribution</a
				>
				<a href={resolve('/kontakt')} class="hover:text-slate-900">Kontakt</a>
			</nav>
			<p class="mt-4 text-xs text-slate-600">
				© {new Date().getFullYear()} worldweathernews — Forschungs-Phase
			</p>
		</div>
	</footer>

	<CookieBanner />
</div>
