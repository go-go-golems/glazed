#!/usr/bin/env bash
set -euo pipefail

# Smoke test for Option A profile bootstrap parsing.
#
# Validates that Geppetto's simple-inference example:
# - loads profile selection from env (PINOCCHIO_PROFILE)
# - can override profile file via env (PINOCCHIO_PROFILE_FILE)
# - fails for unknown profiles (PINOCCHIO_PROFILE=foobar)
#
# Notes:
# - Default profile file (when not overridden) is usually:
#     ~/.config/pinocchio/profiles.yaml
#   (or $XDG_CONFIG_HOME/pinocchio/profiles.yaml if XDG_CONFIG_HOME is set)
# - This script does NOT create or modify your ~/.config by default.

usage() {
  cat <<'USAGE'
Usage:
  01-smoke-test-simple-inference-profiles.sh [--profile-file <path>] [--repo-root <path>]

Options:
  --profile-file <path>  Path to profiles.yaml to use for the run via PINOCCHIO_PROFILE_FILE.
                         Default: <repo-root>/geppetto/misc/profiles.yaml
  --repo-root <path>     Repo root override. Default: `git rev-parse --show-toplevel`

Examples:
  ./01-smoke-test-simple-inference-profiles.sh
  ./01-smoke-test-simple-inference-profiles.sh --profile-file ~/.config/pinocchio/profiles.yaml
USAGE
}

repo_root=""
profile_file=""

while [[ $# -gt 0 ]]; do
  case "$1" in
    --repo-root)
      repo_root="${2:-}"; shift 2;;
    --profile-file)
      profile_file="${2:-}"; shift 2;;
    -h|--help)
      usage; exit 0;;
    *)
      echo "Unknown arg: $1" >&2
      usage
      exit 2;;
  esac
done

if [[ -z "$repo_root" ]]; then
  # Don't require git. This workspace may not be a git repo at the top-level.
  # Instead, walk up from this script's directory until we find a directory
  # that contains both `geppetto/` and `glazed/` (workspace root in this mono-repo).
  script_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
  candidate="$script_dir"
  while [[ "$candidate" != "/" ]]; do
    if [[ -d "$candidate/geppetto" && -d "$candidate/glazed" ]]; then
      repo_root="$candidate"
      break
    fi
    candidate="$(dirname "$candidate")"
  done
  if [[ -z "$repo_root" ]]; then
    echo "ERROR: could not auto-detect repo root (no git, and couldn't find geppetto/ + glazed/ by walking up)" >&2
    echo "Provide --repo-root /abs/path/to/workspace-root" >&2
    exit 2
  fi
fi

geppetto_dir="$repo_root/geppetto"
example_dir="$geppetto_dir/cmd/examples/simple-inference"

if [[ -z "$profile_file" ]]; then
  profile_file="$geppetto_dir/misc/profiles.yaml"
fi

if [[ ! -d "$geppetto_dir" ]]; then
  echo "ERROR: geppetto dir not found: $geppetto_dir" >&2
  exit 2
fi

if [[ ! -f "$profile_file" ]]; then
  echo "ERROR: profile file not found: $profile_file" >&2
  echo "Default location (if not overriding) would be: \$HOME/.config/pinocchio/profiles.yaml" >&2
  exit 2
fi

extract_yaml_from_output() {
  # Strip any leading log lines and keep the YAML starting at the first known top-level key.
  # For these tests we expect the print-parsed-parameters output to include ai-chat as the first key.
  sed -n '/^ai-chat:/,$p'
}

get_profile_value() {
  # Reads YAML on stdin and prints the value of profile-settings.profile
  awk '
    $0=="profile-settings:" {inps=1; next}
    inps && $0=="  profile:" {inprof=1; next}
    # Only capture the final resolved value line for the parameter (4-space indent),
    # not intermediate log entries (which also contain "value: ...").
    inprof && $0 ~ /^    value: / {print $2; exit}
  '
}

get_ai_engine_value() {
  # Reads YAML on stdin and prints the value of ai-chat.ai-engine
  awk '
    $0=="ai-chat:" {inchat=1; next}
    inchat && $0=="  ai-engine:" {inengine=1; next}
    # Only capture the final resolved value line for the parameter (4-space indent),
    # not intermediate log entries (which also contain "value: ...").
    inengine && $0 ~ /^    value: / {print $2; exit}
  '
}

get_ai_engine_block() {
  # Reads YAML on stdin and prints the ai-chat.ai-engine subtree block (until next 2-space key)
  awk '
    $0=="ai-chat:" {inchat=1}
    inchat && $0=="  ai-engine:" {inengine=1}
    inengine {print}
    inengine && $0 ~ /^  [a-zA-Z0-9_-]+:$/ && $0 != "  ai-engine:" {exit}
  '
}

