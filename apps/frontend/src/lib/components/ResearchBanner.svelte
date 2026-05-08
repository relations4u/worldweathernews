<script lang="ts">
	import { X } from 'lucide-svelte';

	const STORAGE_KEY = 'wwn-research-banner-closed';

	let closed = $state(false);

	$effect(() => {
		if (localStorage.getItem(STORAGE_KEY) === 'true') {
			closed = true;
		}
	});

	function close() {
		closed = true;
		localStorage.setItem(STORAGE_KEY, 'true');
	}
</script>

{#if !closed}
	<aside
		aria-label="Hinweis zur Forschungs-Phase"
		class="sticky top-0 z-40 border-b border-amber-200 bg-amber-50 text-amber-900"
	>
		<div class="mx-auto flex max-w-6xl items-center gap-3 px-4 py-2 text-sm">
			<p class="flex-1">
				<strong class="font-semibold">Forschungs-Phase.</strong>
				worldweathernews.com befindet sich noch im Aufbau. Inhalte und Daten können sich kurzfristig ändern.
				<!-- /methodik-Route folgt in Iteration 1.2 — bis dahin plain href, dann auf resolve() umstellen. -->
				<!-- eslint-disable svelte/no-navigation-without-resolve -->
				<a
					href="/methodik"
					class="underline underline-offset-2 hover:no-underline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-amber-600"
				>
					Mehr zur Methodik →
				</a>
				<!-- eslint-enable svelte/no-navigation-without-resolve -->
			</p>
			<button
				type="button"
				class="rounded p-1 hover:bg-amber-100 focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-amber-600"
				aria-label="Hinweis schließen"
				onclick={close}
			>
				<X class="h-4 w-4" aria-hidden="true" />
			</button>
		</div>
	</aside>
{/if}
