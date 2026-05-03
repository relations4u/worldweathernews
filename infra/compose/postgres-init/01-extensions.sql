-- Wird beim allerersten Start ausgeführt, wenn das DB-Verzeichnis leer ist.
-- Bei späteren Starts ohne Wirkung (Postgres-Init-Mechanik).
--
-- Reset über `make dev-reset` (löscht das postgres_data-Volume).

CREATE EXTENSION IF NOT EXISTS postgis;
CREATE EXTENSION IF NOT EXISTS timescaledb;
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS pg_trgm;
