# syntax=docker/dockerfile:1.7

# ─── Stage 1: builder ────────────────────────────────────────────────────────
FROM python:3.12-slim AS builder
ARG VERSION=dev
ARG COMMIT=unknown

# uv aus dem offiziellen astral Image
COPY --from=ghcr.io/astral-sh/uv:latest /uv /usr/local/bin/uv

WORKDIR /app
COPY pyproject.toml uv.lock ./
COPY pyworkers ./pyworkers

ENV UV_COMPILE_BYTECODE=1 \
    UV_LINK_MODE=copy \
    UV_PROJECT_ENVIRONMENT=/opt/venv

RUN uv sync --frozen --no-dev


# ─── Stage 2: runtime ────────────────────────────────────────────────────────
FROM python:3.12-slim
LABEL org.opencontainers.image.source="https://github.com/relations4u/worldweathernews"
LABEL org.opencontainers.image.title="wwn-pyworkers"
LABEL org.opencontainers.image.description="worldweathernews.com Worker-Service"

ARG VERSION=dev
ARG COMMIT=unknown
ARG BUILD_DATE=unknown

ENV PYTHONDONTWRITEBYTECODE=1 \
    PYTHONUNBUFFERED=1 \
    PATH="/opt/venv/bin:$PATH" \
    WWN_PY_VERSION=$VERSION \
    WWN_PY_COMMIT=$COMMIT \
    WWN_PY_BUILD_DATE=$BUILD_DATE

# Non-root User
RUN groupadd -r app && useradd -r -g app -d /home/app -s /sbin/nologin app && \
    mkdir -p /home/app && chown -R app:app /home/app

WORKDIR /app
COPY --from=builder /opt/venv /opt/venv
COPY --from=builder /app/pyworkers /app/pyworkers

USER app
EXPOSE 9100

# prometheus_client serviert auf "/" Metriken zurück.
HEALTHCHECK --interval=10s --timeout=3s --start-period=10s --retries=3 \
    CMD python -c "import urllib.request; urllib.request.urlopen('http://127.0.0.1:9100/').read()" || exit 1

ENTRYPOINT ["python", "-m", "pyworkers"]
