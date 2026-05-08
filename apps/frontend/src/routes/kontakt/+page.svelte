<script lang="ts">
	import { resolve } from '$app/paths';
</script>

<section class="mx-auto max-w-3xl px-4 py-8">
	<h1 class="text-3xl font-bold tracking-tight text-slate-900">Kontakt</h1>

	<p class="mt-4 text-slate-600">
		worldweathernews.com ist ein kleines, ehrenamtlich betriebenes Forschungs-Projekt. Wir freuen
		uns über Hinweise, Datenquellen-Vorschläge und konstruktive Kritik. Bitte beachten: kurze
		Antwortzeiten können wir nicht zusichern.
	</p>

	<h2 class="mt-8 text-xl font-semibold text-slate-900">E-Mail</h2>
	<p class="mt-3 text-slate-600">Allgemeine Anfragen, Feedback, Fehlerberichte:</p>
	<p class="mt-2">
		<a
			href="mailto:ops@worldweathernews.com"
			class="font-medium text-slate-900 underline underline-offset-2 hover:no-underline"
		>
			ops@worldweathernews.com
		</a>
	</p>

	<h2 class="mt-8 text-xl font-semibold text-slate-900">Postanschrift</h2>
	<p class="mt-3 text-slate-600">
		Die ladungsfähige Anschrift findet sich im
		<a
			href={resolve('/impressum')}
			class="font-medium text-slate-900 underline underline-offset-2 hover:no-underline"
		>
			Impressum
		</a>.
	</p>

	<h2 class="mt-8 text-xl font-semibold text-slate-900">Datenschutz-Anfragen</h2>
	<p class="mt-3 text-slate-600">
		Auskunfts-, Berichtigungs- und Löschanfragen gemäß DSGVO bitte ebenfalls an
		<a
			href="mailto:ops@worldweathernews.com"
			class="font-medium text-slate-900 underline underline-offset-2 hover:no-underline"
		>
			ops@worldweathernews.com
		</a>. Details zur Datenverarbeitung in der
		<a
			href={resolve('/datenschutz')}
			class="font-medium text-slate-900 underline underline-offset-2 hover:no-underline"
		>
			Datenschutzerklärung
		</a>.
	</p>
</section>
