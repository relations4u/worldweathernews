<script lang="ts">
	import { onMount } from 'svelte';
	import { cookieConsent } from '$lib/stores/cookieConsent.svelte';

	let functional = $state(false);
	let analytics = $state(false);
	let marketing = $state(false);
	let saved = $state(false);

	onMount(() => {
		cookieConsent.hydrate();
		const c = cookieConsent.consent;
		if (c) {
			functional = c.functional;
			analytics = c.analytics;
			marketing = c.marketing;
		}
	});

	function save() {
		cookieConsent.save({ functional, analytics, marketing });
		saved = true;
	}

	function rejectAll() {
		cookieConsent.rejectAll();
		functional = false;
		analytics = false;
		marketing = false;
		saved = true;
	}

	function acceptAll() {
		cookieConsent.acceptAll();
		functional = true;
		analytics = true;
		marketing = true;
		saved = true;
	}

	function formatTimestamp(ts: number): string {
		return new Intl.DateTimeFormat('de-DE', {
			dateStyle: 'medium',
			timeStyle: 'short'
		}).format(new Date(ts));
	}
</script>

<section class="mx-auto max-w-3xl px-4 py-8">
	<h1 class="text-3xl font-bold tracking-tight text-slate-900">Cookie-Einstellungen</h1>

	<p class="mt-3 text-sm text-slate-600">
		Diese Plattform nutzt ausschließlich technisch notwendige Cookies. Optionale Kategorien können
		Sie hier jederzeit aktivieren oder deaktivieren. Rechtsgrundlage: § 25 TTDSG, Art. 6 Abs. 1 lit.
		a DSGVO.
	</p>

	<div class="mt-6 rounded-md border border-slate-200 bg-slate-50 px-4 py-3 text-sm">
		<strong class="font-semibold text-slate-900">Aktuelle Einstellung:</strong>
		{#if cookieConsent.consent}
			gespeichert am {formatTimestamp(cookieConsent.consent.timestamp)}
		{:else if cookieConsent.hasDecided}
			gespeichert
		{:else}
			noch keine Auswahl getroffen
		{/if}
	</div>

	<fieldset class="mt-6 space-y-3">
		<legend class="sr-only">Cookie-Kategorien</legend>

		<label class="flex items-start gap-3 rounded-md border border-slate-200 px-4 py-3 text-sm">
			<input
				type="checkbox"
				checked
				disabled
				class="mt-1 h-4 w-4 rounded border-slate-300 text-slate-400"
				aria-describedby="cat-essential-desc"
			/>
			<span class="flex-1">
				<span class="font-medium text-slate-900">Essenziell</span>
				<span id="cat-essential-desc" class="block text-slate-600">
					Notwendig für den Betrieb der Plattform (Session, Sicherheit). Immer aktiv, rechtlich
					nicht abwählbar.
				</span>
			</span>
		</label>

		<label class="flex items-start gap-3 rounded-md border border-slate-200 px-4 py-3 text-sm">
			<input
				type="checkbox"
				bind:checked={functional}
				class="mt-1 h-4 w-4 rounded border-slate-300 text-slate-700 focus:ring-slate-600"
				aria-describedby="cat-functional-desc"
			/>
			<span class="flex-1">
				<span class="font-medium text-slate-900">Funktional</span>
				<span id="cat-functional-desc" class="block text-slate-600">
					Komfort-Features wie Spracheinstellung oder Ortswahl-Cache. Aktuell ungenutzt.
				</span>
			</span>
		</label>

		<label class="flex items-start gap-3 rounded-md border border-slate-200 px-4 py-3 text-sm">
			<input
				type="checkbox"
				bind:checked={analytics}
				class="mt-1 h-4 w-4 rounded border-slate-300 text-slate-700 focus:ring-slate-600"
				aria-describedby="cat-analytics-desc"
			/>
			<span class="flex-1">
				<span class="font-medium text-slate-900">Analytics</span>
				<span id="cat-analytics-desc" class="block text-slate-600">
					Anonyme Nutzungsstatistik zur Verbesserung der Plattform. Aktuell ungenutzt.
				</span>
			</span>
		</label>

		<label class="flex items-start gap-3 rounded-md border border-slate-200 px-4 py-3 text-sm">
			<input
				type="checkbox"
				bind:checked={marketing}
				class="mt-1 h-4 w-4 rounded border-slate-300 text-slate-700 focus:ring-slate-600"
				aria-describedby="cat-marketing-desc"
			/>
			<span class="flex-1">
				<span class="font-medium text-slate-900">Marketing</span>
				<span id="cat-marketing-desc" class="block text-slate-600">
					Cookies für externe Inhalte oder Werbe-Partner. Aktuell ungenutzt.
				</span>
			</span>
		</label>
	</fieldset>

	<div class="mt-6 flex flex-col gap-2 sm:flex-row sm:flex-wrap">
		<button
			type="button"
			onclick={acceptAll}
			class="rounded-md bg-slate-900 px-4 py-2 text-sm font-medium text-white hover:bg-slate-800 focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-slate-600"
		>
			Alle akzeptieren
		</button>
		<button
			type="button"
			onclick={rejectAll}
			class="rounded-md bg-slate-900 px-4 py-2 text-sm font-medium text-white hover:bg-slate-800 focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-slate-600"
		>
			Alle ablehnen
		</button>
		<button
			type="button"
			onclick={save}
			class="rounded-md border border-slate-300 bg-white px-4 py-2 text-sm font-medium text-slate-900 hover:bg-slate-50 focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-slate-600"
		>
			Auswahl speichern
		</button>
	</div>

	{#if saved}
		<p
			role="status"
			class="mt-4 rounded-md border border-emerald-200 bg-emerald-50 px-4 py-3 text-sm text-emerald-900"
		>
			Einstellung gespeichert.
		</p>
	{/if}
</section>
