package json

import (
	"bytes"
	"encoding/json"
	"github.com/go-go-golems/glazed/pkg/helpers/markdown"
	"strings"
)

// ExtractJSON extracts potential JSON blocks from the provided input string.
// The input string can either be raw JSON or markdown containing JSON code-blocks.
// The returned strings are sanitized.
// Codeblocks that do not contain valid JSON are ignored.
func ExtractJSON(input string) []string {
	var result []string

	input_ := SanitizeJSONString(input, false)

	// First, try to parse the entire string as JSON
	var temp map[string]interface{}
	if err := json.Unmarshal([]byte(input_), &temp); err == nil {
		result = append(result, input_)
		return result
	}

	// If that fails, scan for quoted blocks
	quotedBlocks := markdown.ExtractQuotedBlocks(input, false)
	for _, block := range quotedBlocks {
		var temp map[string]interface{}
		block = SanitizeJSONString(block, false)
		if err := json.Unmarshal([]byte(block), &temp); err == nil {
			result = append(result, block)
		}
	}

	return result
}

func SanitizeJSONString(input string, fromMarkdown bool) string {
	var result bytes.Buffer

	const (
		outsideQuotes = iota
		insideQuotes
		insideEscape
	)

	isInMarkdown := false

	state := outsideQuotes

	lines := strings.Split(input, "\n")

	for _, line := range lines {
		if strings.HasPrefix(line, "```") {
			isInMarkdown = !isInMarkdown
			continue
		}

		if fromMarkdown && !isInMarkdown {
			continue
		}
		for i := 0; i < len(line); i++ {
			ch := line[i]

			switch state {
			case outsideQuotes:
				if ch == '"' {
					state = insideQuotes
				}
				result.WriteByte(ch)

			case insideQuotes:
				if ch == '\\' {
					state = insideEscape
				} else if ch == '"' {
					state = outsideQuotes
				} else if ch == '\n' {
					result.WriteString("\\n")
					continue
				}
				result.WriteByte(ch)

			case insideEscape:
				state = insideQuotes
				result.WriteByte(ch)
			}
		}
	}

	return result.String()
}
