<script lang="ts">
	import * as m from '$lib/paraglide/messages';

	type Status = 'active' | 'planned';

	type Props = {
		name: string;
		license?: string;
		url?: string;
		region?: string;
		status?: Status;
		children?: import('svelte').Snippet;
	};

	let { name, license, url, region, status = 'planned', children }: Props = $props();

	const statusLabel: Record<Status, () => string> = {
		active: m.data_source_status_active,
		planned: m.data_source_status_planned
	};
</script>

<aside
	class="not-prose my-4 rounded-md border border-slate-200 bg-slate-50 px-4 py-3 text-sm text-slate-700"
>
	<header class="flex flex-wrap items-baseline justify-between gap-2">
		<h3 class="text-base font-semibold text-slate-900">
			{#if url}
				<!-- url ist immer eine externe Datenquellen-URL — resolve() greift hier nicht. -->
				<!-- eslint-disable svelte/no-navigation-without-resolve -->
				<a
					href={url}
					rel="noopener noreferrer"
					target="_blank"
					class="underline underline-offset-2 hover:no-underline"
				>
					{name}
				</a>
				<!-- eslint-enable svelte/no-navigation-without-resolve -->
			{:else}
				{name}
			{/if}
		</h3>
		<span
			class="rounded-full px-2 py-0.5 text-xs font-medium {status === 'active'
				? 'bg-emerald-100 text-emerald-900'
				: 'bg-amber-100 text-amber-900'}"
		>
			{statusLabel[status]()}
		</span>
	</header>

	<dl class="mt-2 grid gap-x-4 gap-y-0.5 text-xs text-slate-600 sm:grid-cols-[auto_1fr]">
		{#if license}
			<dt class="font-medium text-slate-700">Lizenz</dt>
			<dd>{license}</dd>
		{/if}
		{#if region}
			<dt class="font-medium text-slate-700">Region</dt>
			<dd>{region}</dd>
		{/if}
	</dl>

	{#if children}
		<div class="mt-2 text-slate-600">
			{@render children()}
		</div>
	{/if}
</aside>
