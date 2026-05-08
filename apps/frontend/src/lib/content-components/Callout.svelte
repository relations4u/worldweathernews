<script lang="ts">
	import * as m from '$lib/paraglide/messages';

	type Variant = 'info' | 'warning' | 'note';

	type Props = {
		variant?: Variant;
		title?: string;
		children: import('svelte').Snippet;
	};

	let { variant = 'info', title, children }: Props = $props();

	const styles: Record<Variant, { container: string; titleColor: string }> = {
		info: {
			container: 'border-sky-200 bg-sky-50 text-sky-900',
			titleColor: 'text-sky-900'
		},
		warning: {
			container: 'border-amber-200 bg-amber-50 text-amber-900',
			titleColor: 'text-amber-900'
		},
		note: {
			container: 'border-slate-200 bg-slate-50 text-slate-900',
			titleColor: 'text-slate-900'
		}
	};

	const variantLabel: Record<Variant, () => string> = {
		info: m.callout_info,
		warning: m.callout_warning,
		note: m.callout_note
	};
</script>

<aside class="not-prose my-4 rounded-md border px-4 py-3 text-sm {styles[variant].container}">
	<p class="font-semibold {styles[variant].titleColor}">
		{title ?? variantLabel[variant]()}
	</p>
	<div class="mt-1">
		{@render children()}
	</div>
</aside>
