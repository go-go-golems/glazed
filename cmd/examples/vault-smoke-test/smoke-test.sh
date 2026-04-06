#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR=$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)
REPO_ROOT=$(cd -- "$SCRIPT_DIR/../../.." && pwd)
SESSION_NAME="glazed-vault-smoke-$$"
TMP_DIR=$(mktemp -d)
CONFIG_FILE="$TMP_DIR/config.yaml"
VAULT_PORT="${GLAZED_VAULT_SMOKE_TEST_PORT:-18200}"
VAULT_ADDR="http://127.0.0.1:${VAULT_PORT}"
ROOT_TOKEN="root"
EXAMPLE_PKG="./cmd/examples/vault-smoke-test"
SECRET_PATH="secret/glazed-demo"

cleanup() {
  local exit_code=$?
  if tmux has-session -t "$SESSION_NAME" 2>/dev/null; then
    if [[ $exit_code -ne 0 ]]; then
      echo
      echo "Vault server output from tmux session $SESSION_NAME:" >&2
      tmux capture-pane -pt "$SESSION_NAME" >&2 || true
    fi
    tmux kill-session -t "$SESSION_NAME" >/dev/null 2>&1 || true
  fi
  rm -rf "$TMP_DIR"
}
trap cleanup EXIT

require_command() {
  local name=$1
  if ! command -v "$name" >/dev/null 2>&1; then
    echo "missing required command: $name" >&2
    exit 1
  fi
}

assert_contains() {
  local haystack=$1
  local needle=$2
  if ! grep -Fq -- "$needle" <<<"$haystack"; then
    echo "expected output to contain: $needle" >&2
    echo
    echo "$haystack" >&2
    exit 1
  fi
}

assert_not_contains() {
  local haystack=$1
  local needle=$2
  if grep -Fq -- "$needle" <<<"$haystack"; then
    echo "expected output not to contain: $needle" >&2
    echo
    echo "$haystack" >&2
    exit 1
  fi
}

run_example() {
  (
    cd "$REPO_ROOT"
    "$@"
  )
}

require_command go
require_command tmux
require_command vault

cat > "$CONFIG_FILE" <<YAML
vault-settings:
  secret-path: ${SECRET_PATH}
app:
  host: from-config-host
  password: from-config-password
YAML

echo "Starting Vault dev server in tmux session $SESSION_NAME"
tmux new-session -d -s "$SESSION_NAME" "vault server -dev -dev-root-token-id ${ROOT_TOKEN} -dev-listen-address 127.0.0.1:${VAULT_PORT}"

for _ in $(seq 1 30); do
  if VAULT_ADDR="$VAULT_ADDR" VAULT_TOKEN="$ROOT_TOKEN" vault status >/dev/null 2>&1; then
    break
  fi
  sleep 1
done

if ! VAULT_ADDR="$VAULT_ADDR" VAULT_TOKEN="$ROOT_TOKEN" vault status >/dev/null 2>&1; then
  echo "Vault dev server did not become ready" >&2
  exit 1
fi

echo "Seeding Vault secrets at $SECRET_PATH"
VAULT_ADDR="$VAULT_ADDR" VAULT_TOKEN="$ROOT_TOKEN" vault kv put "$SECRET_PATH" \
  password=from-vault-password \
  api-key=from-vault-api-key \
  host=from-vault-host >/dev/null

echo "Case 1: config below Vault, non-secret host remains from config"
case1=$(run_example env \
  VAULT_ADDR="$VAULT_ADDR" \
  VAULT_TOKEN="$ROOT_TOKEN" \
  GLAZED_VAULT_SMOKE_TEST_VAULT_ADDR="$VAULT_ADDR" \
  go run "$EXAMPLE_PKG" --config-file "$CONFIG_FILE")
printf '%s\n' "$case1"
assert_contains "$case1" "secret_path=$SECRET_PATH"
assert_contains "$case1" "host=from-config-host"
assert_contains "$case1" "host_source=config"
assert_contains "$case1" "password=***"
assert_contains "$case1" "password_source=vault"
assert_contains "$case1" "api_key=***"
assert_contains "$case1" "api_key_source=vault"
assert_not_contains "$case1" "host=from-vault-host"

echo "Case 2: env overrides Vault-backed secret"
case2=$(run_example env \
  VAULT_ADDR="$VAULT_ADDR" \
  VAULT_TOKEN="$ROOT_TOKEN" \
  GLAZED_VAULT_SMOKE_TEST_VAULT_ADDR="$VAULT_ADDR" \
  GLAZED_VAULT_SMOKE_TEST_PASSWORD=from-env-password \
  go run "$EXAMPLE_PKG" --config-file "$CONFIG_FILE")
printf '%s\n' "$case2"
assert_contains "$case2" "password=***"
assert_contains "$case2" "password_source=env"

echo "Case 3: flag overrides env"
case3=$(run_example env \
  VAULT_ADDR="$VAULT_ADDR" \
  VAULT_TOKEN="$ROOT_TOKEN" \
  GLAZED_VAULT_SMOKE_TEST_VAULT_ADDR="$VAULT_ADDR" \
  GLAZED_VAULT_SMOKE_TEST_PASSWORD=from-env-password \
  go run "$EXAMPLE_PKG" --config-file "$CONFIG_FILE" --password from-flag-password)
printf '%s\n' "$case3"
assert_contains "$case3" "password=***"
assert_contains "$case3" "password_source=cobra"

echo "Case 4: bootstrap from env without config file"
case4=$(run_example env \
  VAULT_ADDR="$VAULT_ADDR" \
  VAULT_TOKEN="$ROOT_TOKEN" \
  GLAZED_VAULT_SMOKE_TEST_VAULT_ADDR="$VAULT_ADDR" \
  GLAZED_VAULT_SMOKE_TEST_SECRET_PATH="$SECRET_PATH" \
  go run "$EXAMPLE_PKG")
printf '%s\n' "$case4"
assert_contains "$case4" "secret_path=$SECRET_PATH"
assert_contains "$case4" "password=***"
assert_contains "$case4" "api_key=***"

echo "Case 5: parsed-field output redacts secrets"
case5=$(run_example env \
  VAULT_ADDR="$VAULT_ADDR" \
  VAULT_TOKEN="$ROOT_TOKEN" \
  GLAZED_VAULT_SMOKE_TEST_VAULT_ADDR="$VAULT_ADDR" \
  go run "$EXAMPLE_PKG" --config-file "$CONFIG_FILE" --print-parsed-fields)
printf '%s\n' "$case5"
assert_contains "$case5" "password:"
assert_contains "$case5" "api-key:"
assert_contains "$case5" "source: vault"
assert_contains "$case5" "***"
assert_not_contains "$case5" "from-vault-password"
assert_not_contains "$case5" "from-vault-api-key"
assert_not_contains "$case5" "root"

echo

echo "Vault smoke test passed"
