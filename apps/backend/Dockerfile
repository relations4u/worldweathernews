# syntax=docker/dockerfile:1.7

# ─── Stage 1: builder ────────────────────────────────────────────────────────
FROM golang:1.26-alpine AS builder
ARG VERSION=dev
ARG COMMIT=unknown
ARG BUILD_DATE=unknown

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .

ENV CGO_ENABLED=0 GOOS=linux
RUN go build \
    -trimpath \
    -ldflags="-s -w \
      -X github.com/relations4u/worldweathernews/apps/backend/internal/version.Version=${VERSION} \
      -X github.com/relations4u/worldweathernews/apps/backend/internal/version.Commit=${COMMIT} \
      -X github.com/relations4u/worldweathernews/apps/backend/internal/version.BuildDate=${BUILD_DATE}" \
    -o /out/api ./cmd/api

# ─── Stage 2: dev (Hot-Reload mit air) ───────────────────────────────────────
FROM builder AS dev
RUN go install github.com/air-verse/air@v1.65.0
EXPOSE 8080
CMD ["air"]

# ─── Stage 3: final (distroless, nonroot) ────────────────────────────────────
FROM gcr.io/distroless/static-debian12:nonroot AS final
LABEL org.opencontainers.image.source="https://github.com/relations4u/worldweathernews"
LABEL org.opencontainers.image.title="wwn-backend"
LABEL org.opencontainers.image.description="worldweathernews.com API"

COPY --from=builder /out/api /api

USER nonroot:nonroot
EXPOSE 8080
HEALTHCHECK --interval=10s --timeout=3s --start-period=5s --retries=3 \
    CMD ["/api", "healthcheck"]
ENTRYPOINT ["/api"]
