<script lang="ts">
	import { resolve } from '$app/paths';
</script>

<section class="mx-auto max-w-3xl px-4 py-8">
	<h1 class="text-3xl font-bold tracking-tight text-slate-900">Impressum</h1>

	<p class="mt-4 text-slate-600">
		Anbieterkennzeichnung gemäß § 5 Digitale-Dienste-Gesetz (DDG) sowie Verantwortlicher im Sinne
		von § 18 Abs. 2 Medienstaatsvertrag (MStV).
	</p>

	<h2 class="mt-8 text-xl font-semibold text-slate-900">Diensteanbieter</h2>
	<address class="mt-3 text-slate-600 not-italic">
		<strong class="block text-slate-900">Heinz W. Richter</strong>
		<span>Lessingstraße 47</span><br />
		<span>14482 Potsdam</span><br />
		<span>Deutschland</span>
	</address>

	<h2 class="mt-8 text-xl font-semibold text-slate-900">Kontakt</h2>
	<dl class="mt-3 space-y-1 text-slate-600">
		<div class="flex gap-2">
			<dt class="font-medium text-slate-900">E-Mail:</dt>
			<dd>
				<a
					href="mailto:ops@worldweathernews.com"
					class="underline underline-offset-2 hover:no-underline"
				>
					ops@worldweathernews.com
				</a>
			</dd>
		</div>
		<div class="flex gap-2">
			<dt class="font-medium text-slate-900">Telefon:</dt>
			<dd>
				<a href="tel:+49331719207" class="underline underline-offset-2 hover:no-underline">
					+49 331 719207
				</a>
			</dd>
		</div>
	</dl>

	<h2 class="mt-8 text-xl font-semibold text-slate-900">
		Verantwortlicher i.S.v. § 18 Abs. 2 MStV
	</h2>
	<address class="mt-3 text-slate-600 not-italic">
		Heinz W. Richter<br />
		Anschrift wie oben
	</address>

	<h2 class="mt-8 text-xl font-semibold text-slate-900">Berufsbezeichnung und Aufsicht</h2>
	<!-- TODO Maintainer: prüfen, ob Berufs-/Steuerangaben Pflicht sind. Bei rein privatem Hobby-Betrieb ohne wirtschaftliche Tätigkeit i.d.R. nicht erforderlich. Bei Anmeldung als Kleinunternehmer: USt-ID/Steuernummer ergänzen. -->
	<p class="mt-3 text-slate-600">
		worldweathernews.com wird privat und ohne wirtschaftliche Gewinnabsicht betrieben. Eine
		Eintragung in einem Berufsregister besteht nicht.
	</p>

	<h2 class="mt-8 text-xl font-semibold text-slate-900">Haftung für Inhalte</h2>
	<p class="mt-3 text-slate-600">
		Die Inhalte dieser Seiten wurden mit größtmöglicher Sorgfalt erstellt. Für die Richtigkeit,
		Vollständigkeit und Aktualität der Inhalte können wir jedoch keine Gewähr übernehmen.
		Insbesondere übernehmen wir keine Verantwortung für Wetter- oder Klimadaten, die von externen
		Anbietern eingebunden werden. Diese sind zur eigenen Information bestimmt und ersetzen keine
		offiziellen Warnungen oder Empfehlungen der nationalen Wetterdienste.
	</p>

	<h2 class="mt-8 text-xl font-semibold text-slate-900">Haftung für Links</h2>
	<p class="mt-3 text-slate-600">
		Unser Angebot enthält Links zu externen Websites Dritter, auf deren Inhalte wir keinen Einfluss
		haben. Deshalb können wir für diese fremden Inhalte auch keine Gewähr übernehmen. Für die
		Inhalte der verlinkten Seiten ist stets der jeweilige Anbieter oder Betreiber der Seiten
		verantwortlich.
	</p>

	<h2 class="mt-8 text-xl font-semibold text-slate-900">Urheberrecht</h2>
	<p class="mt-3 text-slate-600">
		Der von uns selbst erstellte Quellcode dieser Plattform ist unter der GNU Affero General Public
		License v3.0 (AGPL-3.0) lizenziert. Eingebundene Wetterdaten unterliegen den jeweiligen Lizenzen
		ihrer Quellen — siehe
		<a
			href={resolve('/quellen-attribution')}
			class="underline underline-offset-2 hover:no-underline"
		>
			Quellen und Attribution
		</a>.
	</p>
</section>
