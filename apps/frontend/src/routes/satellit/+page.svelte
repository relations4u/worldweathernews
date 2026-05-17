<script lang="ts">
	import { resolve } from '$app/paths';
	import * as m from '$lib/paraglide/messages';
	import SatelliteMap from '$lib/components/SatelliteMap.svelte';

	let { data } = $props();
</script>

<svelte:head>
	<title>{m.sat_page_title()} — {m.site_title()}</title>
	<meta name="description" content={m.sat_page_lead()} />
</svelte:head>

<section class="mx-auto max-w-5xl px-4 py-8">
	<h1 class="text-3xl font-bold tracking-tight text-slate-900">{m.sat_page_title()}</h1>
	<p class="mt-3 text-slate-600">{m.sat_page_lead()}</p>

	{#if data.index.frames.length === 0}
		<p class="mt-8 text-slate-600">{m.sat_no_data()}</p>
	{:else}
		<div class="mt-8">
			<SatelliteMap index={data.index} />
		</div>
	{/if}

	<p class="mt-8 text-xs text-slate-500">
		{data.index.attribution} —
		<a class="underline hover:text-slate-700" href={resolve('/quellen-attribution')}>
			{m.weather_attribution_link()}
		</a>
	</p>
</section>
