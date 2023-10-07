package main

import (
	_ "embed"
	"fmt"
	"github.com/go-go-golems/glazed/pkg/helpers/markdown"
	"github.com/go-go-golems/glazed/pkg/helpers/strings"
)

//go:embed "test.md"
var md string

func main() {
	// Extract all blocks
	fmt.Println("Extract all blocks")
	fmt.Println("==================")
	blocks := markdown.ExtractAllBlocks(md)
	for _, block := range blocks {
		fmt.Println("---")
		fmt.Println(block.Type, block.Language, block.Content)
	}
	fmt.Println()

	// Extract quoted blocks
	fmt.Println("Extract quoted blocks")
	fmt.Println("====================")
	quotedBlocks := markdown.ExtractQuotedBlocks(md, true)
	for _, block := range quotedBlocks {
		fmt.Println("---")
		fmt.Println(block)
	}
	fmt.Println()

	// Extract code blocks with comments
	fmt.Println("Extract code blocks with comments")
	fmt.Println("================================")
	codeBlocksWithComments := markdown.ExtractCodeBlocksWithComments(md, true)
	for _, block := range codeBlocksWithComments {
		fmt.Println("---")
		fmt.Println(block)
	}
	fmt.Println()

	// Generate comment
	fmt.Println("Generate comment")
	fmt.Println("================")
	comment := strings.GenerateComment("This is a comment.", strings.GoLang)
	fmt.Println(comment)
	fmt.Println()

	// Markdown code block to language
	fmt.Println("Markdown code block to language")
	fmt.Println("==============================")
	lang := strings.MarkdownCodeBlockToLanguage("go")
	fmt.Println(lang)
	fmt.Println()
}
