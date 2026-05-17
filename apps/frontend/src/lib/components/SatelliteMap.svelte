<script lang="ts">
	import { onMount, onDestroy } from 'svelte';
	import { browser } from '$app/environment';
	import * as m from '$lib/paraglide/messages';
	import { getLocale } from '$lib/paraglide/runtime';
	import { MAP_STYLE_URL, MAP_INITIAL_VIEW } from '$lib/config/map';
	import type { SatIndex } from '$lib/config/satellite';
	import { bboxCorners, resolveFrameIndex } from '$lib/satellite';
	// Reiner Typ-Import — wird vom Compiler gelöscht, landet NICHT im
	// Bundle. Laufzeit-Lib kommt nur über den dynamischen import() (lazy).
	import type { Map as MapLibreMap, ImageSource } from 'maplibre-gl';

	let { index }: { index: SatIndex } = $props();

	const SRC_ID = 'wwn-sat';
	const LAYER_ID = 'wwn-sat-layer';

	let container: HTMLDivElement;
	let map: MapLibreMap | undefined;
	let ready = $state(false);
	let loading = $state(true);

	// `selectedIdx === null` ⇒ folge dem neuesten Frame (letzter im
	// aufsteigend sortierten Array). Slider/Loop setzen einen konkreten
	// Index. Kein Prop→$state-Copy (reaktiv-korrekt, lint-clean).
	let selectedIdx = $state<number | null>(null);
	let opacity = $state(0.85);
	let playing = $state(false);
	let timer: ReturnType<typeof setInterval> | undefined;

	const frameIdx = $derived(resolveFrameIndex(selectedIdx, index.frames.length));
	const current = $derived(index.frames[frameIdx]);

	function formatTime(iso: string): string {
		const locale = getLocale() === 'en' ? 'en-GB' : 'de-DE';
		return new Date(iso).toLocaleString(locale, {
			day: '2-digit',
			month: '2-digit',
			hour: '2-digit',
			minute: '2-digit'
		});
	}

	function stopLoop(): void {
		if (timer) {
			clearInterval(timer);
			timer = undefined;
		}
		playing = false;
	}

	function toggleLoop(): void {
		if (playing) {
			stopLoop();
			return;
		}
		if (index.frames.length < 2) return;
		playing = true;
		timer = setInterval(() => {
			selectedIdx = (frameIdx + 1) % index.frames.length;
		}, 700);
	}

	// Frame-Wechsel → das Bild der bestehenden Image-Source tauschen
	// (kein Source-/Layer-Neuaufbau).
	$effect(() => {
		const url = current?.url;
		if (!ready || !map || !url) return;
		const src = map.getSource(SRC_ID) as ImageSource | undefined;
		src?.updateImage({ url });
	});

	$effect(() => {
		if (!ready || !map) return;
		map.setPaintProperty(LAYER_ID, 'raster-opacity', opacity);
	});

	onMount(async () => {
		if (!browser || index.frames.length === 0) {
			loading = false;
			return;
		}
		const { Map, NavigationControl } = await import('maplibre-gl');
		// CSS lazy mit dem maplibre-Chunk — bleibt aus dem Default-Bundle.
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
			instance.addSource(SRC_ID, {
				type: 'image',
				url: index.frames[frameIdx].url,
				coordinates: bboxCorners(index.bbox)
			});
			instance.addLayer({
				id: LAYER_ID,
				type: 'raster',
				source: SRC_ID,
				paint: { 'raster-opacity': opacity, 'raster-fade-duration': 0 }
			});
			ready = true;
			loading = false;
		});
	});

	onDestroy(() => {
		stopLoop();
		map?.remove();
	});
</script>

<div>
	<div class="relative">
		<!-- Server rendert nur den leeren Wrapper (ssr=false); Map-Init
		     erst client-side in onMount. Feste Höhe gegen Layout-Shift. -->
		<div
			bind:this={container}
			class="h-[360px] w-full overflow-hidden rounded-lg border border-slate-200 sm:h-[520px]"
			role="region"
			aria-label={m.sat_map_label()}
		></div>
		{#if loading}
			<div
				class="pointer-events-none absolute inset-0 flex items-center justify-center text-sm text-slate-500"
			>
				{m.sat_loading()}
			</div>
		{/if}
	</div>

	{#if index.frames.length > 0}
		<div class="mt-4 flex flex-wrap items-center gap-x-6 gap-y-3 text-sm">
			<div class="flex items-center gap-2">
				<button
					type="button"
					onclick={toggleLoop}
					disabled={index.frames.length < 2}
					class="rounded-md border border-slate-300 px-3 py-1 font-medium text-slate-700 hover:bg-slate-50 disabled:opacity-50"
				>
					{playing ? m.sat_pause() : m.sat_play()}
				</button>
				<label class="flex items-center gap-2">
					<span class="text-slate-600">{m.sat_time()}</span>
					<input
						type="range"
						min="0"
						max={index.frames.length - 1}
						value={frameIdx}
						oninput={(e) => {
							stopLoop();
							selectedIdx = Number(e.currentTarget.value);
						}}
						class="w-40 sm:w-56"
						aria-label={m.sat_time()}
					/>
				</label>
				<span class="text-slate-900 tabular-nums">{current ? formatTime(current.time) : '—'}</span>
			</div>

			<label class="flex items-center gap-2">
				<span class="text-slate-600">{m.sat_opacity()}</span>
				<input
					type="range"
					min="0.2"
					max="1"
					step="0.05"
					bind:value={opacity}
					class="w-32"
					aria-label={m.sat_opacity()}
				/>
			</label>
		</div>
	{/if}
</div>
