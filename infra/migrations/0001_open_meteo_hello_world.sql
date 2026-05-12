-- +goose Up

CREATE TABLE locations (
    id          BIGSERIAL PRIMARY KEY,
    slug        TEXT NOT NULL UNIQUE,
    name        TEXT NOT NULL,
    country     TEXT NOT NULL DEFAULT 'DE',
    latitude    DOUBLE PRECISION NOT NULL,
    longitude   DOUBLE PRECISION NOT NULL,
    timezone    TEXT NOT NULL DEFAULT 'Europe/Berlin',
    source      TEXT NOT NULL,
    active      BOOLEAN NOT NULL DEFAULT TRUE,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

INSERT INTO locations (slug, name, latitude, longitude, source) VALUES
    ('potsdam', 'Potsdam', 52.3906, 13.0645, 'open-meteo'),
    ('berlin',  'Berlin',  52.5200, 13.4050, 'open-meteo'),
    ('hamburg', 'Hamburg', 53.5511,  9.9937, 'open-meteo');

CREATE TABLE observations (
    location_id    BIGINT NOT NULL REFERENCES locations(id) ON DELETE CASCADE,
    observed_at    TIMESTAMPTZ NOT NULL,
    temperature    DOUBLE PRECISION,
    precipitation  DOUBLE PRECISION,
    wind_speed     DOUBLE PRECISION,
    wind_direction INTEGER,
    source         TEXT NOT NULL,
    fetched_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (location_id, observed_at)
);

SELECT create_hypertable('observations', by_range('observed_at'));

CREATE INDEX idx_observations_location_observed
    ON observations (location_id, observed_at DESC);

CREATE TABLE forecasts (
    location_id    BIGINT NOT NULL REFERENCES locations(id) ON DELETE CASCADE,
    forecast_for   TIMESTAMPTZ NOT NULL,
    run_at         TIMESTAMPTZ NOT NULL,
    temperature    DOUBLE PRECISION,
    precipitation  DOUBLE PRECISION,
    wind_speed     DOUBLE PRECISION,
    wind_direction INTEGER,
    source         TEXT NOT NULL,
    fetched_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (location_id, forecast_for, run_at)
);

SELECT create_hypertable('forecasts', by_range('forecast_for'));

CREATE INDEX idx_forecasts_location_forecast_for
    ON forecasts (location_id, forecast_for DESC);

-- +goose Down

DROP TABLE IF EXISTS forecasts;
DROP TABLE IF EXISTS observations;
DROP TABLE IF EXISTS locations;
