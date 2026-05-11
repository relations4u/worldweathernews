# syntax=docker/dockerfile:1.7

# ─── Stage 1: deps ───────────────────────────────────────────────────────────
FROM node:26-alpine AS deps
WORKDIR /app
COPY package.json pnpm-lock.yaml ./
RUN corepack enable && pnpm install --frozen-lockfile

# ─── Stage 2: builder ────────────────────────────────────────────────────────
FROM node:26-alpine AS builder
ARG VERSION=dev
ARG COMMIT=unknown
ARG PUBLIC_API_BASE_URL=http://api.localhost
ENV PUBLIC_API_BASE_URL=$PUBLIC_API_BASE_URL

WORKDIR /app
COPY --from=deps /app/node_modules ./node_modules
COPY . .
RUN corepack enable && pnpm build

# ─── Stage 3: runner ─────────────────────────────────────────────────────────
FROM node:26-alpine AS runner
LABEL org.opencontainers.image.source="https://github.com/relations4u/worldweathernews"
LABEL org.opencontainers.image.title="wwn-frontend"
LABEL org.opencontainers.image.description="worldweathernews.com Frontend"

WORKDIR /app
RUN addgroup -S app && adduser -S -G app app

COPY --from=builder /app/build ./build
COPY --from=builder /app/package.json ./package.json
COPY --from=deps /app/node_modules ./node_modules

USER app
EXPOSE 3000
ENV NODE_ENV=production
ENV HOST=0.0.0.0
ENV PORT=3000

# 127.0.0.1 statt localhost: busybox-wget resolved `localhost` zu ::1, der
# Node-Server bindet aber nur IPv4 (HOST=0.0.0.0), und es kommt zu „Connection
# refused" obwohl der Server läuft → Container wäre permanent unhealthy.
HEALTHCHECK --interval=10s --timeout=3s --start-period=5s --retries=3 \
    CMD wget -q --spider http://127.0.0.1:3000/ || exit 1

CMD ["node", "build"]
