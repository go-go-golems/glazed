# syntax=docker/dockerfile:1

FROM node:22-bookworm-slim AS web-builder
ENV CI=true \
  COREPACK_ENABLE_DOWNLOAD_PROMPT=0
WORKDIR /src/web
COPY web/package.json web/pnpm-lock.yaml ./
RUN corepack enable \
  && corepack prepare pnpm@10.15.0 --activate \
  && pnpm install --frozen-lockfile --reporter=append-only
COPY web/ ./
RUN pnpm build:all

FROM golang:1.26-bookworm AS builder
ENV CI=true \
  COREPACK_ENABLE_DOWNLOAD_PROMPT=0
WORKDIR /src

# Copy the full repository because the embedded React help browser is generated
# from web/ and pkg/web before compiling cmd/glaze.
COPY . .

# Build embedded web assets first. The generator prefers Dagger when available
# and falls back to local pnpm; the Node image provides the local fallback.
RUN apt-get update \
  && apt-get install -y --no-install-recommends nodejs npm ca-certificates \
  && npm install -g corepack@latest \
  && corepack enable \
  && go generate ./pkg/web \
  && CGO_ENABLED=1 GOOS=linux go build -tags embed -trimpath -ldflags='-s -w' -o /out/glaze ./cmd/glaze \
  && CGO_ENABLED=1 GOOS=linux go build -trimpath -ldflags='-s -w' -o /out/docsctl ./cmd/docsctl \
  && CGO_ENABLED=1 GOOS=linux go build -trimpath -ldflags='-s -w' -o /out/docs-registry ./cmd/docs-registry

FROM debian:bookworm-slim AS runtime
RUN apt-get update \
  && apt-get install -y --no-install-recommends ca-certificates \
  && rm -rf /var/lib/apt/lists/* \
  && useradd --system --uid 65532 --gid nogroup --home-dir /nonexistent --shell /usr/sbin/nologin nonroot
COPY --from=builder /out/glaze /usr/local/bin/glaze
COPY --from=builder /out/docsctl /usr/local/bin/docsctl
COPY --from=builder /out/docs-registry /usr/local/bin/docs-registry
USER nonroot:nogroup
EXPOSE 8088
ENTRYPOINT ["/usr/local/bin/glaze"]
CMD ["serve", "--address", ":8088"]

FROM node:22-bookworm-slim AS ssr
ENV NODE_ENV=production \
  SSR_PORT=8089 \
  COREPACK_ENABLE_DOWNLOAD_PROMPT=0
WORKDIR /app
COPY web/package.json web/pnpm-lock.yaml ./
RUN corepack enable \
  && corepack prepare pnpm@10.15.0 --activate \
  && pnpm install --prod --frozen-lockfile --reporter=append-only \
  && pnpm store prune
COPY web/server.mjs ./server.mjs
COPY --from=web-builder /src/web/dist ./dist
EXPOSE 8089
CMD ["node", "server.mjs"]
