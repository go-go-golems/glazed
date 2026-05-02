package server

import (
	"bufio"
	"regexp"
	"strings"
	"unicode"
)

var headingIDNonWord = regexp.MustCompile(`[^a-z0-9\s-]`)
var headingIDWhitespace = regexp.MustCompile(`[\s-]+`)

// ExtractHeadings extracts lightweight subsection metadata from Markdown
// content. It supports ATX headings (# through ####), ignores fenced code
// blocks, and skips a duplicate top heading matching the section title.
func ExtractHeadings(content, sectionTitle string) []SectionHeading {
	var headings []SectionHeading
	scanner := bufio.NewScanner(strings.NewReader(content))
	inFence := false
	fenceMarker := ""

	for scanner.Scan() {
		line := scanner.Text()
		trimmed := strings.TrimSpace(line)
		if isFenceLine(trimmed) {
			marker := trimmed[:3]
			if !inFence {
				inFence = true
				fenceMarker = marker
			} else if marker == fenceMarker {
				inFence = false
				fenceMarker = ""
			}
			continue
		}
		if inFence || !strings.HasPrefix(trimmed, "#") {
			continue
		}

		level := countLeadingHashes(trimmed)
		if level < 1 || level > 4 || len(trimmed) <= level || !unicode.IsSpace(rune(trimmed[level])) {
			continue
		}
		text := strings.TrimSpace(trimmed[level:])
		text = strings.TrimSpace(strings.TrimRight(text, "#"))
		if text == "" || strings.EqualFold(text, sectionTitle) {
			continue
		}
		headings = append(headings, SectionHeading{
			ID:    SlugifyHeading(text),
			Level: level,
			Text:  text,
		})
	}

	return headings
}

// SlugifyHeading mirrors the frontend heading slug algorithm used by the help
// browser Markdown renderer.
func SlugifyHeading(text string) string {
	lower := strings.ToLower(strings.TrimSpace(text))
	lower = headingIDNonWord.ReplaceAllString(lower, "")
	lower = headingIDWhitespace.ReplaceAllString(lower, "-")
	lower = strings.Trim(lower, "-")
	if lower == "" {
		return "section"
	}
	return lower
}

func isFenceLine(trimmed string) bool {
	return strings.HasPrefix(trimmed, "```") || strings.HasPrefix(trimmed, "~~~")
}

func countLeadingHashes(s string) int {
	count := 0
	for _, r := range s {
		if r != '#' {
			break
		}
		count++
	}
	return count
}
