# syntax=docker/dockerfile:1

FROM golang:1.25-bookworm AS builder
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
  && CGO_ENABLED=1 GOOS=linux go build -trimpath -ldflags='-s -w' -o /out/glaze ./cmd/glaze \
  && CGO_ENABLED=1 GOOS=linux go build -trimpath -ldflags='-s -w' -o /out/docsctl ./cmd/docsctl \
  && CGO_ENABLED=1 GOOS=linux go build -trimpath -ldflags='-s -w' -o /out/docs-registry ./cmd/docs-registry

FROM debian:bookworm-slim
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
