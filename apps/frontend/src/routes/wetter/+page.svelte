<script lang="ts">
	import { resolve } from '$app/paths';
	import * as m from '$lib/paraglide/messages';
	import WeatherCard from '$lib/components/WeatherCard.svelte';

	let { data } = $props();
</script>

<svelte:head>
	<title>{m.weather_page_title()} — {m.site_title()}</title>
	<meta name="description" content={m.weather_page_lead()} />
</svelte:head>

<section class="mx-auto max-w-5xl px-4 py-8">
	<h1 class="text-3xl font-bold tracking-tight text-slate-900">{m.weather_page_title()}</h1>
	<p class="mt-3 text-slate-600">{m.weather_page_lead()}</p>

	{#if data.details.length === 0}
		<p class="mt-8 text-slate-600">{m.weather_no_locations()}</p>
	{:else}
		<div class="mt-8 grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
			{#each data.details as detail (detail.location.slug)}
				<WeatherCard {detail} />
			{/each}
		</div>
	{/if}

	<p class="mt-8 text-xs text-slate-500">
		{data.attribution} —
		<a class="underline hover:text-slate-700" href={resolve('/quellen-attribution')}>
			{m.weather_attribution_link()}
		</a>
	</p>
</section>
