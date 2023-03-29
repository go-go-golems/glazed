package cmds

import (
	"fmt"
	"github.com/go-go-golems/glazed/pkg/cli"
	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/spf13/cobra"
	"golang.org/x/net/html"
	"os"
)

func NewHTMLCommand() (*cobra.Command, error) {
	cmd := &cobra.Command{
		Use:   "html",
		Short: "HTML commands",
	}

	parseCmd := &cobra.Command{
		Use:   "parse",
		Short: "Parse HTML",
		Args:  cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			gp, of, err := cli.CreateGlazedProcessorFromCobra(cmd)
			cobra.CheckErr(err)

			for _, arg := range args {
				if arg == "-" {
					arg = "/dev/stdin"
				}
				f, err := os.Open(arg)
				cobra.CheckErr(err)
				defer f.Close()

				doc, err := html.Parse(f)
				cobra.CheckErr(err)

				err = outputNodesDepthFirst(doc, gp)
				cobra.CheckErr(err)
			}

			s, err := of.Output()
			cobra.CheckErr(err)
			if _, ok := err.(*cmds.ExitWithoutGlazeError); ok {
				os.Exit(0)
			}
			fmt.Print(s)
		},
	}

	g, err := cli.NewGlazedParameterLayers()
	if err != nil {
		return nil, err
	}
	err = g.AddFlagsToCobraCommand(parseCmd)
	if err != nil {
		return nil, err
	}

	cmd.AddCommand(parseCmd)

	return cmd, nil
}

type htmlAttribute struct {
	Key       string `json:"key"`
	Value     string `json:"value"`
	Namespace string `json:"namespace"`
}

func htmlNodeTypeToString(t html.NodeType) string {
	switch t {
	case html.ErrorNode:
		return "ErrorNode"
	case html.TextNode:
		return "TextNode"
	case html.DocumentNode:
		return "DocumentNode"
	case html.ElementNode:
		return "ElementNode"
	case html.CommentNode:
		return "CommentNode"
	case html.DoctypeNode:
		return "DoctypeNode"
	case html.RawNode:
		return "RawNode"
	default:
		return "Unknown"
	}
}

func outputNodesDepthFirst(doc *html.Node, gp *cmds.GlazeProcessor) error {
	attributes := make([]htmlAttribute, 0, len(doc.Attr))
	for _, attr := range doc.Attr {
		attributes = append(attributes, htmlAttribute{
			Key:       attr.Key,
			Value:     attr.Val,
			Namespace: attr.Namespace,
		})
	}

	obj := map[string]interface{}{
		"Type":       htmlNodeTypeToString(doc.Type),
		"Atom":       doc.DataAtom,
		"Data":       doc.Data,
		"Namespace":  doc.Namespace,
		"Attributes": attributes,
	}

	err := gp.ProcessInputObject(obj)
	if err != nil {
		return err
	}

	for c := doc.FirstChild; c != nil; c = c.NextSibling {
		err = outputNodesDepthFirst(c, gp)
		if err != nil {
			return err
		}
	}

	return nil
}
