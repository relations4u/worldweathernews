<script lang="ts">
	import { locales, getLocale, setLocale, type Locale } from '$lib/i18n';
	import * as m from '$lib/paraglide/messages';

	const tagLabel: Record<Locale, string> = {
		'de-de': m.locale_de(),
		en: m.locale_en()
	};

	function onChange(event: Event) {
		const target = event.currentTarget as HTMLSelectElement;
		setLocale(target.value as Locale);
	}
</script>

<div class="relative">
	<label for="locale-switcher" class="sr-only">{m.locale_switcher_label()}</label>
	<select
		id="locale-switcher"
		value={getLocale()}
		onchange={onChange}
		class="rounded-md border border-slate-300 bg-white px-2 py-1 text-sm text-slate-700 hover:bg-slate-50 focus:outline-2 focus:outline-offset-2 focus:outline-slate-600"
	>
		{#each locales as tag (tag)}
			<option value={tag}>{tagLabel[tag]}</option>
		{/each}
	</select>
</div>
