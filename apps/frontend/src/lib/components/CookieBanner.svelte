<script lang="ts">
	import { onMount } from 'svelte';
	import { resolve } from '$app/paths';
	import { cookieConsent } from '$lib/stores/cookieConsent.svelte';

	let functional = $state(false);
	let analytics = $state(false);
	let marketing = $state(false);

	onMount(() => {
		cookieConsent.hydrate();
	});

	function acceptAll() {
		cookieConsent.acceptAll();
	}

	function rejectAll() {
		cookieConsent.rejectAll();
	}

	function saveSelection() {
		cookieConsent.save({ functional, analytics, marketing });
	}
</script>

{#if cookieConsent.showBanner}
	<div
		role="dialog"
		aria-modal="false"
		aria-labelledby="cookie-banner-title"
		aria-describedby="cookie-banner-desc"
		class="fixed inset-x-0 bottom-0 z-50 border-t border-slate-200 bg-white shadow-2xl"
	>
		<div class="mx-auto max-w-4xl px-4 py-5 sm:px-6">
			<h2 id="cookie-banner-title" class="text-lg font-semibold text-slate-900">
				Cookie-Einstellungen
			</h2>
			<p id="cookie-banner-desc" class="mt-2 text-sm text-slate-600">
				Wir nutzen ausschließlich technisch notwendige Cookies für den Betrieb dieser Plattform.
				Optionale Kategorien sind aktuell deaktiviert. Sie können diese Einstellung jederzeit über
				den Footer-Link
				<a
					href={resolve('/cookie-einstellungen')}
					class="underline underline-offset-2 hover:no-underline">Cookie-Einstellungen</a
				>
				ändern. Details in der
				<a href={resolve('/datenschutz')} class="underline underline-offset-2 hover:no-underline"
					>Datenschutzerklärung</a
				>.
			</p>

			<fieldset class="mt-4 space-y-2">
				<legend class="sr-only">Cookie-Kategorien</legend>

				<label class="flex items-start gap-3 text-sm text-slate-700">
					<input
						type="checkbox"
						checked
						disabled
						class="mt-1 h-4 w-4 rounded border-slate-300 text-slate-400"
						aria-describedby="cat-essential-desc"
					/>
					<span>
						<span class="font-medium text-slate-900">Essenziell</span>
						<span id="cat-essential-desc" class="block text-slate-600">
							Notwendig für den Betrieb der Plattform (Session, Sicherheit). Immer aktiv, rechtlich
							nicht abwählbar.
						</span>
					</span>
				</label>

				<label class="flex items-start gap-3 text-sm text-slate-700">
					<input
						type="checkbox"
						bind:checked={functional}
						class="mt-1 h-4 w-4 rounded border-slate-300 text-slate-700 focus:ring-slate-600"
						aria-describedby="cat-functional-desc"
					/>
					<span>
						<span class="font-medium text-slate-900">Funktional</span>
						<span id="cat-functional-desc" class="block text-slate-600">
							Komfort-Features wie Spracheinstellung oder Ortswahl-Cache. Aktuell ungenutzt.
						</span>
					</span>
				</label>

				<label class="flex items-start gap-3 text-sm text-slate-700">
					<input
						type="checkbox"
						bind:checked={analytics}
						class="mt-1 h-4 w-4 rounded border-slate-300 text-slate-700 focus:ring-slate-600"
						aria-describedby="cat-analytics-desc"
					/>
					<span>
						<span class="font-medium text-slate-900">Analytics</span>
						<span id="cat-analytics-desc" class="block text-slate-600">
							Anonyme Nutzungsstatistik zur Verbesserung der Plattform. Aktuell ungenutzt.
						</span>
					</span>
				</label>

				<label class="flex items-start gap-3 text-sm text-slate-700">
					<input
						type="checkbox"
						bind:checked={marketing}
						class="mt-1 h-4 w-4 rounded border-slate-300 text-slate-700 focus:ring-slate-600"
						aria-describedby="cat-marketing-desc"
					/>
					<span>
						<span class="font-medium text-slate-900">Marketing</span>
						<span id="cat-marketing-desc" class="block text-slate-600">
							Cookies für externe Inhalte oder Werbe-Partner. Aktuell ungenutzt.
						</span>
					</span>
				</label>
			</fieldset>

			<div class="mt-5 flex flex-col gap-2 sm:flex-row sm:flex-wrap">
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
					onclick={saveSelection}
					class="rounded-md border border-slate-300 bg-white px-4 py-2 text-sm font-medium text-slate-900 hover:bg-slate-50 focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-slate-600"
				>
					Auswahl speichern
				</button>
			</div>
		</div>
	</div>
{/if}
