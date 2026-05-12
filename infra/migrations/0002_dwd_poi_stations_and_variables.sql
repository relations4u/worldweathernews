-- +goose Up

-- Erweitere observations-PK um source, damit DWD und Open-Meteo parallel
-- für dieselbe (location_id, observed_at) speichern können, ohne sich zu
-- überschreiben. observed_at bleibt drin, deshalb akzeptiert TimescaleDB
-- die neue PK auf der Hypertable.
ALTER TABLE observations DROP CONSTRAINT observations_pkey;
ALTER TABLE observations ADD PRIMARY KEY (location_id, source, observed_at);

-- locations: dwd_station_id (Schlüssel für POI-CSV-Fetch) + altitude_m
-- (Frontend zeigt Höhe für Klimakontrast-Stationen).
ALTER TABLE locations
    ADD COLUMN dwd_station_id TEXT,
    ADD COLUMN altitude_m     INTEGER;

CREATE INDEX idx_locations_dwd_station_id
    ON locations (dwd_station_id)
    WHERE dwd_station_id IS NOT NULL;

-- observations: pressure (hPa MSL) + humidity (% relative).
ALTER TABLE observations
    ADD COLUMN pressure DOUBLE PRECISION,
    ADD COLUMN humidity DOUBLE PRECISION;

-- Bestehende Stadt-Locations: DWD-Station-ID + Höhe nachtragen.
-- locations.source bleibt 'open-meteo' (Legacy-Wert aus 2.1); die API
-- leitet availableSources aus den tatsächlich gespeicherten
-- observations-Rows ab, nicht aus locations.source.
--
-- IDs sind 5-stellige WMO-Synop-Kennungen (die Plan-Skizze listete
-- DWD-Legacy/CDC-Kennungen, die der POI-Endpoint nicht kennt — POI
-- adressiert ausschließlich über WMO-Synop). Verifiziert gegen
-- https://opendata.dwd.de/weather/weather_reports/poi/.
UPDATE locations SET dwd_station_id = '10379', altitude_m = 81 WHERE slug = 'potsdam';
UPDATE locations SET dwd_station_id = '10384', altitude_m = 48 WHERE slug = 'berlin';
UPDATE locations SET dwd_station_id = '10147', altitude_m = 11 WHERE slug = 'hamburg';

-- Drei neue Klimakontrast-Locations (DWD-only).
INSERT INTO locations
    (slug, name, country, latitude, longitude, timezone, source, dwd_station_id, altitude_m)
VALUES
    ('brocken',   'Brocken',   'DE', 51.7991, 10.6178, 'Europe/Berlin', 'dwd', '10454', 1134),
    ('zugspitze', 'Zugspitze', 'DE', 47.4209, 10.9854, 'Europe/Berlin', 'dwd', '10961', 2964),
    ('helgoland', 'Helgoland', 'DE', 54.1827,  7.8868, 'Europe/Berlin', 'dwd', '10015',    4);

-- +goose Down

DELETE FROM locations WHERE slug IN ('brocken', 'zugspitze', 'helgoland');

UPDATE locations
   SET dwd_station_id = NULL, altitude_m = NULL
 WHERE slug IN ('potsdam', 'berlin', 'hamburg');

ALTER TABLE observations
    DROP COLUMN humidity,
    DROP COLUMN pressure;

DROP INDEX IF EXISTS idx_locations_dwd_station_id;

ALTER TABLE locations
    DROP COLUMN altitude_m,
    DROP COLUMN dwd_station_id;

ALTER TABLE observations DROP CONSTRAINT observations_pkey;
ALTER TABLE observations ADD PRIMARY KEY (location_id, observed_at);
