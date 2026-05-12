<script lang="ts">
	import * as m from '$lib/paraglide/messages';
	import { getLocale } from '$lib/paraglide/runtime';
	import type { LocationDetail } from '$lib/api/client';

	let { detail }: { detail: LocationDetail } = $props();

	const current = $derived(detail.current);

	// Niederschlags-Summe der nächsten 24 h aus dem Forecast.
	const precipitationNext24h = $derived(
		detail.forecast.reduce<number>(
			(sum, f) => sum + (typeof f.precipitation === 'number' ? f.precipitation : 0),
			0
		)
	);

	// Kompass-Kurzform der Windrichtung (DE/EN identisch, ist Standard-Notation).
	function compass(deg: number | null | undefined): string {
		if (typeof deg !== 'number') return '—';
		const dirs = ['N', 'NO', 'O', 'SO', 'S', 'SW', 'W', 'NW'];
		return dirs[Math.round(deg / 45) % 8];
	}

	function formatTime(iso: string): string {
		const locale = getLocale() === 'en' ? 'en-GB' : 'de-DE';
		return new Date(iso).toLocaleTimeString(locale, {
			hour: '2-digit',
			minute: '2-digit'
		});
	}
</script>

<article
	class="rounded-lg border border-slate-200 bg-white p-6 shadow-sm"
	aria-labelledby={`weather-card-${detail.location.slug}`}
>
	<h2 id={`weather-card-${detail.location.slug}`} class="text-xl font-semibold text-slate-900">
		{detail.location.name}
	</h2>

	{#if current}
		<div class="mt-4 flex items-baseline gap-2">
			<span class="text-4xl font-bold text-slate-900">
				{typeof current.temperature === 'number' ? current.temperature.toFixed(1) : '—'}
			</span>
			<span class="text-slate-600">°C</span>
		</div>

		<dl class="mt-4 grid grid-cols-2 gap-x-4 gap-y-2 text-sm">
			<dt class="text-slate-600">{m.weather_precipitation_24h()}</dt>
			<dd class="text-slate-900">{precipitationNext24h.toFixed(1)} mm</dd>

			<dt class="text-slate-600">{m.weather_wind_speed()}</dt>
			<dd class="text-slate-900">
				{typeof current.windSpeed === 'number' ? `${current.windSpeed.toFixed(1)} km/h` : '—'}
			</dd>

			<dt class="text-slate-600">{m.weather_wind_direction()}</dt>
			<dd class="text-slate-900">
				{#if typeof current.windDirection === 'number'}
					{compass(current.windDirection)} ({current.windDirection}°)
				{:else}
					—
				{/if}
			</dd>
		</dl>

		<p class="mt-4 text-xs text-slate-500">
			{m.weather_observed_at()}: {formatTime(current.observedAt)}
		</p>
	{:else}
		<p class="mt-4 text-slate-600">{m.weather_no_data()}</p>
	{/if}
</article>
