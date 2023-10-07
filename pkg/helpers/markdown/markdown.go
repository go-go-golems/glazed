package markdown

import (
	"bufio"
	strings2 "github.com/go-go-golems/glazed/pkg/helpers/strings"
	"strings"
)

type State int

const (
	OutsideBlock State = iota
	InsideBlock
)

type BlockType string

const (
	Normal BlockType = "Normal"
	Code   BlockType = "Code"
)

type MarkdownBlock struct {
	Type     BlockType
	Language string
	Content  string
}

// ExtractAllBlocks processes a given markdown string to split it into a series of blocks.
//
// **Usage**:
// Call this function with a markdown string as the argument. It will return a slice
// of MarkdownBlock structs, each representing a block of content (either normal or code).
//
// **Inner Workings**:
// The function employs a state machine approach to parse through the markdown content.
// It recognizes and separates out code blocks (enclosed with ```) and normal text blocks.
// For code blocks, it also identifies the optional language specifier.
func ExtractAllBlocks(input string) []MarkdownBlock {
	var result []MarkdownBlock
	state := OutsideBlock
	var blockLines []string

	scanner := bufio.NewScanner(strings.NewReader(input))
	language := ""
	for scanner.Scan() {
		line := scanner.Text()

		switch state {
		case OutsideBlock:
			if strings.HasPrefix(line, "```") {
				state = InsideBlock
				language = strings.TrimPrefix(line, "```")
				if len(blockLines) > 0 {
					// For normal blocks
					result = append(result, MarkdownBlock{Type: Normal, Content: strings.Join(blockLines, "\n")})
				}
				blockLines = nil // reset blockLines
				continue
			}
			blockLines = append(blockLines, line)
		case InsideBlock:
			if strings.HasPrefix(line, "```") {
				state = OutsideBlock
				content := strings.Join(blockLines, "\n") // excluding the language line
				result = append(result, MarkdownBlock{Type: Code, Language: language, Content: content})
				blockLines = nil // reset blockLines
				language = ""
			} else {
				blockLines = append(blockLines, line)
			}
		}
	}

	// Handle any remaining normal content outside of code blocks
	if len(blockLines) > 0 && state == OutsideBlock {
		result = append(result, MarkdownBlock{Type: Normal, Content: strings.Join(blockLines, "\n")})
	}

	return result
}

// ExtractQuotedBlocks extracts only the code blocks from a markdown string.
//
// **Usage**:
// This function takes in a markdown string and a boolean `withQuotes`. If `withQuotes`
// is set to true, the output will include the enclosing ``` for each code block.
// Otherwise, only the inner content of the code block is returned.
//
// **Inner Workings**:
// This function leverages the `ExtractAllBlocks` function to first get all blocks
// from the markdown content. It then filters out only the code blocks and processes
// them based on the `withQuotes` parameter to decide on the inclusion of the enclosing ``` marks.
func ExtractQuotedBlocks(input string, withQuotes bool) []string {
	blocks := ExtractAllBlocks(input)
	var result []string
	for _, block := range blocks {
		if block.Type != Code {
			continue
		}
		if withQuotes {
			result = append(result, "```"+block.Language+"\n"+block.Content+"\n```")
		} else {
			result = append(result, block.Content)
		}
	}

	return result
}

// ExtractCodeBlocksWithComments processes a markdown string to return only code blocks.
// Any non-code content preceding a code block is added as a comment within the respective code block.
// If there are non-code contents left at the end without a following code block, they are appended
// as comments to the last code block.
//
// **Usage**:
// Call this function with a markdown string and a boolean `withQuotes`. If `withQuotes` is set to true,
// the output will include the enclosing ``` for each code block. Otherwise, only the inner content of
// the code block (and the preceding comments) is returned.
func ExtractCodeBlocksWithComments(input string, withQuotes bool) []string {
	blocks := ExtractAllBlocks(input)
	var result []string
	var pendingComments []string
	lastLanguage := strings2.None
	lastBlockLanguage := ""
	codeContent := ""

	for _, block := range blocks {
		if block.Type == Code {
			// If there are pending comments (non-code blocks), convert them to comments in the current language
			codeContent = block.Content
			language := strings2.MarkdownCodeBlockToLanguage(block.Language)
			lastLanguage = language
			lastBlockLanguage = block.Language

			if len(pendingComments) > 0 {

				// only append comments for languages we know how to generate comments for
				if language != strings2.None {
					commentedText := strings2.GenerateComment(strings.Join(pendingComments, "\n"), language)
					codeContent = commentedText + "\n" + codeContent
				}
				pendingComments = nil
			}

			if withQuotes {
				result = append(result, "```"+block.Language+"\n"+codeContent+"\n```")
			} else {
				result = append(result, codeContent)
			}
		} else {
			// Accumulate non-code blocks to be commented later
			pendingComments = append(pendingComments, block.Content)
		}
	}

	// If there's any pending comment without a following code block, append to the previous block
	if len(pendingComments) > 0 {
		if lastLanguage != strings2.None && len(result) > 0 {
			codeContent += "\n" + strings2.GenerateComment(strings.Join(pendingComments, "\n"), lastLanguage)
			res := ""
			if withQuotes {
				res = "```" + lastBlockLanguage + "\n" + codeContent + "\n```"
			} else {
				res = codeContent
			}
			result[len(result)-1] = res
		}
	}

	return result
}
