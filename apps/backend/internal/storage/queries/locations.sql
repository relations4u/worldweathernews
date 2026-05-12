-- name: ListActiveLocations :many
SELECT
    l.id, l.slug, l.name, l.country, l.latitude, l.longitude,
    l.timezone, l.source, l.dwd_station_id, l.altitude_m,
    COALESCE(
        (SELECT array_agg(DISTINCT o.source ORDER BY o.source)
         FROM observations o WHERE o.location_id = l.id),
        ARRAY[]::TEXT[]
    )::TEXT[] AS available_sources
FROM locations l
WHERE l.active = TRUE
ORDER BY l.name;

-- name: GetLocationBySlug :one
SELECT
    l.id, l.slug, l.name, l.country, l.latitude, l.longitude,
    l.timezone, l.source, l.dwd_station_id, l.altitude_m,
    COALESCE(
        (SELECT array_agg(DISTINCT o.source ORDER BY o.source)
         FROM observations o WHERE o.location_id = l.id),
        ARRAY[]::TEXT[]
    )::TEXT[] AS available_sources
FROM locations l
WHERE l.slug = $1 AND l.active = TRUE;
