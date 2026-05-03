<script lang="ts">
	import { cn } from '$lib/utils';
	import { badgeVariants, type BadgeVariant } from './index';

	interface Props {
		variant?: BadgeVariant;
		class?: string;
		children?: import('svelte').Snippet;
	}

	let { variant = 'default', class: className, children }: Props = $props();
</script>

<span class={cn(badgeVariants({ variant }), className)}>
	{@render children?.()}
</span>
