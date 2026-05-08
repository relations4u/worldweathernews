---
title: 'Methodology'
slug: methodik
lang: en
lead: 'How worldweathernews.com aggregates, contextualizes and publishes weather data.'
updated_at: 2026-05-08
status: published
---

<script lang="ts">
	import DataSourceCard from '$lib/content-components/DataSourceCard.svelte';
	import Callout from '$lib/content-components/Callout.svelte';
</script>

worldweathernews.com is a curated aggregator: we don't produce our own
forecasts. Instead we offer a searchable, meteorologically annotated view of
the outputs from national weather services and open data sources.

<Callout variant="info" title="Research phase">The platform is currently being built. There is no SLA, no guaranteed data freshness and no automatic failover. Once enough sources are live and the platform is stable, we'll migrate to a true production environment.</Callout>

## Data sources

The platform launches with a selection of open sources and expands gradually.
A full list of all live sources, including licensing notes, lives on the
**Sources & Attribution** page.

<DataSourceCard name="Open-Meteo" url="https://open-meteo.com" license="CC-BY-4.0 (historical), public (live API)" region="EU-based, global" status="planned">First primary source for the research phase. Global weather data and model outputs without an API key, EU-operated infrastructure.</DataSourceCard>

<DataSourceCard name="Deutscher Wetterdienst (DWD) — OpenData" url="https://opendata.dwd.de" license="CC-BY-4.0 (German Geo-Reuse Regulation)" region="Germany, Central Europe" status="planned">MOSMIX forecasts, ICON model output, station observations, RADOLAN precipitation. Authoritative source for German weather data.</DataSourceCard>

## Methodology promise

- Sources are **transparently** attributed, and every view links back to the
  original.
- Processing and aggregation steps are **documented** and traceable in the
  open-source code (AGPL-3.0).
- Where sources disagree, discrepancies are **not hidden** but shown side by
  side.
- Climate indicators (anomalies, trends) ship with their **methodology and
  data basis**.

## Updates

This methodology page is updated with every significant step of the platform's
development. The `updated_at` date above reflects the date of the current
revision.
