"""OpenTelemetry-Tracing-Setup für die Worker-Services.

`init_tracing` setzt den globalen Tracer-Provider. Bei `enabled=False`
passiert nichts — die Pyworkers laufen dann ohne Tracing weiter.
`instrument_libraries` aktiviert Auto-Instrumentation für asyncpg, redis
und httpx; muss EINMAL beim Start aufgerufen werden.

`add_trace_context` ist ein structlog-Processor, der Trace- und Span-IDs
an jeden Log-Eintrag hängt — Loki ↔ Tempo verlinkt sich darüber.
"""

from __future__ import annotations

from typing import Any

from opentelemetry import trace
from opentelemetry.exporter.otlp.proto.grpc.trace_exporter import OTLPSpanExporter
from opentelemetry.instrumentation.asyncpg import AsyncPGInstrumentor
from opentelemetry.instrumentation.httpx import HTTPXClientInstrumentor
from opentelemetry.instrumentation.redis import RedisInstrumentor
from opentelemetry.sdk.resources import Resource
from opentelemetry.sdk.trace import TracerProvider
from opentelemetry.sdk.trace.export import BatchSpanProcessor
from opentelemetry.semconv.resource import ResourceAttributes


def init_tracing(
    *,
    enabled: bool,
    endpoint: str,
    service_name: str,
    environment: str,
    version: str,
) -> None:
    """Initialisiert den OTLP-Trace-Exporter und setzt den globalen TracerProvider.

    Bei ``enabled=False`` no-op, damit Tests und CLI ohne Tempo laufen.
    """
    if not enabled:
        return

    resource = Resource.create(
        {
            ResourceAttributes.SERVICE_NAME: service_name,
            ResourceAttributes.SERVICE_VERSION: version,
            ResourceAttributes.DEPLOYMENT_ENVIRONMENT: environment,
        }
    )

    provider = TracerProvider(resource=resource)
    exporter = OTLPSpanExporter(endpoint=endpoint, insecure=True)
    provider.add_span_processor(BatchSpanProcessor(exporter))

    trace.set_tracer_provider(provider)


def instrument_libraries() -> None:
    """Aktiviert OTel-Auto-Instrumentation für asyncpg, redis und httpx.

    Idempotent — die Instrumentor-Klassen sind Singletons und prüfen selbst,
    ob bereits instrumentiert wurde.
    """
    AsyncPGInstrumentor().instrument()
    RedisInstrumentor().instrument()
    HTTPXClientInstrumentor().instrument()


def add_trace_context(
    _logger: Any,
    _method_name: str,
    event_dict: dict[str, Any],
) -> dict[str, Any]:
    """Structlog-Processor — fügt trace_id/span_id ein, wenn ein Span aktiv ist."""
    span = trace.get_current_span()
    ctx = span.get_span_context()
    if ctx.is_valid:
        event_dict["trace_id"] = format(ctx.trace_id, "032x")
        event_dict["span_id"] = format(ctx.span_id, "016x")
    return event_dict
