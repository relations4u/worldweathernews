# syntax=docker/dockerfile:1.7

# ─── Stage 1: builder ────────────────────────────────────────────────────────
FROM golang:1.25.10-alpine AS builder
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
    -ldflags="-s -w" \
    -o /out/wwn-cms-auth ./cmd/cms-auth

# ─── Stage 2: final (distroless, nonroot) ────────────────────────────────────
FROM gcr.io/distroless/static-debian12:nonroot AS final
LABEL org.opencontainers.image.source="https://github.com/relations4u/worldweathernews"
LABEL org.opencontainers.image.title="wwn-cms-auth"
LABEL org.opencontainers.image.description="Self-hosted GitHub OAuth proxy for Sveltia CMS"

COPY --from=builder /out/wwn-cms-auth /wwn-cms-auth

USER nonroot:nonroot
EXPOSE 8090
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD ["/wwn-cms-auth", "healthcheck"]
ENTRYPOINT ["/wwn-cms-auth"]
