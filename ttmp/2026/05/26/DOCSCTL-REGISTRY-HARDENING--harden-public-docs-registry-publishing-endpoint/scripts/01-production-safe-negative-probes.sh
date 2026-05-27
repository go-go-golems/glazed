#!/usr/bin/env bash
set -euo pipefail

# Production-safe negative probes for docs-registry.
#
# This script does not require or print secrets. It performs only unauthenticated
# and read-only checks against the public registry by default. The unauthenticated
# PUT sends a tiny invalid body and must be rejected before validation/storage.

REGISTRY_URL="${REGISTRY_URL:-https://docs-registry.yolo.scapegoat.dev}"
PACKAGE_NAME="${PACKAGE_NAME:-glazed}"
VERSION="${VERSION:-negative-proof-$(date -u +%Y%m%dT%H%M%SZ)}"
TMP_BODY=""

fail() {
  echo "FAIL: $*" >&2
  exit 1
}

expect_status() {
  local expected="$1"
  local method="$2"
  local url="$3"
  local body_file="${4:-}"

  local status
  if [[ -n "$body_file" ]]; then
    status=$(curl -sS -o /tmp/docs-registry-negative-body.json -w '%{http_code}' \
      -X "$method" --data-binary "@$body_file" "$url")
  else
    status=$(curl -sS -o /tmp/docs-registry-negative-body.json -w '%{http_code}' \
      -X "$method" "$url")
  fi

  if [[ "$status" != "$expected" ]]; then
    echo "Response body:" >&2
    cat /tmp/docs-registry-negative-body.json >&2 || true
    fail "$method $url returned $status, expected $expected"
  fi
  echo "OK: $method $url -> $status"
}

main() {
  TMP_BODY=$(mktemp)
  trap 'rm -f "$TMP_BODY" /tmp/docs-registry-negative-body.json' EXIT
  printf 'not a sqlite database\n' > "$TMP_BODY"

  expect_status 200 GET "$REGISTRY_URL/healthz"
  expect_status 401 PUT "$REGISTRY_URL/v1/packages/$PACKAGE_NAME/versions/$VERSION/sqlite" "$TMP_BODY"

  local metrics_status
  metrics_status=$(curl -sS -o /tmp/docs-registry-negative-body.json -w '%{http_code}' "$REGISTRY_URL/metrics" || true)
  if [[ "$metrics_status" == "200" ]]; then
    echo "OK: GET $REGISTRY_URL/metrics -> 200"
    if ! grep -q 'docs_registry_' /tmp/docs-registry-negative-body.json; then
      fail "metrics endpoint returned 200 but did not include docs_registry metrics"
    fi
  else
    echo "WARN: GET $REGISTRY_URL/metrics -> $metrics_status (acceptable if metrics are intentionally ingress-restricted)"
  fi
}

main "$@"
