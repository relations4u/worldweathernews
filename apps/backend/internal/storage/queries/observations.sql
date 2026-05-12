-- name: GetLatestObservation :one
SELECT observed_at, temperature, precipitation, wind_speed, wind_direction,
       pressure, humidity, source, fetched_at
FROM observations
WHERE location_id = $1
ORDER BY observed_at DESC
LIMIT 1;

-- name: GetLatestObservationBySource :one
SELECT observed_at, temperature, precipitation, wind_speed, wind_direction,
       pressure, humidity, source, fetched_at
FROM observations
WHERE location_id = $1 AND source = $2
ORDER BY observed_at DESC
LIMIT 1;
