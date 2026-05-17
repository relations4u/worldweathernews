"""Konfiguration via pydantic-settings — ENV-Prefix `WWN_PY_`."""

from pydantic import Field, PostgresDsn, RedisDsn, field_validator
from pydantic_settings import BaseSettings, SettingsConfigDict


class Settings(BaseSettings):
    """Backend-übergreifende Config der Worker-Services."""

    model_config = SettingsConfigDict(
        env_prefix="WWN_PY_",
        env_file=".env",
        env_file_encoding="utf-8",
        extra="ignore",
        case_sensitive=False,
    )

    # General
    environment: str = Field(default="production", pattern="^(dev|staging|production)$")
    log_level: str = Field(default="INFO", pattern="^(DEBUG|INFO|WARNING|ERROR)$")
    log_format: str = Field(default="json", pattern="^(json|text)$")

    # Storage
    database_url: PostgresDsn
    redis_url: RedisDsn

    # Metrics
    metrics_enabled: bool = True
    metrics_port: int = Field(default=9100, ge=1, le=65535)

    # Heartbeat
    heartbeat_interval_seconds: int = Field(default=30, ge=1)

    # Open-Meteo Worker
    open_meteo_enabled: bool = True
    open_meteo_current_interval_seconds: int = Field(default=600, ge=60)
    open_meteo_hourly_interval_seconds: int = Field(default=3600, ge=60)

    # DWD POI Worker (deutsche Stationsbeobachtungen, halbstündlich)
    dwd_enabled: bool = True
    dwd_poi_interval_seconds: int = Field(default=1800, ge=60)

    # EUMETSAT-Satelliten-Worker (EUMETView WMS, IR 10.8, Europa).
    # Pfad A: server-seitig ziehen, in den A.13-Bucket ablegen, das
    # Frontend lädt nur über media.worldweathernews.com. Kein Auth für
    # die WMS (Q4 verifiziert) — der eumetsat.env-Secret bleibt für den
    # K1-Pfad (~2.6) und wird hier NICHT gebraucht.
    eumetsat_enabled: bool = True
    eumetsat_interval_seconds: int = Field(default=900, ge=60)  # 15 min
    eumetsat_window_hours: int = Field(default=24, ge=1)

    # S3-Ziel (A.13 Hetzner Object Storage). Credentials werden beim
    # Deploy aus dem media-storage-SOPS-File als WWN_PY_S3_* injiziert
    # (nicht im Repo). Leer lassen ⇒ der Job loggt + skippt sauber.
    s3_endpoint: str = ""
    s3_region: str = "fsn1"
    s3_bucket: str = "media-worldweathernews-prod"
    s3_access_key_id: str = ""
    s3_secret_access_key: str = ""
    # Bucket-Prefix für die Satellitenframes (muss in der Bucket-Policy
    # public-read sein — siehe infra/object-storage/bucket-policy.json).
    sat_prefix: str = "sat/ir108"
    # Basis-URL, unter der die Frames öffentlich ausgeliefert werden.
    media_base_url: str = "https://media.worldweathernews.com"

    # Tracing (OTLP gRPC → Tempo). Endpoint ist host:port ohne Schema.
    tracing_enabled: bool = False
    tracing_endpoint: str = "tempo:4317"

    @field_validator("log_level", mode="before")
    @classmethod
    def _upper_log_level(cls, v: object) -> object:
        # ENV-Werte wie "debug" tolerieren — slog/structlog erwartet uppercase.
        return v.upper() if isinstance(v, str) else v


def load_settings() -> Settings:
    # database_url + redis_url sind Pflichtfelder, kommen aus ENV.
    return Settings()
