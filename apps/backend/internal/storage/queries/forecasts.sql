-- name: GetForecastNext24h :many
SELECT f.forecast_for, f.temperature, f.precipitation, f.wind_speed,
       f.wind_direction, f.run_at
FROM forecasts f
WHERE f.location_id = $1
  AND f.forecast_for > NOW()
  AND f.forecast_for <= NOW() + INTERVAL '24 hours'
  AND f.run_at = (SELECT MAX(f2.run_at)
                  FROM forecasts f2
                  WHERE f2.location_id = $1)
ORDER BY f.forecast_for;
