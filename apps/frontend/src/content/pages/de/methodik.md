---
title: 'Methodik'
slug: methodik
lang: de-de
lead: 'Wie worldweathernews.com Wetterdaten zusammenführt, einordnet und veröffentlicht.'
updated_at: 2026-05-08
status: published
---

<script lang="ts">
	import DataSourceCard from '$lib/content-components/DataSourceCard.svelte';
	import Callout from '$lib/content-components/Callout.svelte';
</script>

worldweathernews.com versteht sich als kuratierte Aggregation: keine eigenen
Vorhersagen, sondern eine durchsuchbare und meteorologisch eingeordnete Sicht
auf die Outputs nationaler Wetterdienste und offener Datenquellen.

<Callout variant="info" title="Forschungs-Phase">Die Plattform ist aktuell im Aufbau. Es gibt kein SLA, keine garantierte Datenaktualität und kein automatisches Failover. Sobald genug Quellen produktiv und die Plattform stabil ist, erfolgt die Migration in eine echte Produktionsumgebung.</Callout>

## Datenquellen

Die Plattform startet mit ausgewählten offenen Quellen und erweitert das
Spektrum schrittweise. Eine vollständige Liste aller produktiv eingebundenen
Quellen mit Lizenz-Hinweisen findet sich auf der Seite **Quellen und
Attribution**.

<DataSourceCard name="Open-Meteo" url="https://open-meteo.com" license="CC-BY-4.0 (historisch), öffentlich (Live-API)" region="EU-basiert, global" status="planned">Erste primäre Quelle für die Forschungs-Phase. Globale Wetterdaten und Modell-Outputs ohne API-Key, EU-betriebene Infrastruktur.</DataSourceCard>

<DataSourceCard name="Deutscher Wetterdienst (DWD) — OpenData" url="https://opendata.dwd.de" license="CC-BY-4.0 (Geo-Nutzungs-VO)" region="Deutschland, Mitteleuropa" status="planned">MOSMIX-Vorhersagen, ICON-Modelle, Stationsbeobachtungen, RADOLAN-Niederschlag. Authoritative Quelle für deutsche Wetterdaten.</DataSourceCard>

## Methodik-Versprechen

- Quellen werden **transparent** ausgewiesen, jede Anzeige verlinkt zurück zur
  Originalquelle.
- Bearbeitungs- und Aggregationsschritte sind **dokumentiert** und im Code
  nachvollziehbar (AGPL-3.0).
- Bei abweichenden Werten zwischen Quellen werden Diskrepanzen **nicht
  verschleiert**, sondern nebeneinander dargestellt.
- Klima-Indikatoren (z.B. Anomalien, Trends) werden mit ihrer **Methodik
  und Datengrundlage** ausgewiesen.

## Aktualisierung

Diese Methodik-Seite wird mit jedem signifikanten Schritt der
Plattform-Entwicklung aktualisiert. Das `updated_at`-Datum oben spiegelt den
Stand der jeweils aktuellen Veröffentlichung.
