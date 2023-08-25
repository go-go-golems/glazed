package json

import (
	"bufio"
	"bytes"
	"encoding/json"
	"strings"
)

// ExtractJSON extracts potential JSON blocks from the provided input string.
// The input string can either be raw JSON or markdown containing JSON code-blocks.
// The returned strings are sanitized.
// Codeblocks that do not contain valid JSON are ignored.
func ExtractJSON(input string) []string {
	var result []string

	input_ := SanitizeJSONString(input)

	// First, try to parse the entire string as JSON
	var temp map[string]interface{}
	if err := json.Unmarshal([]byte(input_), &temp); err == nil {
		result = append(result, input_)
		return result
	}

	// If that fails, scan for quoted blocks
	quotedBlocks := ExtractQuotedBlocks(input)
	for _, block := range quotedBlocks {
		var temp map[string]interface{}
		block = SanitizeJSONString(block)
		if err := json.Unmarshal([]byte(block), &temp); err == nil {
			result = append(result, block)
		}
	}

	return result
}

type State int

const (
	OutsideBlock State = iota
	InsideBlock
)

// ExtractQuotedBlocks extracts blocks enclosed by ``` using a state machine.
func ExtractQuotedBlocks(input string) []string {
	var result []string
	state := OutsideBlock
	var blockLines []string

	scanner := bufio.NewScanner(strings.NewReader(input))
	for scanner.Scan() {
		line := scanner.Text()

		switch state {
		case OutsideBlock:
			if strings.HasPrefix(line, "```") {
				state = InsideBlock
				blockLines = nil // reset blockLines
			}
		case InsideBlock:
			if strings.HasPrefix(line, "```") {
				state = OutsideBlock
				result = append(result, strings.Join(blockLines, "\n"))
			} else {
				blockLines = append(blockLines, line)
			}
		}
	}

	return result
}

func SanitizeJSONString(input string) string {
	var result bytes.Buffer

	const (
		outsideQuotes = iota
		insideQuotes
		insideEscape
	)

	state := outsideQuotes

	for i := 0; i < len(input); i++ {
		ch := input[i]

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

	return result.String()
}
