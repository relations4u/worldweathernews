<script lang="ts">
	import { onMount, onDestroy } from 'svelte';
	import { browser } from '$app/environment';
	import * as m from '$lib/paraglide/messages';
	import { compass, windArrowRotationDeg } from '$lib/wind';
	import { MAP_STYLE_URL, MAP_INITIAL_VIEW } from '$lib/config/map';
	import type { LocationDetail } from '$lib/api/client';
	// Reine Typ-Imports — werden vom Compiler gelöscht, landen NICHT im
	// Bundle. Die Laufzeit-Lib kommt ausschließlich über den dynamischen
	// import() in onMount (Q6: maplibre-gl bleibt lazy).
	import type { Map as MapLibreMap, Marker as MapLibreMarker } from 'maplibre-gl';

	let { details }: { details: LocationDetail[] } = $props();

	let container: HTMLDivElement;
	let map: MapLibreMap | undefined;
	const markers: MapLibreMarker[] = [];
	let loading = $state(true);

	function tempLabel(d: LocationDetail): string {
		const t = d.current?.temperature;
		return typeof t === 'number' ? `${t.toFixed(0)}°` : '—';
	}

	function sourceLabel(d: LocationDetail): string {
		return d.current?.source === 'dwd' ? m.weather_source_dwd() : m.weather_source_open_meteo();
	}

	function buildMarkerElement(d: LocationDetail): HTMLButtonElement {
		const btn = document.createElement('button');
		btn.type = 'button';
		btn.className = 'wwn-marker';
		btn.setAttribute('aria-label', `${d.location.name}: ${tempLabel(d)}C`);

		const temp = document.createElement('span');
		temp.className = 'wwn-marker__temp';
		temp.textContent = tempLabel(d);
		btn.appendChild(temp);

		const rot = windArrowRotationDeg(d.current?.windDirection);
		if (rot !== null) {
			const arrow = document.createElement('span');
			arrow.className = 'wwn-marker__arrow';
			arrow.style.transform = `rotate(${rot}deg)`;
			arrow.setAttribute('aria-hidden', 'true');
			arrow.textContent = '↑';
			btn.appendChild(arrow);
		}
		return btn;
	}

	function appendRow(parent: HTMLElement, label: string, value: string): void {
		const dt = document.createElement('dt');
		dt.textContent = label;
		const dd = document.createElement('dd');
		dd.textContent = value;
		parent.append(dt, dd);
	}

	// Popup-Inhalt wird imperativ via DOM-API gebaut (textContent escaped
	// automatisch — kein innerHTML/XSS). Phase-1-Set (Q3): Name, Quelle,
	// Temp, Niederschlag jetzt, Wind, Detail-Link.
	function buildPopupContent(d: LocationDetail): HTMLElement {
		const root = document.createElement('div');
		root.className = 'wwn-popup';

		const head = document.createElement('div');
		head.className = 'wwn-popup__head';
		const h = document.createElement('h3');
		h.textContent = d.location.name;
		head.appendChild(h);
		if (d.current) {
			const badge = document.createElement('span');
			badge.className = 'wwn-popup__badge';
			badge.textContent = sourceLabel(d);
			head.appendChild(badge);
		}
		root.appendChild(head);

		const c = d.current;
		if (!c) {
			const p = document.createElement('p');
			p.className = 'wwn-popup__nodata';
			p.textContent = m.weather_no_data();
			root.appendChild(p);
		} else {
			const t = document.createElement('p');
			t.className = 'wwn-popup__temp';
			t.textContent = typeof c.temperature === 'number' ? `${c.temperature.toFixed(1)} °C` : '—';
			root.appendChild(t);

			const dl = document.createElement('dl');
			dl.className = 'wwn-popup__dl';
			appendRow(
				dl,
				m.weather_precipitation_now(),
				typeof c.precipitation === 'number' ? `${c.precipitation.toFixed(1)} mm` : '—'
			);
			appendRow(
				dl,
				m.weather_wind_speed(),
				typeof c.windSpeed === 'number' ? `${c.windSpeed.toFixed(1)} km/h` : '—'
			);
			appendRow(
				dl,
				m.weather_wind_direction(),
				typeof c.windDirection === 'number'
					? `${compass(c.windDirection)} (${c.windDirection}°)`
					: '—'
			);
			root.appendChild(dl);
		}

		// Kein eigener Detail-Route in der App — der Link springt zur
		// passenden WeatherCard unter der Karte (Anker `weather-card-<slug>`,
		// von WeatherCard.svelte gesetzt). Passt zu N2 (Karte Hero, Cards drunter).
		const link = document.createElement('a');
		link.className = 'wwn-popup__link';
		link.href = `#weather-card-${d.location.slug}`;
		link.textContent = m.weather_detail_link();
		root.appendChild(link);

		return root;
	}

	onMount(async () => {
		if (!browser) return;
		const { Map, Marker, Popup, NavigationControl } = await import('maplibre-gl');
		// CSS lazy mit dem maplibre-Chunk — bleibt aus dem Default-Bundle (Q6).
		await import('maplibre-gl/dist/maplibre-gl.css');

		const instance = new Map({
			container,
			style: MAP_STYLE_URL,
			center: MAP_INITIAL_VIEW.center,
			zoom: MAP_INITIAL_VIEW.zoom,
			attributionControl: { compact: true }
		});
		map = instance;
		instance.addControl(new NavigationControl({ showCompass: false }), 'top-right');
		instance.on('load', () => {
			loading = false;
		});

		for (const detail of details) {
			const { latitude, longitude } = detail.location;
			if (typeof latitude !== 'number' || typeof longitude !== 'number') continue;
			const popup = new Popup({ offset: 24, closeButton: true }).setDOMContent(
				buildPopupContent(detail)
			);
			const marker = new Marker({ element: buildMarkerElement(detail) })
				.setLngLat([longitude, latitude])
				.setPopup(popup)
				.addTo(instance);
			markers.push(marker);
		}
	});

	onDestroy(() => {
		for (const mk of markers) mk.remove();
		map?.remove();
	});
