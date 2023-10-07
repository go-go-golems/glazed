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
				{Type: Normal, Content: "```python"},
				{Type: Normal, Content: "print(\"Hello, World!\""},
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
