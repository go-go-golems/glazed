package markdown

import (
	"reflect"
	"testing"
)

func TestExtractQuotedBlocks(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []MarkdownBlock
	}{
		{
			name:  "Basic Code Block Extraction",
			input: "```go\nfmt.Println(\"Hello, World!\")\n```",
			expected: []MarkdownBlock{
				{Type: Code, Language: "go", Content: "fmt.Println(\"Hello, World!\")"},
			},
		},
		{
			name:  "Multiple Code Block Extractions",
			input: "```go\nfmt.Println(\"Hello, Go!\")\n```\n```python\nprint(\"Hello, Python!\")\n```",
			expected: []MarkdownBlock{
				{Type: Code, Language: "go", Content: "fmt.Println(\"Hello, Go!\")"},
				{Type: Code, Language: "python", Content: "print(\"Hello, Python!\")"},
			},
		},
		{
			name:  "Language Detection",
			input: "```rust\nprintln!(\"Hello, Rust!\");\n```",
			expected: []MarkdownBlock{
				{Type: Code, Language: "rust", Content: "println!(\"Hello, Rust!\");"},
			},
		},
		{
			name:  "Mixed Content Extraction",
			input: "This is a normal text\n```js\nconsole.log(\"Hello, JavaScript!\");\n```\nThis is another normal text",
			expected: []MarkdownBlock{
				{Type: Normal, Content: "This is a normal text"},
				{Type: Code, Language: "js", Content: "console.log(\"Hello, JavaScript!\");"},
				{Type: Normal, Content: "This is another normal text"},
			},
		},

		{
			name:  "No Code Block",
			input: "This is just a normal text without any code block.",
			expected: []MarkdownBlock{
				{Type: Normal, Content: "This is just a normal text without any code block."},
			},
		},
		{
			name:     "Empty Input",
			input:    "",
			expected: []MarkdownBlock{},
		},

		{
			name:  "Incomplete Code Block",
			input: "```python\nprint(\"Hello, World!\"",
			expected: []MarkdownBlock{
				{Type: Code, Language: "python", Content: "print(\"Hello, World!\""},
			},
		},
		{
			name:  "Code Block with No Language",
			input: "```\nconsole.log('No language specified');\n```",
			expected: []MarkdownBlock{
				{Type: Code, Language: "", Content: "console.log('No language specified');"},
			},
		},
		{
			name:  "Normal Text Before and After Code Blocks",
			input: "This is normal text.\n```go\nfmt.Println(\"Hello\")\n```\nThis is also normal text.",
			expected: []MarkdownBlock{
				{Type: Normal, Content: "This is normal text."},
				{Type: Code, Language: "go", Content: "fmt.Println(\"Hello\")"},
				{Type: Normal, Content: "This is also normal text."},
			},
		},
		{
			name:  "Multiple Lines Within Blocks",
			input: "This is a normal block\nwith multiple lines.\n```python\nprint(\"Line 1\")\nprint(\"Line 2\")\n```",
			expected: []MarkdownBlock{
				{Type: Normal, Content: "This is a normal block\nwith multiple lines."},
				{Type: Code, Language: "python", Content: "print(\"Line 1\")\nprint(\"Line 2\")"},
			},
		},
		{
			name:  "Special Characters in Code Block",
			input: "```js\nconsole.log(\"<>&*\");\n```",
			expected: []MarkdownBlock{
				{Type: Code, Language: "js", Content: "console.log(\"<>&*\");"},
			},
		},
		{
			name:  "Whitespace Handling",
			input: "   This has leading spaces.\n```python\n   print(\"Indented Python Code\")\n```",
			expected: []MarkdownBlock{
				{Type: Normal, Content: "   This has leading spaces."},
				{Type: Code, Language: "python", Content: "   print(\"Indented Python Code\")"},
			},
		},
		{
			name:  "Blocks with Only Language Specified",
			input: "```python\n```",
			expected: []MarkdownBlock{
				{Type: Code, Language: "python", Content: ""},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ExtractAllBlocks(tt.input)
			// compare in a loop
			for index, block := range got {
				if index >= len(tt.expected) {
					t.Errorf("Expected: none: Got: %v", block)
					continue
				}
				if !reflect.DeepEqual(block, tt.expected[index]) {
					t.Errorf("Expected: %v, Got: %v", tt.expected[index], block)
				}
			}
		})
	}
}