</script>

<div class="relative">
	<!-- Server rendert nur diesen leeren Wrapper (S1). Feste Höhe gegen
	     Layout-Shift; Map-Init erst client-side in onMount. -->
	<div
		bind:this={container}
		class="h-[320px] w-full overflow-hidden rounded-lg border border-slate-200 sm:h-[440px]"
		role="region"
		aria-label={m.weather_map_label()}
	></div>
	{#if loading}
		<div
			class="pointer-events-none absolute inset-0 flex items-center justify-center text-sm text-slate-500"
		>
			{m.weather_map_loading()}
		</div>
	{/if}
</div>

<style>
	/* Marker/Popup werden imperativ erzeugt (nicht im Svelte-Template),
	   daher :global. Klassen-Präfix `wwn-` gegen Kollision mit maplibre. */
	:global(.wwn-marker) {
		display: flex;
		flex-direction: column;
		align-items: center;
		gap: 1px;
		cursor: pointer;
		border: 0;
		background: transparent;
		padding: 0;
	}
	:global(.wwn-marker__temp) {
		background: #0f172a;
		color: #fff;
		font-size: 12px;
		font-weight: 600;
		line-height: 1;
		padding: 3px 6px;
		border-radius: 9999px;
		box-shadow: 0 1px 2px rgba(0, 0, 0, 0.3);
	}
	:global(.wwn-marker__arrow) {
		color: #0f172a;
		font-size: 14px;
		line-height: 1;
	}
	:global(.wwn-popup) {
		min-width: 180px;
	}
	:global(.wwn-popup__head) {
		display: flex;
		align-items: center;
		justify-content: space-between;
		gap: 8px;
	}
	:global(.wwn-popup__head h3) {
		font-size: 15px;
		font-weight: 600;
		margin: 0;
	}
	:global(.wwn-popup__badge) {
		background: #f1f5f9;
		color: #334155;
		font-size: 11px;
		font-weight: 500;
		border-radius: 9999px;
		padding: 2px 8px;
		white-space: nowrap;
	}
	:global(.wwn-popup__temp) {
		font-size: 22px;
		font-weight: 700;
		margin: 6px 0;
	}
	:global(.wwn-popup__dl) {
		display: grid;
		grid-template-columns: auto auto;
		gap: 2px 12px;
		font-size: 12px;
		margin: 0;
	}
	:global(.wwn-popup__dl dt) {
		color: #64748b;
	}
	:global(.wwn-popup__dl dd) {
		color: #0f172a;
		margin: 0;
		text-align: right;
	}
	:global(.wwn-popup__nodata) {
		font-size: 12px;
		color: #64748b;
		margin: 6px 0;
	}
	:global(.wwn-popup__link) {
		display: inline-block;
		margin-top: 8px;
		font-size: 12px;
		color: #0f172a;
		text-decoration: underline;
	}
</style>
