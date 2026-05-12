<section class="mx-auto max-w-3xl px-4 py-8">
	<h1 class="text-3xl font-bold tracking-tight text-slate-900">Quellen und Attribution</h1>

	<p class="mt-4 text-slate-600">
		worldweathernews.com aggregiert Daten aus offenen meteorologischen und klimatologischen Quellen.
		Diese Seite listet alle eingebundenen Datenquellen mit Lizenz, Zugriffs-Pfad und Datenstand.
	</p>

	<h2 class="mt-8 text-xl font-semibold text-slate-900">Aktive Datenquellen</h2>

	<article class="mt-4 rounded-md border border-slate-200 p-4">
		<h3 class="text-lg font-semibold text-slate-900">Open-Meteo</h3>
		<dl class="mt-3 grid grid-cols-[max-content_1fr] gap-x-4 gap-y-1 text-sm text-slate-600">
			<dt class="font-medium text-slate-900">API</dt>
			<dd>
				<a
					href="https://open-meteo.com/en/docs"
					rel="noopener noreferrer"
					class="underline hover:text-slate-900">open-meteo.com</a
				>
			</dd>
			<dt class="font-medium text-slate-900">Lizenz</dt>
			<dd>
				<a
					href="https://creativecommons.org/licenses/by/4.0/"
					rel="noopener noreferrer"
					class="underline hover:text-slate-900">CC BY 4.0</a
				>
			</dd>
			<dt class="font-medium text-slate-900">Attribution</dt>
			<dd>„Daten von Open-Meteo.com, CC BY 4.0"</dd>
			<dt class="font-medium text-slate-900">Genutzte Variablen</dt>
			<dd>temperature_2m, precipitation, wind_speed_10m, wind_direction_10m</dd>
			<dt class="font-medium text-slate-900">Aktualität</dt>
			<dd>Current alle 10 min, Stundenvorhersage (24 h) alle 60 min</dd>
			<dt class="font-medium text-slate-900">Eingebunden seit</dt>
			<dd>Iteration 2.1 (Mai 2026)</dd>
		</dl>
	</article>

	<h2 class="mt-8 text-xl font-semibold text-slate-900">Geplante Datenquellen</h2>
	<p class="mt-3 text-slate-600">
		Die folgenden Quellen sind für die Integration in zukünftigen Iterationen vorgesehen. Endgültige
		Lizenz- und Attribuierungs-Hinweise folgen bei der jeweiligen produktiven Anbindung.
	</p>

	<ul class="mt-3 list-disc space-y-2 pl-6 text-slate-600">
		<li>
			<strong class="text-slate-900">Deutscher Wetterdienst (DWD)</strong> — OpenData-Portal, MOSMIX,
			ICON-Modelle, CC-BY-4.0 (Iteration 2.2)
		</li>
		<li>
			<strong class="text-slate-900">NOAA / National Weather Service</strong> — USA, Public Domain
		</li>
		<li>
			<strong class="text-slate-900">Met Office (UK), JMA (Japan), Météo-France</strong> — phasenweise,
			jeweilige Open-Data-Lizenzen
		</li>
		<li>
			<strong class="text-slate-900">EUMETSAT</strong> — Satellitenbilder, Lizenzbedingungen je nach Produkt
		</li>
		<li>
			<strong class="text-slate-900">USGS</strong> — Erdbebendaten, Public Domain
		</li>
		<li>
			<strong class="text-slate-900">NOAA Space Weather Prediction Center</strong> — Aurora-Vorhersagen,
			Public Domain
		</li>
	</ul>

	<h2 class="mt-8 text-xl font-semibold text-slate-900">Karten und Geodaten</h2>
	<ul class="mt-3 list-disc space-y-2 pl-6 text-slate-600">
		<li>
			<strong class="text-slate-900">Natural Earth</strong> — Public Domain (geplant für Karten-Basemaps)
		</li>
		<li>
			<strong class="text-slate-900">OpenStreetMap-Daten</strong> — ODbL (geplant, sofern Standort-Karten
			ergänzt werden)
		</li>
	</ul>

	<h2 class="mt-8 text-xl font-semibold text-slate-900">Software</h2>
	<p class="mt-3 text-slate-600">
		worldweathernews.com selbst ist Open-Source unter der AGPL-3.0. Eine Liste der eingesetzten
		Open-Source-Bibliotheken folgt mit der ersten Release-Version.
	</p>
</section>
