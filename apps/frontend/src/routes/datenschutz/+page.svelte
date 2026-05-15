<script lang="ts">
	import { resolve } from '$app/paths';
</script>

<section class="mx-auto max-w-3xl px-4 py-8">
	<h1 class="text-3xl font-bold tracking-tight text-slate-900">Datenschutzerklärung</h1>

	<p class="mt-4 text-slate-600">
		Diese Datenschutzerklärung informiert über Art, Umfang und Zweck der Verarbeitung
		personenbezogener Daten auf worldweathernews.com gemäß Art. 13/14 DSGVO. Stand der Erklärung:
		Forschungs-Phase, ohne Tracking, ohne Werbung, ohne Drittanbieter-Embeds.
	</p>

	<!-- TODO Maintainer: Nach erstem juristischen Review (eRecht24-Generator oder Anwalt) prüfen, ob alle Pflichtangaben vollständig sind. Bei Aktivierung von Plausible/Analytics oder externen Embeds: Abschnitte ergänzen, Cookie-Banner-Verhalten anpassen, Consent-Logik prüfen. -->

	<h2 class="mt-8 text-xl font-semibold text-slate-900">1. Verantwortlicher</h2>
	<address class="mt-3 text-slate-600 not-italic">
		<strong class="block text-slate-900">Heinz W. Richter</strong>
		<span>Lessingstraße 47</span><br />
		<span>14482 Potsdam</span><br />
		<span>Deutschland</span><br />
		E-Mail:
		<a
			href="mailto:ops@worldweathernews.com"
			class="underline underline-offset-2 hover:no-underline"
		>
			ops@worldweathernews.com
		</a>
	</address>

	<h2 class="mt-8 text-xl font-semibold text-slate-900">2. Datenschutzbeauftragter</h2>
	<p class="mt-3 text-slate-600">
		Ein Datenschutzbeauftragter ist nach § 38 BDSG nicht erforderlich. Anfragen zum Datenschutz
		bitte direkt an die oben genannte E-Mail-Adresse.
	</p>

	<h2 class="mt-8 text-xl font-semibold text-slate-900">3. Server-Logfiles</h2>
	<p class="mt-3 text-slate-600">
		Beim Aufruf der Plattform werden vom Webserver automatisch folgende Daten in Logfiles
		gespeichert:
	</p>
	<ul class="mt-2 list-disc space-y-1 pl-6 text-slate-600">
		<li>IP-Adresse des anfragenden Rechners</li>
		<li>Datum und Uhrzeit des Zugriffs</li>
		<li>aufgerufene URL und HTTP-Methode</li>
		<li>HTTP-Statuscode der Antwort</li>
		<li>übertragene Datenmenge</li>
		<li>Referrer-URL (sofern vom Browser übermittelt)</li>
		<li>User-Agent-String (Browser- und Betriebssystem-Information)</li>
	</ul>
	<p class="mt-3 text-slate-600">
		Rechtsgrundlage ist Art. 6 Abs. 1 lit. f DSGVO (berechtigtes Interesse an Betriebssicherheit,
		Fehleranalyse, Schutz vor Missbrauch). Speicherdauer: maximal 14 Tage, danach automatische
		Löschung.
	</p>

	<h2 class="mt-8 text-xl font-semibold text-slate-900">4. Cookies und lokaler Speicher</h2>
	<p class="mt-3 text-slate-600">
		Die Plattform setzt aktuell ausschließlich technisch notwendige Cookies sowie
		localStorage-Einträge. Es findet kein Tracking, keine Reichweitenmessung und keine
		Werbe-Personalisierung statt.
	</p>
	<table class="mt-4 w-full border-collapse text-sm text-slate-600">
		<thead class="border-b border-slate-200 text-left text-slate-900">
			<tr>
				<th class="py-2 pr-4 font-semibold">Bezeichnung</th>
				<th class="py-2 pr-4 font-semibold">Zweck</th>
				<th class="py-2 font-semibold">Speicherdauer</th>
			</tr>
		</thead>
		<tbody>
			<tr class="border-b border-slate-100">
				<td class="py-2 pr-4">wwn-cookie-consent-v1</td>
				<td class="py-2 pr-4">Speichert Cookie-Einwilligung (localStorage)</td>
				<td class="py-2">unbefristet (vom Nutzer löschbar)</td>
			</tr>
			<tr>
				<td class="py-2 pr-4">wwn-research-banner-closed</td>
				<td class="py-2 pr-4">Forschungs-Phase-Hinweis ausgeblendet (localStorage)</td>
				<td class="py-2">unbefristet (vom Nutzer löschbar)</td>
			</tr>
		</tbody>
	</table>
	<p class="mt-3 text-slate-600">
		Cookie-Einstellungen können jederzeit in der
		<a
			href={resolve('/cookie-einstellungen')}
			class="underline underline-offset-2 hover:no-underline"
		>
			Cookie-Verwaltung
		</a>
		angepasst werden.
	</p>

	<h2 class="mt-8 text-xl font-semibold text-slate-900">
		5. Auftragsverarbeiter und Drittland-Transfers
	</h2>
	<p class="mt-3 text-slate-600">Die Plattform setzt folgende Dienstleister ein:</p>
	<ul class="mt-2 list-disc space-y-2 pl-6 text-slate-600">
		<li>
			<strong class="text-slate-900">Cloudflare, Inc.</strong> (USA): DNS-Auflösung für die Domain. Aktuell
			im DNS-only-Modus, ohne Reverse-Proxy. Datenübermittlung in die USA auf Grundlage der EU-Standardvertragsklauseln
			(SCC) und des EU-US Data Privacy Framework.
		</li>
		<li>
			<strong class="text-slate-900">Hetzner Online GmbH</strong> (Deutschland, Falkenstein): Object
			Storage für Bild- und Media-Assets unter <em>media.worldweathernews.com</em>. Verarbeitung
			ausschließlich in Deutschland. Auftragsverarbeitungsvertrag (AVV) gemäß Art. 28 DSGVO
			geschlossen.
		</li>
		<li>
			<strong class="text-slate-900">Joker.com</strong> (Deutschland): Domain-Registrar für worldweathernews.com.
		</li>
		<li>
			<strong class="text-slate-900">Proton AG</strong> (Schweiz): E-Mail-Provider für Adressen unter
			@worldweathernews.com. Schweiz gilt nach Art. 45 DSGVO als sicheres Drittland mit Angemessenheitsbeschluss.
		</li>
		<li>
			<strong class="text-slate-900">OpenFreeMap</strong> (EU-basiert): Kartenkachel-Server für die
			interaktive Stationskarte auf der Wetter-Seite. Die Kacheln werden vom Browser direkt geladen;
			dabei wird — wie bei jeder HTTP-Anfrage — die IP-Adresse an den Kachel-Server übermittelt.
			OpenFreeMap setzt keine Cookies und betreibt kein Tracking. Es werden keine weiteren
			personenbezogenen Daten übertragen. Details und Lizenz in der
			<a href={resolve('/quellen-attribution')} class="underline hover:text-slate-900"
				>Quellen- und Attributionsübersicht</a
			>.
		</li>
	</ul>
	<p class="mt-3 text-slate-600">
		Hosting der Plattform selbst läuft auf eigener Infrastruktur in Deutschland. Es findet keine
		Übermittlung der Plattform-Logs an externe Cloud-Provider statt.
	</p>

	<h2 class="mt-8 text-xl font-semibold text-slate-900">6. Backend-API und Trace-IDs</h2>
	<p class="mt-3 text-slate-600">
		Bei Anfragen an die Backend-API (api.research.worldweathernews.com) wird zur Fehleranalyse eine
		zufällige Trace-ID erzeugt und in den Antwort-Headern übermittelt. Diese Trace-ID erlaubt es,
		einzelne Anfragen in den Server-Logs nachzuvollziehen, enthält aber keine personenbezogenen
		Daten.
	</p>

	<h2 class="mt-8 text-xl font-semibold text-slate-900">7. Ihre Rechte als betroffene Person</h2>
	<p class="mt-3 text-slate-600">
		Sie haben gegenüber dem Verantwortlichen folgende Rechte hinsichtlich Ihrer personenbezogenen
		Daten:
	</p>
	<ul class="mt-2 list-disc space-y-1 pl-6 text-slate-600">
		<li>Recht auf Auskunft (Art. 15 DSGVO)</li>
		<li>Recht auf Berichtigung (Art. 16 DSGVO)</li>
		<li>Recht auf Löschung (Art. 17 DSGVO)</li>
		<li>Recht auf Einschränkung der Verarbeitung (Art. 18 DSGVO)</li>
		<li>Recht auf Datenübertragbarkeit (Art. 20 DSGVO)</li>
		<li>Widerspruchsrecht gegen die Verarbeitung (Art. 21 DSGVO)</li>
		<li>
			Recht auf Beschwerde bei einer Aufsichtsbehörde (Art. 77 DSGVO). Zuständig ist die
			Landesbeauftragte für den Datenschutz und für das Recht auf Akteneinsicht Brandenburg (LDA
			Brandenburg), Stahnsdorfer Damm 77, 14532 Kleinmachnow,
			<a
				href="mailto:Poststelle@LDA.Brandenburg.de"
				class="underline underline-offset-2 hover:no-underline"
			>
				Poststelle@LDA.Brandenburg.de
			</a>.
		</li>
	</ul>
	<p class="mt-3 text-slate-600">
		Anfragen zur Ausübung dieser Rechte richten Sie bitte an
		<a
			href="mailto:ops@worldweathernews.com"
			class="underline underline-offset-2 hover:no-underline"
		>
			ops@worldweathernews.com
		</a>.
	</p>

	<h2 class="mt-8 text-xl font-semibold text-slate-900">8. Änderungen dieser Erklärung</h2>
	<p class="mt-3 text-slate-600">
		Diese Datenschutzerklärung wird angepasst, wenn sich die Datenverarbeitung ändert (z.B. bei
		Aktivierung neuer Dienste). Maßgeblich ist immer die zum Zeitpunkt des Zugriffs auf der
		Plattform verfügbare Version.
	</p>
</section>
