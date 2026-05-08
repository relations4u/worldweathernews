<script lang="ts">
	import { onMount } from 'svelte';
	import { resolve } from '$app/paths';
	import { ping } from '$lib/api/client';
	import { Badge } from '$lib/components/ui/badge';

	let status = $state<'pending' | 'ok' | 'error'>('pending');
	let traceId = $state<string | null>(null);
	let errorMsg = $state<string | null>(null);

	onMount(async () => {
		try {
			const result = await ping();
			status = 'ok';
			traceId = result.traceId;
		} catch (err) {
			status = 'error';
			errorMsg = err instanceof Error ? err.message : 'unknown error';
		}
	});
</script>

<section class="mx-auto max-w-3xl px-4 py-12 text-center">
	<h1 class="text-4xl font-bold tracking-tight text-slate-900 sm:text-5xl">
		Wetter und Klima.<br />Quellen, Kontext, Community.
	</h1>
	<p class="mt-6 text-lg text-slate-600">
		worldweathernews.com aggregiert Wetterdaten und Klima-Indikatoren aus nationalen Wetterdiensten
		weltweit, ordnet sie meteorologisch ein und macht sie für Beobachter, Citizen Scientists und
		alle Wetter-Interessierten zugänglich.
	</p>
	<p class="mt-4 text-sm text-slate-600">
		Die Plattform befindet sich in der Forschungs-Phase. Inhalte und Daten werden schrittweise
		ergänzt.
	</p>

	<div class="mt-8 flex flex-wrap justify-center gap-3">
		<a
			href={resolve('/about')}
			class="rounded-md bg-slate-900 px-5 py-2.5 text-sm font-medium text-white hover:bg-slate-800"
		>
			Über die Plattform
		</a>
		<!-- /methodik-Route folgt in Iteration 1.2 — bis dahin plain href, dann auf resolve() umstellen. -->
		<!-- eslint-disable svelte/no-navigation-without-resolve -->
		<a
			href="/methodik"
			class="rounded-md border border-slate-300 bg-white px-5 py-2.5 text-sm font-medium text-slate-900 hover:bg-slate-50"
		>
			Methodik (folgt)
		</a>
		<!-- eslint-enable svelte/no-navigation-without-resolve -->
	</div>
</section>

<section class="mx-auto max-w-3xl px-4 pb-12 text-center text-xs text-slate-500">
	{#if status === 'pending'}
		<Badge variant="outline">Backend-Verbindung wird geprüft …</Badge>
	{:else if status === 'ok'}
		<Badge>Backend erreichbar</Badge>
		{#if traceId}
			<span class="ml-2">Trace: {traceId}</span>
		{/if}
	{:else}
		<Badge variant="destructive">Backend nicht erreichbar</Badge>
		{#if errorMsg}
			<span class="ml-2">{errorMsg}</span>
		{/if}
	{/if}
</section>
