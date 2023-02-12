package cmds

import (
	"bufio"
	"fmt"
	"github.com/go-go-golems/glazed/pkg/cli"
	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/spf13/cobra"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/text"
	"gopkg.in/errgo.v2/fmt/errors"
	"os"
	"strings"
)

type ExtensionFlag struct {
	FlagName string
	FlagDesc string
	Extender goldmark.Extender
}

// TODO(manuel, 2023-02-04) Interesting extensions to add
//
// - https://github.com/yuin/goldmark-meta (for sure!)
// - https://github.com/litao91/goldmark-mathjax
// - https://github.com/abhinav/goldmark-hashtag
// - https://github.com/abhinav/goldmark-wikilink

// I don't think these matter, but if I build a general purpose markdown CLI that might be useful
// - https://github.com/yuin/goldmark-highlighting (i don't think relevant to us, that's for HTML rendering)

// extensions: github, table, strikethrough, linkify, tasklist, definitionlist, footnote, typographer, cjk
var extensionFlags = []ExtensionFlag{
	{
		FlagName: "md-github",
		FlagDesc: "Use GitHub Flavored Markdown",
		Extender: extension.GFM,
	},
	{
		FlagName: "md-table",
		FlagDesc: "Use Markdown Tables",
		Extender: extension.Table,
	},
	{
		FlagName: "md-tasklist",
		FlagDesc: "Use Markdown Task Lists",
		Extender: extension.TaskList,
	},
	{
		FlagName: "md-strikethrough",
		FlagDesc: "Use Markdown Strikethrough",
		Extender: extension.Strikethrough,
	},
	{
		FlagName: "md-linkify",
		FlagDesc: "Use Markdown Linkify",
		Extender: extension.Linkify,
	},
	{
		FlagName: "md-definition-list",
		FlagDesc: "Use Markdown Definition Lists",
		Extender: extension.DefinitionList,
	},
	{
		FlagName: "md-footnote",
		FlagDesc: "Use Markdown Footnotes",
		Extender: extension.Footnote,
	},
	{
		FlagName: "md-typographer",
		FlagDesc: "Use Markdown Typographer",
		Extender: extension.Typographer,
	},
	{
		FlagName: "md-cjk",
		FlagDesc: "Use Markdown CJK",
		Extender: extension.CJK,
	},
}

func getExtensions(cmd *cobra.Command) ([]goldmark.Extender, error) {
	extensions := []goldmark.Extender{}
	for _, ext := range extensionFlags {
		flagValue, err := cmd.Flags().GetBool(ext.FlagName)
		if err != nil {
			return nil, err
		}
		if flagValue {
			extensions = append(extensions, ext.Extender)
		}
	}

	return extensions, nil
}

func addExtensionFlags(cmd *cobra.Command) {
	for _, ext := range extensionFlags {
		cmd.Flags().Bool(ext.FlagName, false, ext.FlagDesc)
	}
}

// What functionality to I want for a markdown command:
// - what I need right now: split on headings
// - a simple one that just prints a linearized version of the AST
// - a version with a nested DOM-like structure

type outputElement = map[string]interface{}

var MarkdownCmd = &cobra.Command{
	Use:   "markdown",
	Short: "Convert markdown data",
}

var parseCmd = &cobra.Command{
	Use:   "parse",
	Short: "Parse markdown data as AST and process further",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		gp, of, err := cli.SetupProcessor(cmd)
		if err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "Could not create glaze  processors: %v\n", err)
			os.Exit(1)
		}

		// TODO(manuel, 2023-02-04) Add support for HTML
		rendererOptions := []renderer.Option{}

		// NOTE(manuel, 2023-02-04) not sure what this really does yet lol
		extentions, err := getExtensions(cmd)
		cobra.CheckErr(err)

		md := goldmark.New(
			goldmark.WithExtensions(extentions...),
			goldmark.WithParserOptions(
				parser.WithAutoHeadingID(),
			),
			goldmark.WithRendererOptions(rendererOptions...),
		)

		parser_, _ := cmd.Flags().GetString("parser")

		// open args[0] and get reader
		for _, arg := range args {
			s, err := os.ReadFile(arg)
			cobra.CheckErr(err)

			if parser_ == "simple" {
				err = simpleLinearize(md, s, gp)
			} else if parser_ == "split" {
				err = splitByHeading(md, s, gp)
			} else {
				cobra.CheckErr(errors.Newf("unknown parser: %s", parser_))
			}
			cobra.CheckErr(err)
		}

		_ = gp

		s, err := of.Output()
		if err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "Error rendering output: %s\n", err)
			os.Exit(1)
		}
		fmt.Print(s)
	},
}