func TestExtractCodeBlocksWithComments(t *testing.T) {
	tests := []struct {
		name       string
		input      string
		withQuotes bool
		expected   []string
	}{
		{
			name:       "Basic Functionality",
			input:      "This is a description.\n```Go\nfmt.Println(\"Hello, World!\")\n```",
			withQuotes: true,
			expected:   []string{"```Go\n// This is a description.\nfmt.Println(\"Hello, World!\")\n```"},
		},
		{
			name:       "Multiple Code Blocks",
			input:      "Description 1.\n```Go\nfmt.Println(\"Hello, World!\")\n```\nDescription 2.\n```Python\nprint(\"Hello, World!\")\n```",
			withQuotes: true,
			expected: []string{
				"```Go\n// Description 1.\nfmt.Println(\"Hello, World!\")\n```",
				"```Python\n# Description 2.\nprint(\"Hello, World!\")\n```",
			},
		},
		{
			name:       "End Non-Code Content",
			input:      "```Go\nfmt.Println(\"Hello, World!\")\n```\nThis is a description.",
			withQuotes: true,
			expected:   []string{"```Go\nfmt.Println(\"Hello, World!\")\n// This is a description.\n```"},
		},
		{
			name:       "No Code Block",
			input:      "This is just a text description without any code.",
			withQuotes: true,
			expected:   nil, // no code blocks to return
		},

		{
			name:       "Unknown Language",
			input:      "This is a description.\n```bliss\nSome bliss code here\n```",
			withQuotes: true,
			expected:   []string{"```bliss\nSome bliss code here\n```"}, // Description is not added as a comment since "bliss" is unknown
		},
		{
			name:       "Empty Language",
			input:      "This is a description.\n```\nSome code without language\n```",
			withQuotes: true,
			expected:   []string{"```\nSome code without language\n```"}, // Description is not added as a comment since language is unspecified
		},
		{
			name:       "Non-Code Content Between Multiple Code Blocks",
			input:      "Description 1.\n```Go\nfmt.Println(\"Hello from Go!\")\n```\nDescription 2.\n```Python\nprint(\"Hello from Python!\")\n```",
			withQuotes: true,
			expected: []string{
				"```Go\n// Description 1.\nfmt.Println(\"Hello from Go!\")\n```",
				"```Python\n# Description 2.\nprint(\"Hello from Python!\")\n```",
			},
		},
		{
			name:       "Code Block Without Preceding Text",
			input:      "```Go\nfmt.Println(\"Hello, World!\")\n```",
			withQuotes: true,
			expected:   []string{"```Go\nfmt.Println(\"Hello, World!\")\n```"}, // No preceding text, so the code block is unchanged
		},
		{
			name:       "Multiple Non-Code Content",
			input:      "Description 1.\nDescription 2.\n```Go\nfmt.Println(\"Hello, World!\")\n```",
			withQuotes: true,
			expected:   []string{"```Go\n// Description 1.\n// Description 2.\nfmt.Println(\"Hello, World!\")\n```"},
		},
		{
			name:       "Mixed Known and Unknown Languages",
			input:      "Description 1.\n```Go\nfmt.Println(\"Hello from Go!\")\n```\nDescription 2.\n```bliss\nSome bliss code here\n```",
			withQuotes: true,
			expected: []string{
				"```Go\n// Description 1.\nfmt.Println(\"Hello from Go!\")\n```",
				"```bliss\nSome bliss code here\n```", // Description 2 is not added as a comment since "bliss" is unknown
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output := ExtractCodeBlocksWithComments(tt.input, tt.withQuotes)
			if !reflect.DeepEqual(output, tt.expected) {
				t.Errorf("got %v, want %v", output, tt.expected)
			}
		})
	}
}
