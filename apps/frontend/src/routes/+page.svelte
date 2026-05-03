<script lang="ts">
	import { onMount } from 'svelte';
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

<section class="py-16 text-center">
	<h1 class="text-5xl font-bold tracking-tight">worldweathernews</h1>
	<p class="mt-4 text-lg text-muted-foreground">
		A global community for weather and climate observations.
	</p>
	<p class="mt-2 text-sm text-muted-foreground">Coming soon. Currently building the platform.</p>
</section>

<section class="text-center text-sm">
	{#if status === 'pending'}
		<Badge variant="outline">Connecting to backend…</Badge>
	{:else if status === 'ok'}
		<Badge>Backend connected</Badge>
		{#if traceId}
			<span class="ml-2 text-muted-foreground">Trace: {traceId}</span>
		{/if}
	{:else}
		<Badge variant="destructive">Backend offline</Badge>
		{#if errorMsg}
			<span class="ml-2 text-muted-foreground">{errorMsg}</span>
		{/if}
	{/if}
</section>
