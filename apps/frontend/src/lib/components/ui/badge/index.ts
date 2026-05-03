import { tv, type VariantProps } from 'tailwind-variants';

export const badgeVariants = tv({
	base: 'inline-flex items-center rounded-full border px-2.5 py-0.5 text-xs font-semibold transition-colors',
	variants: {
		variant: {
			default: 'bg-primary text-primary-foreground border-transparent',
			secondary: 'bg-secondary text-secondary-foreground border-transparent',
			destructive: 'bg-destructive text-destructive-foreground border-transparent',
			outline: 'text-foreground'
		}
	},
	defaultVariants: {
		variant: 'default'
	}
});

export type BadgeVariant = NonNullable<VariantProps<typeof badgeVariants>['variant']>;

export { default as Badge } from './badge.svelte';
