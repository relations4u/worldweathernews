<script lang="ts">
	import '../app.css';
	import { resolve } from '$app/paths';
	import favicon from '$lib/assets/favicon.svg';

	let { children } = $props();
</script>

<svelte:head>
	<link rel="icon" href={favicon} />
</svelte:head>

<div class="flex min-h-screen flex-col bg-background text-foreground">
	<header class="border-b">
		<div class="container mx-auto flex items-center justify-between px-4 py-4">
			<a href={resolve('/')} class="text-xl font-bold">worldweathernews</a>
			<nav class="text-sm text-muted-foreground">
				<span>WIP</span>
			</nav>
		</div>
	</header>

	<main class="container mx-auto flex-1 px-4 py-8">
		{@render children()}
	</main>

	<footer class="border-t text-sm text-muted-foreground">
		<div class="container mx-auto px-4 py-4">
			© {new Date().getFullYear()} worldweathernews — under construction
		</div>
	</footer>
</div>