// NODE(manuel, 2023-02-04)
//
// When I split by heading, say 2, I want:
//
// so maybe the easiest is to do this using simple string parsing instead of using
// a full stack apparatus. The node walking business is definitely too much brain for me
// this morning, and I don't have the patience.
//
//		[
//	  	{
//	  	  "heading": "XXX",
//		      "body": "XXX"
//		    }
//
// ]
func splitByHeading(md goldmark.Markdown, s []byte, gp *cmds.GlazeProcessor) error {
	r := text.NewReader(s)

	// parse options are:
	// - parser.WithBlockParsers
	// - parser.WithInlineParsers
	// - parser.WithParagraphTransformers
	// - parser.WithASTTransformers
	// - parser.WithAutoHeadingID
	// - parser.WithAttribute
	node := md.Parser().Parse(r)

	parseStack := []outputElement{}
	outputStack := []outputElement{}

	// fuck my brain can't deal with stacks right now lol, i need paper
	err := ast.Walk(node, func(node ast.Node, entering bool) (ast.WalkStatus, error) {
		if entering {
			elt := outputElement{
				"kind": node.Kind().String(),
				"text": string(node.Text(s)),
			}
			switch node.Kind() {
			case ast.KindHeading:
				elt["level"] = node.(*ast.Heading).Level
				elt["children"] = []outputElement{}
			}
			parseStack = append(parseStack, elt)
		} else {
			top := parseStack[len(parseStack)-1]
			parseStack = parseStack[:len(parseStack)-1]

			// skip Text and Document
			if top["kind"] == "Text" || top["kind"] == "Document" {
				return ast.WalkContinue, nil
			}

			// when leaving, check the top of the output stack
			// if it's a heading and we are not, add ourselves to its children
			// we should end up with a list of headings, and nothing else, which we can then fold
			// in a second pass
			if len(outputStack) > 0 {
				outputTop := outputStack[len(outputStack)-1]
				if outputTop["kind"] == "Heading" && top["kind"] != "Heading" {
					outputTop["children"] = append(outputTop["children"].([]outputElement), top)
					return ast.WalkContinue, nil
				}
			}
			outputStack = append(outputStack, top)

		}
		return ast.WalkContinue, nil
	})
	if err != nil {
		return err
	}

	// fold the headings

	for _, elt := range outputStack {
		err = gp.ProcessInputObject(elt)
		if err != nil {
			return err
		}
	}

	return nil

}

// simpleLinearize is a simple walker that will linearize the blocks encountered,
// and filter out the Document and Text blocks
// to avoid duplicates, for a very simple document.
func simpleLinearize(md goldmark.Markdown, s []byte, gp *cmds.GlazeProcessor) error {
	r := text.NewReader(s)

	// parse options are:
	// - parser.WithBlockParsers
	// - parser.WithInlineParsers
	// - parser.WithParagraphTransformers
	// - parser.WithASTTransformers
	// - parser.WithAutoHeadingID
	// - parser.WithAttribute
	node := md.Parser().Parse(r)

	parseStack := []outputElement{}
	outputStack := []outputElement{}
	err := ast.Walk(node, func(node ast.Node, entering bool) (ast.WalkStatus, error) {
		if entering {
			elt := outputElement{
				"kind": node.Kind().String(),
				"text": string(node.Text(s)),
			}
			parseStack = append(parseStack, elt)
		} else {
			switch node.Kind() {
			case ast.KindDocument:
			case ast.KindText:
			default:
				top := parseStack[len(parseStack)-1]
				outputStack = append(outputStack, top)
			}
			// pop the stack
			parseStack = parseStack[:len(parseStack)-1]
		}
		return ast.WalkContinue, nil
	})
	if err != nil {
		return err
	}

	for _, elt := range outputStack {
		err = gp.ProcessInputObject(elt)
		if err != nil {
			return err
		}
	}

	return nil
}

var splitByHeadingCmd = &cobra.Command{
	Use:   "split-by-heading",
	Short: "Split a markdown file by heading",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		gp, of, err := cli.SetupProcessor(cmd)
		if err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "Could not create glaze  processors: %v\n", err)
			os.Exit(1)
		}

		level, _ := cmd.Flags().GetInt("level")
		keepEmptyHeadings, _ := cmd.Flags().GetBool("keep-empty-headings")

		// repeat # level number of times
		splitLevelString := strings.Repeat("#", level) + " "

		for _, arg := range args {
			func() {
				if arg == "-" {
					arg = "/dev/stdin"
				}
				f, err := os.Open(arg)
				cobra.CheckErr(err)
				defer f.Close()

				s := bufio.NewScanner(f)
				var currentTitle string
				var current []string

				processSection := func() {
					if len(current) == 0 && currentTitle == "" {
						currentTitle = ""
						current = []string{}
						return
					}

					if currentTitle == "" && !keepEmptyHeadings {
						currentTitle = ""
						current = []string{}
						return
					}

					row := map[string]interface{}{
						"heading": currentTitle,
						"content": strings.Trim(strings.Join(current, "\n"), " \n\t"),
					}
					err = gp.ProcessInputObject(row)
					cobra.CheckErr(err)

					currentTitle = ""
					current = []string{}

				}
				for s.Scan() {
					line := s.Text()
					if strings.HasPrefix(line, splitLevelString) {
						if len(current) > 0 {
							processSection()
						}
						currentTitle = strings.TrimSpace(line[len(splitLevelString):])
					} else {
						current = append(current, line)
					}
				}

				processSection()
			}()
		}

		_ = gp

		s, err := of.Output()
		if err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "Error rendering output: %s\n", err)
			os.Exit(1)
		}
		fmt.Print(s)

	},
}

func init() {
	parseCmd.Flags().SortFlags = false
	cli.AddFlags(parseCmd, cli.NewFlagsDefaults())
	// parser can be "simple" or "dom"
	parseCmd.Flags().StringP("parser", "t", "simple", "Type of output to generate")
	addExtensionFlags(parseCmd)
	MarkdownCmd.AddCommand(parseCmd)

	splitByHeadingCmd.Flags().SortFlags = false
	cli.AddFlags(splitByHeadingCmd, cli.NewFlagsDefaults())
	splitByHeadingCmd.Flags().Bool("keep-empty-headings", false, "Keep empty headings")
	splitByHeadingCmd.Flags().Int("level", 2, "Heading level to split by")
	MarkdownCmd.AddCommand(splitByHeadingCmd)
}
