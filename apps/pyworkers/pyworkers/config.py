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
