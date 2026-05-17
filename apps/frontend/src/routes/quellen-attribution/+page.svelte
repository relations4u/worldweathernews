<script lang="ts">
	import { resolve } from '$app/paths';
</script>

<section class="mx-auto max-w-3xl px-4 py-8">
	<h1 class="text-3xl font-bold tracking-tight text-slate-900">Quellen und Attribution</h1>

	<p class="mt-4 text-slate-600">
		worldweathernews.com aggregiert Daten aus offenen meteorologischen und klimatologischen Quellen.
		Diese Seite listet alle eingebundenen Datenquellen mit Lizenz, Zugriffs-Pfad und Datenstand.
	</p>

	<h2 class="mt-8 text-xl font-semibold text-slate-900">Aktive Datenquellen</h2>

	<article class="mt-4 rounded-md border border-slate-200 p-4">
		<h3 class="text-lg font-semibold text-slate-900">Deutscher Wetterdienst (DWD)</h3>
		<p class="mt-2 text-sm text-slate-600">
			Aktuelle Stationsbeobachtungen (POI) für Standorte in Deutschland. Datenfile-Server unter
			<code class="rounded bg-slate-100 px-1 py-0.5 text-xs">opendata.dwd.de</code>; pro Station
			eine halbstündliche CSV-Datei.
		</p>
		<dl class="mt-3 grid grid-cols-[max-content_1fr] gap-x-4 gap-y-1 text-sm text-slate-600">
			<dt class="font-medium text-slate-900">Endpoint</dt>
			<dd>
				<a
					href="https://opendata.dwd.de/weather/weather_reports/poi/"
					rel="noopener noreferrer"
					class="underline hover:text-slate-900">opendata.dwd.de/.../poi/</a
				>
			</dd>
			<dt class="font-medium text-slate-900">Lizenz</dt>
			<dd>
				<a
					href="https://www.dwd.de/DE/service/copyright/copyright_node.html"
					rel="noopener noreferrer"
					class="underline hover:text-slate-900">GeoNutzV (CC-BY-äquivalent für offene Geodaten)</a
				>
			</dd>
			<dt class="font-medium text-slate-900">Attribution</dt>
			<dd>„Datenbasis: Deutscher Wetterdienst, eigene Bearbeitung"</dd>
			<dt class="font-medium text-slate-900">Genutzte Stationen</dt>
			<dd>
				Potsdam-Telegrafenberg (10379), Berlin-Tempelhof (10384), Hamburg-Fuhlsbüttel (10147),
				Brocken (10454), Zugspitze (10961), Helgoland (10015)
			</dd>
			<dt class="font-medium text-slate-900">Genutzte Variablen</dt>
			<dd>
				Lufttemperatur (2 m), Niederschlag (letzte Stunde), Windgeschwindigkeit + Windrichtung
				(10-min-Mittel, 10 m), Luftdruck (auf Meereshöhe reduziert), relative Luftfeuchte
			</dd>
			<dt class="font-medium text-slate-900">Aktualität</dt>
			<dd>Halbstündlich (POI-Dateien werden ~30 Min nach Beobachtungszeitpunkt aktualisiert)</dd>
			<dt class="font-medium text-slate-900">Eingebunden seit</dt>
			<dd>Iteration 2.2 (Mai 2026)</dd>
		</dl>
	</article>

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
			<dd>
				temperature_2m, precipitation, wind_speed_10m, wind_direction_10m, pressure_msl,
				relative_humidity_2m
			</dd>
			<dt class="font-medium text-slate-900">Aktualität</dt>
			<dd>Current alle 10 min, Stundenvorhersage (24 h) alle 60 min</dd>
			<dt class="font-medium text-slate-900">Eingebunden seit</dt>
			<dd>Iteration 2.1 (Mai 2026)</dd>
		</dl>
	</article>

	<article class="mt-4 rounded-md border border-slate-200 p-4">
		<h3 class="text-lg font-semibold text-slate-900">EUMETSAT (Meteosat-Satellitenbilder)</h3>
		<p class="mt-2 text-sm text-slate-600">
			Infrarot-Satellitenbilder (10,8 µm) über Europa für die Satelliten-Ansicht. Die Bilder werden
			server-seitig vom öffentlichen EUMETView-WMS geholt und ausschließlich über
			<code class="rounded bg-slate-100 px-1 py-0.5 text-xs">media.worldweathernews.com</code>
			ausgeliefert — keine direkte Verbindung des Browsers zu EUMETSAT.
		</p>
		<dl class="mt-3 grid grid-cols-[max-content_1fr] gap-x-4 gap-y-1 text-sm text-slate-600">
			<dt class="font-medium text-slate-900">Dienst</dt>
			<dd>
				<a
					href="https://view.eumetsat.int/"
					rel="noopener noreferrer"
					class="underline hover:text-slate-900">EUMETView WMS (view.eumetsat.int)</a
				>
			</dd>
			<dt class="font-medium text-slate-900">Lizenz</dt>
			<dd>
				<a
					href="https://user.eumetsat.int/resources/user-guides/data-registration-and-licensing"
					rel="noopener noreferrer"
					class="underline hover:text-slate-900"
					>EUMETSAT-Datenpolitik — die genutzten Meteosat-Bildprodukte sind kostenfrei und
					lizenzfrei</a
				>
			</dd>
			<dt class="font-medium text-slate-900">Attribution</dt>
			<dd>„© EUMETSAT"</dd>
			<dt class="font-medium text-slate-900">Genutztes Produkt</dt>
			<dd>Meteosat (0°) High Rate SEVIRI IR 10.8 µm, Europa-Ausschnitt</dd>
			<dt class="font-medium text-slate-900">Aktualität</dt>
			<dd>Etwa alle 15 Minuten, rollierendes 24-Stunden-Fenster</dd>
			<dt class="font-medium text-slate-900">Eingebunden seit</dt>
			<dd>Iteration 2.4 (Mai 2026)</dd>
		</dl>
	</article>

	<h2 class="mt-8 text-xl font-semibold text-slate-900">Geplante Datenquellen</h2>
	<p class="mt-3 text-slate-600">
		Die folgenden Quellen sind für die Integration in zukünftigen Iterationen vorgesehen. Endgültige
		Lizenz- und Attribuierungs-Hinweise folgen bei der jeweiligen produktiven Anbindung.
	</p>

	<ul class="mt-3 list-disc space-y-2 pl-6 text-slate-600">
		<li>
			<strong class="text-slate-900">DWD MOSMIX / ICON</strong> — Forecast-Pfade ergänzend zu POI (Iteration
			2.2b und Folge-Klima-Iterationen)
		</li>
		<li>
			<strong class="text-slate-900">NOAA / National Weather Service</strong> — USA, Public Domain
		</li>
		<li>
			<strong class="text-slate-900">Met Office (UK), JMA (Japan), Météo-France</strong> — phasenweise,
			jeweilige Open-Data-Lizenzen
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

	<article class="mt-4 rounded-md border border-slate-200 p-4">
		<h3 class="text-lg font-semibold text-slate-900">OpenFreeMap (OpenStreetMap-Daten)</h3>
		<p class="mt-2 text-sm text-slate-600">
			Die interaktive Stationskarte auf der Wetter-Seite nutzt Vektor-Kartenkacheln von OpenFreeMap
			(Liberty-Style). Die Kacheln werden vom Browser direkt nachgeladen.
		</p>
		<dl class="mt-3 grid grid-cols-[max-content_1fr] gap-x-4 gap-y-1 text-sm text-slate-600">
			<dt class="font-medium text-slate-900">Kachel-Quelle</dt>
			<dd>
				<a
					href="https://openfreemap.org/"
					rel="noopener noreferrer"
					class="underline hover:text-slate-900">tiles.openfreemap.org</a
				>
			</dd>
			<dt class="font-medium text-slate-900">Lizenz</dt>
			<dd>
				OpenFreeMap unter <a
					href="https://github.com/hyperknot/openfreemap/blob/main/LICENSE"
					rel="noopener noreferrer"
					class="underline hover:text-slate-900">MIT</a
				>; Kartendaten von
				<a
					href="https://www.openstreetmap.org/copyright"
					rel="noopener noreferrer"
					class="underline hover:text-slate-900">OpenStreetMap unter ODbL</a
				>
			</dd>
			<dt class="font-medium text-slate-900">Attribution</dt>
			<dd>„OpenFreeMap © OpenMapTiles Data from OpenStreetMap"</dd>
			<dt class="font-medium text-slate-900">Datenschutz</dt>
			<dd>
				Setzt keine Cookies und kein Tracking. Beim Laden der Kacheln wird — wie bei jeder
				HTTP-Anfrage — die IP-Adresse des Browsers an den Kachel-Server übertragen. Siehe <a
					href={resolve('/datenschutz')}
					class="underline hover:text-slate-900">Datenschutzerklärung</a
				>.
			</dd>
			<dt class="font-medium text-slate-900">Eingebunden seit</dt>
			<dd>Iteration 2.3 (Mai 2026)</dd>
		</dl>
	</article>

	<p class="mt-3 text-slate-600">
		<strong class="text-slate-900">Natural Earth</strong> — Public Domain, geplant als ergänzende Karten-Basemap.
	</p>

	<h2 class="mt-8 text-xl font-semibold text-slate-900">Software</h2>
	<p class="mt-3 text-slate-600">
		worldweathernews.com selbst ist Open-Source unter der AGPL-3.0. Eine Liste der eingesetzten
		Open-Source-Bibliotheken folgt mit der ersten Release-Version.
	</p>
</section>
