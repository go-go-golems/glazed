package strings

import (
	"github.com/muesli/reflow/wordwrap"
	"strings"
)

// CommentType represents the type of comment (block or line)
type CommentType int

const (
	Block CommentType = iota
	Line
)

// Language represents the supported languages
type Language string

const (
	GoLang       Language = "Go"
	C            Language = "C"
	CPP          Language = "C++"
	Java         Language = "Java"
	Python       Language = "Python"
	JavaScript   Language = "JavaScript"
	Ruby         Language = "Ruby"
	Perl         Language = "Perl"
	Shell        Language = "Shell"
	PHP          Language = "PHP"
	Swift        Language = "Swift"
	Rust         Language = "Rust"
	Haskell      Language = "Haskell"
	HTML         Language = "HTML"
	XML          Language = "XML"
	CSS          Language = "CSS"
	SQL          Language = "SQL"
	R            Language = "Language"
	Scala        Language = "Scala"
	Kotlin       Language = "Kotlin"
	TypeScript   Language = "TypeScript"
	Matlab       Language = "Matlab"
	Assembly     Language = "Assembly"
	Lua          Language = "Lua"
	Fortran      Language = "Fortran"
	Bash         Language = "Bash"
	Groovy       Language = "Groovy"
	Prolog       Language = "Prolog"
	YAML         Language = "YAML"
	Markdown     Language = "Markdown"
	Dart         Language = "Dart"
	CoffeeScript Language = "CoffeeScript"
	FSharp       Language = "FSharp"
	Pascal       Language = "Pascal"
	Lisp         Language = "Lisp"
	Erlang       Language = "Erlang"
	Elixir       Language = "Elixir"
	HCL          Language = "HCL"
)

// CommentDelimiters represents the delimiters for each language
type CommentDelimiters struct {
	Type   CommentType
	Start  string
	End    string
	Inline string
}

// CommentDict holds the delimiters for each supported language
var CommentDict = map[Language]CommentDelimiters{
	GoLang:     {Line, "", "", "//"},
	C:          {Block, "/*", "*/", "//"},
	CPP:        {Block, "/*", "*/", "//"},
	Java:       {Block, "/*", "*/", "//"},
	Python:     {Line, "", "", "#"},
	JavaScript: {Block, "/*", "*/", "//"},
	Ruby:       {Line, "", "", "#"},

	Perl:         {Line, "", "", "#"},
	Shell:        {Line, "", "", "#"},
	PHP:          {Block, "/*", "*/", "//"},
	Swift:        {Block, "/*", "*/", "//"},
	Rust:         {Block, "/*", "*/", "//"},
	Haskell:      {Line, "", "", "--"},
	HTML:         {Block, "<!--", "-->", ""},
	XML:          {Block, "<!--", "-->", ""},
	CSS:          {Block, "/*", "*/", ""},
	SQL:          {Line, "", "", "--"},
	R:            {Line, "", "", "#"},
	Scala:        {Block, "/*", "*/", "//"},
	Kotlin:       {Block, "/*", "*/", "//"},
	TypeScript:   {Block, "/*", "*/", "//"},
	Matlab:       {Block, "%{", "%}", "%"},
	Assembly:     {Line, "", "", ";"},
	Lua:          {Block, "--[[", "--]]", "--"},
	Fortran:      {Line, "", "", "!"},
	Bash:         {Line, "", "", "#"},
	Groovy:       {Block, "/*", "*/", "//"},
	Prolog:       {Line, "", "", "%"},
	YAML:         {Line, "", "", "#"},
	Markdown:     {Line, "", "", "//"},
	Dart:         {Block, "/*", "*/", "//"},
	CoffeeScript: {Line, "", "", "#"},
	FSharp:       {Line, "", "", "//"},
	Pascal:       {Block, "{", "}", ""},
	Lisp:         {Line, "", "", ";;"},
	Erlang:       {Block, "%{", "%}", "%"},
	Elixir:       {Line, "", "", "#"},
	HCL:          {Line, "", "", "#"},
}

// GenerateComment converts a given string into a comment based on the specified programming language.
//
// **Usage**:
// Call this function with a string and a language to get the string formatted
// as a comment in that language. For example: `comment := GenerateComment("TODO: Implement this", GoLang)`.
//
// **Inner Workings**:
// The function uses the CommentDict map to determine the appropriate comment delimiters
// for the provided language. It then formats the string as a block or inline comment
// based on the language's conventions.
func GenerateComment(comment string, language Language) string {
	delimiters, exists := CommentDict[language]
	if !exists {
		return "Unsupported language"
	}

	if delimiters.Type == Line {
		lines := strings.Split(comment, "\n")
		lines_ := []string{}
		for _, line := range lines {
			// wordwrap the line
			ww := wordwrap.NewWriter(100)
			_, err := ww.Write([]byte(line))
			// if there is an error, we just stick with line
			if err != nil {
				lines_ = append(lines_, delimiters.Inline+" "+line)
				continue
			}

			_ = ww.Close()
			line = ww.String()
			for _, line := range strings.Split(line, "\n") {
				lines_ = append(lines_, delimiters.Inline+" "+line)
			}
		}
		return strings.Join(lines_, "\n")
	} else {
		return delimiters.Start + "\n" + comment + "\n" + delimiters.End
	}
}

const (
	None Language = "None"
)

var markdownToLanguageMap = map[string]Language{
	"go":           GoLang,
	"c":            C,
	"cpp":          CPP,
	"c++":          CPP,
	"java":         Java,
	"python":       Python,
	"py":           Python,
	"js":           JavaScript,
	"javascript":   JavaScript,
	"ruby":         Ruby,
	"rb":           Ruby,
	"perl":         Perl,
	"shell":        Shell,
	"sh":           Shell,
	"bash":         Bash,
	"php":          PHP,
	"swift":        Swift,
	"rust":         Rust,
	"haskell":      Haskell,
	"html":         HTML,
	"xml":          XML,
	"css":          CSS,
	"sql":          SQL,
	"r":            R,
	"scala":        Scala,
	"kotlin":       Kotlin,
	"kt":           Kotlin,
	"typescript":   TypeScript,
	"ts":           TypeScript,
	"matlab":       Matlab,
	"assembly":     Assembly,
	"asm":          Assembly,
	"lua":          Lua,
	"fortran":      Fortran,
	"f":            Fortran,
	"groovy":       Groovy,
	"prolog":       Prolog,
	"yaml":         YAML,
	"yml":          YAML,
	"md":           Markdown,
	"dart":         Dart,
	"coffeescript": CoffeeScript,
	"coffee":       CoffeeScript,
	"fsharp":       FSharp,
	"fs":           FSharp,
	"pascal":       Pascal,
	"lisp":         Lisp,
	"erlang":       Erlang,
	"elixir":       Elixir,
	"ex":           Elixir,
	"hcl":          HCL,
}

// MarkdownCodeBlockToLanguage translates a markdown code block language specifier to its corresponding Language type.
//
// **Usage**:
// Use this function to identify the language of a code block in markdown.
// For example, given the string "py", the function will return the Python constant.
func MarkdownCodeBlockToLanguage(codeBlock string) Language {
	codeBlock = strings.ToLower(codeBlock)
	language, exists := markdownToLanguageMap[codeBlock]
	if !exists {
		return None
	}
	return language
}
