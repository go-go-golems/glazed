#!/usr/bin/env bash
set -euo pipefail

# Ensure we run from the glazed module root, regardless of where this script lives
SCRIPT_DIR="$(dirname "${BASH_SOURCE[0]}")"
cd "${SCRIPT_DIR}/../.."

echo "=== PWD: $(pwd) ==="

run() {
  echo "\n--- $*" >&2
  bash -lc "$*"
}

backup() {
  local f="$1"
  cp -f "$f" "$f.bak"
}

restore() {
  local f="$1"
  mv -f "$f.bak" "$f"
}

echo "=== config-single: valid ==="
run "go run ./cmd/examples/config-single validate"

echo "=== config-single: invalid unknown layer ==="
backup cmd/examples/config-single/config.yaml
cat > cmd/examples/config-single/config.yaml << 'YAML'
other:
  foo: bar
YAML
run "go run ./cmd/examples/config-single validate || true"
restore cmd/examples/config-single/config.yaml

echo "=== config-single: invalid unknown parameter ==="
backup cmd/examples/config-single/config.yaml
cat > cmd/examples/config-single/config.yaml << 'YAML'
demo:
  extra: 1
YAML
run "go run ./cmd/examples/config-single validate || true"
restore cmd/examples/config-single/config.yaml

echo "=== config-single: invalid type mismatch ==="
backup cmd/examples/config-single/config.yaml
cat > cmd/examples/config-single/config.yaml << 'YAML'
demo:
  api-key: cfg-one
  threshold: "oops"
YAML
run "go run ./cmd/examples/config-single validate || true"
restore cmd/examples/config-single/config.yaml

echo "=== config-overlay: valid ==="
run "go run ./cmd/examples/config-overlay validate"

echo "=== config-overlay: invalid base unknown layer ==="
backup cmd/examples/config-overlay/base.yaml
cat > cmd/examples/config-overlay/base.yaml << 'YAML'
other:
  foo: bar
YAML
run "go run ./cmd/examples/config-overlay validate || true"
restore cmd/examples/config-overlay/base.yaml

echo "=== config-overlay: invalid env unknown parameter ==="
backup cmd/examples/config-overlay/env.yaml
cat > cmd/examples/config-overlay/env.yaml << 'YAML'
demo:
  extra: 1
YAML
run "go run ./cmd/examples/config-overlay validate || true"
restore cmd/examples/config-overlay/env.yaml

echo "=== config-overlay: invalid local type mismatch ==="
backup cmd/examples/config-overlay/local.yaml
cat > cmd/examples/config-overlay/local.yaml << 'YAML'
demo:
  api-key: local
  threshold: "oops"
YAML
run "go run ./cmd/examples/config-overlay validate || true"
restore cmd/examples/config-overlay/local.yaml

echo "=== config-pattern-mapper: valid ==="
run "go run ./cmd/examples/config-pattern-mapper validate ./cmd/examples/config-pattern-mapper/config-example.yaml"

echo "=== config-pattern-mapper: invalid unknown dynamic parameter (staging) ==="
cat > cmd/examples/config-pattern-mapper/config-invalid-staging.yaml << 'YAML'
environments:
  staging:
    settings:
      api_key: "staging-secret"
YAML
run "go run ./cmd/examples/config-pattern-mapper validate ./cmd/examples/config-pattern-mapper/config-invalid-staging.yaml || true"

echo "=== config-custom-mapper: valid ==="
run "go run ./cmd/examples/config-custom-mapper validate"

echo "=== config-custom-mapper: invalid type mismatch ==="
backup cmd/examples/config-custom-mapper/config.yaml
cat > cmd/examples/config-custom-mapper/config.yaml << 'YAML'
api_key: "secret-from-flat-config"
threshold: "oops"
YAML
run "go run ./cmd/examples/config-custom-mapper validate || true"
restore cmd/examples/config-custom-mapper/config.yaml

echo "\nAll validation runs completed."


