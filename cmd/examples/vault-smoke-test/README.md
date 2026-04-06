# Vault Smoke Test Example

This example is the real end-to-end harness for the Vault support added in GL-009.

It proves five things:

- `TypeSecret` fields can hydrate from Vault
- non-secret fields do not hydrate from Vault
- config can provide the bootstrap `vault-settings`
- environment variables still override Vault-backed application secrets
- flags still override environment variables, and `--print-parsed-fields` stays redacted

## Fields

Application fields:

- `host` is a normal string field
- `password` is a `TypeSecret` field
- `api-key` is a `TypeSecret` field

Vault settings:

- `--vault-addr`
- `--vault-token`
- `--vault-token-source`
- `--vault-token-file`
- `--secret-path`

General command settings:

- `--config-file`
- `--print-parsed-fields`

## Environment Variables

The example uses the prefix `GLAZED_VAULT_SMOKE_TEST`.

Useful variables:

- `GLAZED_VAULT_SMOKE_TEST_HOST`
- `GLAZED_VAULT_SMOKE_TEST_PASSWORD`
- `GLAZED_VAULT_SMOKE_TEST_API_KEY`
- `GLAZED_VAULT_SMOKE_TEST_VAULT_ADDR`
- `GLAZED_VAULT_SMOKE_TEST_VAULT_TOKEN`
- `GLAZED_VAULT_SMOKE_TEST_SECRET_PATH`

## Expected Precedence

The example intentionally uses this chain:

`defaults -> config -> vault -> env -> args -> cobra`

That means:

- config can provide `vault-settings.secret-path`
- Vault can fill in `TypeSecret` fields after config is loaded
- environment variables can still override the Vault-backed app secrets
- flags still win last

## Manual Run

Start a local Vault dev server in another terminal or `tmux` session:

```bash
vault server -dev -dev-root-token-id root
```

Set the connection variables:

```bash
export VAULT_ADDR=http://127.0.0.1:8200
export VAULT_TOKEN=root
```

Seed a secret path:

```bash
vault kv put kv/glazed-demo \
  password=from-vault-password \
  api-key=from-vault-api-key \
  host=from-vault-host
```

Write a config file:

```yaml
vault-settings:
  secret-path: kv/glazed-demo
app:
  host: from-config-host
  password: from-config-password
```

Run the example:

```bash
go run ./cmd/examples/vault-smoke-test --config-file ./config.yaml
```

Expected output shape:

```text
env_prefix=GLAZED_VAULT_SMOKE_TEST
vault_addr=http://127.0.0.1:8200
secret_path=kv/glazed-demo
host=from-config-host
host_source=config
password=***
password_source=vault
api_key=***
api_key_source=vault
```

Redaction check:

```bash
go run ./cmd/examples/vault-smoke-test --config-file ./config.yaml --print-parsed-fields
```

The example masks secret values in its normal `key=value` output, and the parsed-field dump should also show `***` instead of the raw password or API key.

## Smoke Script

Use `./smoke-test.sh` to run the full automated local check. The script starts Vault in `tmux`, seeds test data, runs the example in several precedence modes, and asserts the expected output.
