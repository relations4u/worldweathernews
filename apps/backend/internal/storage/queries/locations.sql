-- name: ListActiveLocations :many
SELECT id, slug, name, country, latitude, longitude, timezone, source
FROM locations
WHERE active = TRUE
ORDER BY name;

-- name: GetLocationBySlug :one
SELECT id, slug, name, country, latitude, longitude, timezone, source
FROM locations
WHERE slug = $1 AND active = TRUE;
