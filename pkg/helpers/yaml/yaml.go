package yaml

import (
	"regexp"
	"strings"
)

// Clean tries to cleanup YAML that might be invalid, for example
// if coming out of a LLM.
// This is quite a hacky solution, and will only work for simple YAML.
func Clean(s string, fromMarkdown bool) string {
	// split by lines and iterate over them
	lines := strings.Split(s, "\n")
	ret := []string{}

	isInMarkdown := false
	isInQuotedString := false
	quoteIndent := 0

	re := regexp.MustCompile(`\s+$`)

	for _, line := range lines {
		// check if the first string is a ``` markdown marker
		if strings.HasPrefix(line, "```") {
			// if it is, remove it and the next line
			isInMarkdown = !isInMarkdown
			continue
		}

		if fromMarkdown && !isInMarkdown {
			continue
		}

		if isInQuotedString {
			// if quoteIndent is 0, we just started a multiline string
			if quoteIndent == 0 {
				quoteIndent, _ = getIndentLevel(line)
			}

			// count start indent
			indent, isWhitespace := getIndentLevel(line)

			// check if we are still in the quoted string
			if indent >= quoteIndent || isWhitespace {
				// chop off rightmost whitespace because otherwise yaml encoder won't output it as multiline string
				line = re.ReplaceAllString(line, "")
				ret = append(ret, line)
				continue
			} else {
				isInQuotedString = false
			}
		}

		// check if we have a colon on the line
		if strings.Contains(line, ":") {
			// split on the colon
			parts := strings.Split(line, ":")
			if len(parts) > 2 {
				key := parts[0]
				value := strings.TrimSpace(strings.Join(parts[1:], ":"))

				// check if value is a quoted string
				if strings.HasPrefix(value, "\"") && strings.HasSuffix(value, "\"") ||
					strings.HasPrefix(value, "'") && strings.HasSuffix(value, "'") {
					ret = append(ret, line)
					continue
				}

				// otherwise quote
				ret = append(ret, key+": \""+value+"\"")
				continue
			} else if len(parts) == 2 {
				key := parts[0]
				value := strings.TrimSpace(parts[1])

				// check if value is a quoted string
				if strings.HasPrefix(value, "\"") && strings.HasSuffix(value, "\"") ||
					strings.HasPrefix(value, "'") && strings.HasSuffix(value, "'") {
					ret = append(ret, line)
					continue
				}

				// check if this is the start of a multine string
				if strings.HasPrefix(value, "|") {
					isInQuotedString = true

					// quoteIndent of 0 signals a new block start, indent will be measured on next line
					quoteIndent = 0
					// check if starts with special characters
					//  &, *, !, %, |, [, ], {, }, ,, >, ', " or ?
				} else if len(value) > 0 && strings.ContainsAny(value[:1], "&*!%|[]{}>,?'\"?") {
					ret = append(ret, key+": \""+value+"\"")
					continue
				}
				ret = append(ret, line)
				continue
			}
		}

		ret = append(ret, line)
	}

	ret_ := strings.Join(ret, "\n")
	return ret_
}

func getIndentLevel(line string) (int, bool) {
	indent := 0
	isWhitespace := true
	for _, c := range line {
		if c == ' ' {
			indent++
		} else {
			isWhitespace = false
			break
		}
	}
	return indent, isWhitespace
}
