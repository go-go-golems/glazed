#!/usr/bin/env bash
set -euo pipefail

# OpenAI Codex CLI Code Review Tool
#
# Uses @openai/codex CLI in non-interactive mode for code reviews.
# See: https://developers.openai.com/codex/noninteractive
#      https://developers.openai.com/codex/models/
#
# Receives prompt on stdin, outputs ReviewResult JSON.
#
# Required env:
#   CODEX_API_KEY or OPENAI_API_KEY - API key for authentication
#
# Optional env:
#   CODEX_MODEL - Model to use (default: gpt-5.1-codex-mini)
#                 Options: gpt-5.2-codex (most capable), gpt-5.1-codex-mini (cost-effective)
#   CODEX_SANDBOX - Sandbox mode: read-only, workspace-write, danger-full-access (default: read-only)
#   LOG_INPUT - Log input to stderr (0|1|true|false)
#   LOG_OUTPUT - Log output to stderr (0|1|true|false)
#   USE_API_FALLBACK - Fall back to Chat API if codex CLI fails (0|1)

# Support both CODEX_API_KEY (preferred per docs) and OPENAI_API_KEY
if [[ -n "${CODEX_API_KEY:-}" ]]; then
    API_KEY="$CODEX_API_KEY"
elif [[ -n "${OPENAI_API_KEY:-}" ]]; then
    API_KEY="$OPENAI_API_KEY"
    export CODEX_API_KEY="$OPENAI_API_KEY"
else
    echo "Error: CODEX_API_KEY or OPENAI_API_KEY is required" >&2
    exit 1
fi

# Model selection - see https://developers.openai.com/codex/models/
# Recommended: gpt-5.2-codex (most advanced) or gpt-5.1-codex-mini (cost-effective)
MODEL="${CODEX_MODEL:-gpt-5.1-codex-mini}"
SANDBOX="${CODEX_SANDBOX:-read-only}"
LOG_INPUT="${LOG_INPUT:-0}"
LOG_OUTPUT="${LOG_OUTPUT:-0}"
USE_API_FALLBACK="${USE_API_FALLBACK:-1}"

is_truthy() {
    case "${1,,}" in
        1|true|yes|y|on) return 0 ;;
        *) return 1 ;;
    esac
}

# Create temp directory for schema and output
TEMP_DIR=$(mktemp -d)
trap 'rm -rf "$TEMP_DIR"' EXIT

# ReviewResult JSON Schema for --output-schema
SCHEMA_SOURCE="/opt/go-go-agent-action/schema/review-result.schema.json"
if [[ ! -f "$SCHEMA_SOURCE" ]]; then
    echo "Error: ReviewResult schema not found at $SCHEMA_SOURCE" >&2
    exit 1
fi
cp "$SCHEMA_SOURCE" "$TEMP_DIR/schema.json"

# Read stdin (could be JSON with prompt_text, or just the prompt)
stdin_content="$(cat)"

# Try to extract prompt_text if input is JSON with that field
if echo "$stdin_content" | jq -e . >/dev/null 2>&1; then
    if echo "$stdin_content" | jq -e '.prompt_text' >/dev/null 2>&1; then
        prompt_text="$(echo "$stdin_content" | jq -r '.prompt_text')"
        if [[ -z "$prompt_text" || "$prompt_text" == "null" ]]; then
            echo "Error: prompt_text is empty. Ensure prompt_template_path is set." >&2
            exit 1
        fi
    else
        echo "Error: JSON input missing prompt_text. Ensure prompt_template_path is set." >&2
        exit 1
    fi
else
    # Assume stdin is the raw prompt
    prompt_text="$stdin_content"
fi

if is_truthy "$LOG_INPUT"; then
    echo "=== CODEX REVIEW TOOL INPUT ===" >&2
    echo "$prompt_text" >&2
    echo "=== END INPUT ===" >&2
fi

# Build the full prompt with output instructions
full_prompt=$(cat << PROMPT
$prompt_text

IMPORTANT: Your final response MUST be a valid JSON object matching this exact schema:
Schema path: /opt/go-go-agent-action/schema/review-result.schema.json

Output ONLY the JSON object, no other text.
PROMPT
)

# Always print full prompt for debugging
echo "=== FULL PROMPT BEING SENT ===" >&2
echo "$full_prompt" >&2
echo "=== END FULL PROMPT ===" >&2
echo "" >&2
echo "Model: $MODEL, Sandbox: $SANDBOX" >&2

# Determine codex command
if command -v codex &>/dev/null; then
    CODEX_CMD="codex"
else
    echo "Installing @openai/codex..." >&2
    CODEX_CMD="npx --yes @openai/codex"
fi

output_file="$TEMP_DIR/output.json"
events_file="$TEMP_DIR/events.jsonl"

# Run codex exec in non-interactive mode
# See: https://developers.openai.com/codex/noninteractive
echo "Running: codex exec -m $MODEL --sandbox $SANDBOX ..." >&2

codex_success=0
if $CODEX_CMD exec \
    -m "$MODEL" \
    --sandbox "$SANDBOX" \
    --json \
    --output-schema "$TEMP_DIR/schema.json" \
    -o "$output_file" \
    --skip-git-repo-check \
    --full-auto \
    "$full_prompt" \
    > "$events_file" 2>&1; then
    codex_success=1
    echo "Codex exec completed successfully" >&2
else
    echo "Codex exec failed with exit code $?" >&2
    echo "Events log:" >&2
    cat "$events_file" >&2 || true
fi

# Try to extract the result
review_json=""

if [[ "$codex_success" == "1" ]] && [[ -f "$output_file" ]]; then
    review_json="$(cat "$output_file")"
    
    # Validate it's proper JSON
    if ! echo "$review_json" | jq -e . >/dev/null 2>&1; then
        echo "Codex output was not valid JSON, trying to extract..." >&2
        echo "Raw output:" >&2
        cat "$output_file" >&2
        # Try to extract JSON from the output
        review_json=$(echo "$review_json" | grep -o '{.*}' | head -1 || true)
    fi
fi

# Final validation
if [[ -z "$review_json" ]] || ! echo "$review_json" | jq -e . >/dev/null 2>&1; then
    echo "Failed to get valid review output" >&2
    cat << 'EOF'
{
  "summary_markdown": "### ❌ Review Failed\n\nCodex CLI could not generate a valid review",
  "review_decision": "COMMENT",
  "review_body": "Failed to generate review. Check workflow logs for details.",
  "comments": []
}
EOF
    exit 0
fi

if is_truthy "$LOG_OUTPUT"; then
    echo "=== CODEX REVIEW TOOL OUTPUT ===" >&2
    echo "$review_json" | jq . >&2
    echo "=== END OUTPUT ===" >&2
fi

# Output the review JSON
echo "$review_json"