run_cmd() {
  # run_cmd <label> [ENV_K=V ...] -- [args...]
  local label="$1"
  shift

  echo "==> $label"

  # Collect env assignments until we hit "--"
  local envs=()
  while [[ $# -gt 0 && "$1" != "--" ]]; do
    envs+=("$1")
    shift
  done
  if [[ $# -gt 0 && "$1" == "--" ]]; then
    shift
  fi

  out="$(
    cd "$geppetto_dir" && \
      env "${envs[@]}" \
        go run "./cmd/examples/simple-inference" simple-inference --print-parsed-parameters "$@" 2>&1
  )"

  # Always print the output (helps debugging in CI/term logs)
  echo "$out"

  # Return both full output and extracted YAML via globals
  yaml_out="$(printf '%s\n' "$out" | extract_yaml_from_output)"
}

assert_ok_profile_engine() {
  local expected_profile="$1"
  local expected_engine="$2"

  actual_profile="$(printf '%s\n' "$yaml_out" | get_profile_value || true)"
  actual_engine="$(printf '%s\n' "$yaml_out" | get_ai_engine_value || true)"

  if [[ "$actual_profile" != "$expected_profile" ]]; then
    echo "ERROR: expected profile=$expected_profile, got profile=$actual_profile" >&2
    return 1
  fi
  if [[ "$actual_engine" != "$expected_engine" ]]; then
    echo "ERROR: expected ai-engine=$expected_engine, got ai-engine=$actual_engine" >&2
    return 1
  fi
}

assert_engine_block_has_source() {
  local expected_source="$1"
  local expected_value="$2"
  block="$(printf '%s\n' "$yaml_out" | get_ai_engine_block)"
  echo "$block" | grep -q "source: ${expected_source}" || {
    echo "ERROR: expected ai-engine block to include source: ${expected_source}" >&2
    echo "$block" >&2
    return 1
  }
  echo "$block" | grep -q "value: ${expected_value}" || {
    echo "ERROR: expected ai-engine block to include value: ${expected_value}" >&2
    echo "$block" >&2
    return 1
  }
}

run_fail() {
  local profile="$1"
  echo "==> expecting FAIL: PINOCCHIO_PROFILE=$profile"
  set +e
  out="$(
    cd "$geppetto_dir" && \
      PINOCCHIO_PROFILE_FILE="$profile_file" \
      PINOCCHIO_PROFILE="$profile" \
        go run "./cmd/examples/simple-inference" simple-inference --print-parsed-parameters "hello" 2>&1
  )"
  code=$?
  set -e

  if [[ $code -eq 0 ]]; then
    echo "ERROR: expected failure, but command succeeded" >&2
    echo "$out" >&2
    return 1
  fi

  # Be flexible: depending on Cobra wrapping/printing, error text may vary.
  if echo "$out" | grep -q "profile ${profile} not found"; then
    echo "OK: failed with expected 'profile not found' error"
    return 0
  fi
  if echo "$out" | grep -q "profile file .* does not exist"; then
    echo "OK: failed with expected 'profile file does not exist' error"
    return 0
  fi

  echo "ERROR: failed as expected (exit=$code) but error text did not match expectations" >&2
  echo "$out" >&2
  return 1
}

echo "Repo root:     $repo_root"
echo "Geppetto dir:  $geppetto_dir"
echo "Example dir:   $example_dir"
echo "Profile file:  $profile_file"
echo

# -----------------------------------------------------------------------------
# Test matrix
#
# We validate:
# - profile selection via env
# - profile selection via flags (overrides env)
# - profile file selection via env and flags
# - overriding values that came from profiles via config/env/flags
# -----------------------------------------------------------------------------

# 1) Env selects profile + env selects profile file
run_cmd "env profile selects engine (gemini-2.5-pro)" \
  PINOCCHIO_PROFILE_FILE="$profile_file" \
  PINOCCHIO_PROFILE="gemini-2.5-pro" \
  -- "hello"
assert_ok_profile_engine "gemini-2.5-pro" "gemini-2.5-pro"
assert_engine_block_has_source "profiles" "gemini-2.5-pro"

# 2) Flags select profile + flags select profile file (should override env values)
run_cmd "flags override env profile selection (expect gemini-2.5-flash)" \
  PINOCCHIO_PROFILE_FILE="$profile_file" \
  PINOCCHIO_PROFILE="gemini-2.5-pro" \
  -- --profile-file "$profile_file" --profile "gemini-2.5-flash" "hello"
assert_ok_profile_engine "gemini-2.5-flash" "gemini-2.5-flash"
assert_engine_block_has_source "profiles" "gemini-2.5-flash"

# 3) Override a profile-provided value via env (env > profiles)
run_cmd "env overrides ai-engine (env > profiles)" \
  PINOCCHIO_PROFILE_FILE="$profile_file" \
  PINOCCHIO_PROFILE="gemini-2.5-pro" \
  PINOCCHIO_AI_ENGINE="sonnet-4-5" \
  -- "hello"
assert_ok_profile_engine "gemini-2.5-pro" "sonnet-4-5"
assert_engine_block_has_source "env" "sonnet-4-5"

# 4) Override env via flags (flags > env)
run_cmd "flags override env ai-engine (flags > env > profiles)" \
  PINOCCHIO_PROFILE_FILE="$profile_file" \
  PINOCCHIO_PROFILE="gemini-2.5-pro" \
  PINOCCHIO_AI_ENGINE="sonnet-4-5" \
  -- --ai-engine "gemini-2.5-flash" "hello"
assert_ok_profile_engine "gemini-2.5-pro" "gemini-2.5-flash"
assert_engine_block_has_source "cobra" "gemini-2.5-flash"

# 5) Override profile via config (config > profiles, but < env/flags)
cfg="$(mktemp)"
cfg_yaml="${cfg}.yaml"
mv "$cfg" "$cfg_yaml"
cfg="$cfg_yaml"
cat >"$cfg" <<EOF
ai-chat:
  ai-engine: gemini-2.5-flash
EOF
run_cmd "config overrides ai-engine (config > profiles)" \
  PINOCCHIO_PROFILE_FILE="$profile_file" \
  PINOCCHIO_PROFILE="gemini-2.5-pro" \
  -- --config-file "$cfg" "hello"
assert_ok_profile_engine "gemini-2.5-pro" "gemini-2.5-flash"
assert_engine_block_has_source "config" "gemini-2.5-flash"
rm -f "$cfg"

# 6) Unknown profile should fail
run_fail "foobar"

echo
echo "ALL OK"


