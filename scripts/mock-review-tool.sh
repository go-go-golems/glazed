#!/usr/bin/env bash
set -euo pipefail

# Mock Code Review Tool
#
# Receives the same prompt as the Codex tool, prints it for debugging,
# but returns a canned mock response without calling any API.
#
# This is useful for testing the prompt rendering pipeline.
#
# Optional env:
#   MOCK_REVIEW_SCENARIO - Scenario to use (default: canned)
#                          Options: canned, approve, request_changes, comment, full
#   LOG_INPUT - Always log input to stderr (0|1|true|false, default: 1 for mock)
#   LOG_OUTPUT - Log output to stderr (0|1|true|false)

SCENARIO="${MOCK_REVIEW_SCENARIO:-canned}"
LOG_INPUT="${LOG_INPUT:-1}"
LOG_OUTPUT="${LOG_OUTPUT:-0}"

is_truthy() {
    case "${1,,}" in
        1|true|yes|y|on) return 0 ;;
        *) return 1 ;;
    esac
}

# Read stdin (the rendered prompt)
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

# Always print the prompt for debugging (that's the point of mock mode)
echo "=== MOCK REVIEW TOOL ===" >&2
echo "Scenario: $SCENARIO" >&2
echo "" >&2

if is_truthy "$LOG_INPUT"; then
    echo "=== FULL RENDERED PROMPT ===" >&2
    echo "$prompt_text" >&2
    echo "=== END PROMPT ===" >&2
    echo "" >&2
fi

echo "NOTE: This is MOCK mode - no API calls are made." >&2
echo "The prompt above is exactly what would be sent to Codex CLI." >&2
echo "" >&2

# Generate mock response based on scenario
case "$SCENARIO" in
    approve)
        review_json=$(cat << 'EOF'
{
  "summary_markdown": "### ✅ Mock Review: Approved\n\n**This is a mock review - no actual analysis performed.**\n\n- Scenario: `approve`\n- Files would be analyzed in real mode",
  "review_decision": "APPROVE",
  "review_body": "Mock approval. In real mode, Codex would analyze the changes and provide detailed feedback.",
  "comments": []
}
EOF
)
        ;;
    request_changes)
        review_json=$(cat << 'EOF'
{
  "summary_markdown": "### ⚠️ Mock Review: Changes Requested\n\n**This is a mock review - no actual analysis performed.**\n\n- Scenario: `request_changes`\n- Would identify issues in real mode",
  "review_decision": "REQUEST_CHANGES",
  "review_body": "Mock request for changes. In real mode, Codex would identify specific issues requiring attention.",
  "comments": [
    {
      "path": "web/src/example.tsx",
      "body": "🔧 **Mock Comment**: This is a placeholder comment. In real mode, Codex would provide specific feedback about deprecated patterns.",
      "subject_type": "file"
    }
  ]
}
EOF
)
        ;;
    comment)
        review_json=$(cat << 'EOF'
{
  "summary_markdown": "### 💬 Mock Review: Comments Only\n\n**This is a mock review - no actual analysis performed.**\n\n- Scenario: `comment`\n- Would provide suggestions in real mode",
  "review_decision": "COMMENT",
  "review_body": "Mock comment review. In real mode, Codex would provide helpful suggestions without blocking.",
  "comments": []
}
EOF
)
        ;;
    full)
        review_json=$(cat << 'EOF'
{
  "summary_markdown": "### 🔍 Mock Review: Full Analysis\n\n**This is a mock review - no actual analysis performed.**\n\n#### What would happen in real mode:\n- Analyze all changed files in `web/`\n- Check for deprecated React patterns\n- Check for deprecated Redux patterns\n- Identify manual fetch() calls\n- Flag useState+useEffect anti-patterns\n\n#### Mock Statistics\n- Files analyzed: 0 (mock)\n- Issues found: 0 (mock)\n- Suggestions: 0 (mock)",
  "review_decision": "COMMENT",
  "review_body": "This is a **full mock review**.\n\nIn real mode with Codex CLI, this would:\n\n1. Parse the diff from the PR\n2. Analyze each file for deprecated patterns\n3. Generate inline comments with specific line numbers\n4. Provide actionable suggestions\n\nThe prompt shown in the logs above is exactly what would be sent to the model.",
  "issue_comment": "📋 **Mock Review Complete**\n\nThis mock review demonstrates the full prompt rendering pipeline. Check the workflow logs to see the exact prompt that would be sent to Codex.",
  "comments": [
    {
      "path": "web/src/main.tsx",
      "body": "📌 **Mock Inline Comment**\n\nIn real mode, Codex would analyze this file and provide specific feedback about:\n- Deprecated lifecycle methods\n- Legacy patterns\n- Suggested improvements",
      "subject_type": "file"
    }
  ]
}
EOF
)
        ;;
    canned|*)
        review_json=$(cat << 'EOF'
{
  "summary_markdown": "### 🧪 Mock Review (Canned Response)\n\n**This is a mock review - no actual analysis performed.**\n\nUse this mode to test prompt rendering without API calls.\n\n- Mode: `mock`\n- Scenario: `canned`",
  "review_decision": "COMMENT",
  "review_body": "Canned mock response. The full rendered prompt has been logged to stderr for debugging.\n\nTo run a real review, use `mode: codex` instead of `mode: mock`.",
  "comments": []
}
EOF
)
        ;;
esac

if is_truthy "$LOG_OUTPUT"; then
    echo "=== MOCK REVIEW OUTPUT ===" >&2
    echo "$review_json" | jq . >&2
    echo "=== END OUTPUT ===" >&2
fi

# Output the mock review JSON
echo "$review_json"
