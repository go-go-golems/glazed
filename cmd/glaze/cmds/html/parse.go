package html

import (
	"fmt"
	"github.com/go-go-golems/glazed/pkg/cmds"
	"golang.org/x/net/html"
	"strings"
)

// HTMLSplitParser is a GlazeProcessor that splits an HTML document into sections.
// When encountering one of the tags in splitTags, it extracts the content below the tag as Title
// (if extractTitle is true) and the following siblings until the next split tag is encountered as body.
type HTMLSplitParser struct {
	gp           cmds.Processor
	removeTags   []string
	splitTags    []string
	extractTitle bool
}

func NewHTMLSplitParser(gp cmds.Processor, removeTags, splitTags []string, extractTitle bool) *HTMLSplitParser {
	return &HTMLSplitParser{
		gp:           gp,
		removeTags:   removeTags,
		splitTags:    splitTags,
		extractTitle: extractTitle,
	}
}

// NewHTMLHeadingSplitParser creates a new HTMLSplitParser that splits the document into sections
// and keeps the titles, by splitting at h1, h2, h3...
func NewHTMLHeadingSplitParser(gp cmds.Processor, removeTags []string) *HTMLSplitParser {
	tags := []string{"h1", "h2", "h3", "h4", "h5", "h6"}
	removeTags = append(removeTags, tags...)
	return NewHTMLSplitParser(gp, removeTags, tags, true)
}

func (hsp *HTMLSplitParser) shouldSplit(n *html.Node) bool {
	if n.Type == html.ElementNode {
		for _, tag := range hsp.splitTags {
			if n.Data == tag {
				return true
			}
		}
	}
	return false
}

func (hsp *HTMLSplitParser) shouldRemove(n *html.Node) bool {
	if n.Type == html.ElementNode {
		for _, tag := range hsp.removeTags {
			if n.Data == tag {
				return true
			}
		}
	}
	return false
}

var depth = 0

// ProcessNode extracts the content below a header tag and sends it to the GlazeProcessor.
// It extracts the header tag content as Title, and the following siblings until the next header tag is
// encountered as body.
//
// It returns the next node to be parsed (because we need to split a certain amount of
// sibling nodes).
func (hsp *HTMLSplitParser) ProcessNode(n *html.Node) (*html.Node, error) {
	data := n.Data
	_ = data
	indent := strings.Repeat(" ", depth)

	fmt.Printf("%s Processing node %s (%s)\n", indent, n.Data, htmlNodeTypeToString(n.Type))
	depth++
	defer func() {
		depth--
		fmt.Printf("%s Done processing node %s (%s)\n", indent, n.Data, htmlNodeTypeToString(n.Type))
	}()

	next := n.NextSibling

	if n.Type == html.ElementNode && hsp.shouldSplit(n) {
		fmt.Printf("%s Processing split node %s (%s)\n", indent, n.Data, htmlNodeTypeToString(n.Type))
		var c *html.Node
		var title = ""
		var body strings.Builder

		if hsp.extractTitle {
			title = hsp.extractText(n)
		} else {
			// if we are not extracting the title, just add the section to the body
			s := hsp.extractText(n)
			body.WriteString(s)
		}

		// TODO(manuel, 2023-03-29) We should add a level attribute here

		row := map[string]interface{}{
			"Tag":   n.Data,
			"Title": title,
		}
		for c = n.NextSibling; c != nil; c = c.NextSibling {
			if c.Type == html.ElementNode && hsp.shouldSplit(c) {
				break
			}
			hsp.extractTextHelper(c, &body)
		}
		row["Body"] = strings.TrimSpace(body.String())

		err := hsp.gp.ProcessInputObject(row)
		if err != nil {
			return nil, err
		}

		next = c
	} else {
		fmt.Printf("%s Processing first child of %s (%s): %v\n", indent, n.Data, htmlNodeTypeToString(n.Type), n.FirstChild)
		if n.FirstChild != nil {
			next, err := hsp.ProcessNode(n.FirstChild)
			if err != nil {
				return nil, err
			}
			fmt.Printf("%s Done processing first child of %s (%s): next: %v\n", indent, n.Data, htmlNodeTypeToString(n.Type), next)
		}
	}

	for next != nil {
		fmt.Printf("%s Processing next sibling %s (%s)\n", indent, next.Data, htmlNodeTypeToString(next.Type))
		var err error
		fmt.Printf("%s Processing next sub node %s (%s)\n", indent, next.Data, htmlNodeTypeToString(next.Type))

		current := next
		next, err = hsp.ProcessNode(current)
		if err != nil {
			return nil, err
		}
		fmt.Printf("%s Done processing next sibling %s (%s): %v\n", indent, current.Data, htmlNodeTypeToString(current.Type), next)

		if next == nil {
			break
		}
		next = next.NextSibling
	}

	return next, nil
}

func (hsp *HTMLSplitParser) extractText(n *html.Node) string {
	var text strings.Builder
	hsp.extractTextHelper(n, &text)

	return strings.TrimSpace(text.String())
}

func (hsp *HTMLSplitParser) extractTextHelper(n *html.Node, text *strings.Builder) {
	if n.Type == html.TextNode {
		text.WriteString(n.Data)
	} else if n.Type == html.ElementNode && !hsp.shouldRemove(n) {
		text.WriteString("<" + n.Data + ">")
	}

	for c := n.FirstChild; c != nil; c = c.NextSibling {
		hsp.extractTextHelper(c, text)
	}

	if n.Type == html.ElementNode && !hsp.shouldRemove(n) {
		text.WriteString("</" + n.Data + ">")
	}
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
