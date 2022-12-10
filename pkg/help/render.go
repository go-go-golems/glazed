package help

import (
	_ "embed"
	"fmt"
	"github.com/charmbracelet/glamour"
	"github.com/kopoli/go-terminal-size"
	"github.com/wesen/glazed/pkg/helpers"
	"strings"
	"text/template"
)

func RenderToMarkdown(t *template.Template, data map[string]interface{}) (string, error) {
	sz, err := tsize.GetSize()
	if err != nil {
		sz.Width = 80
	}

	// get markdown output
	var sb strings.Builder
	r, _ := glamour.NewTermRenderer(
		glamour.WithWordWrap(sz.Width),
		//glamour.WithAutoStyle(),
		// TODO(manuel, 2022-12-04): We need to check if we can use colors here,
		// which is not the case if we render things out to a file / pipe,
		// in the context of a redirect, or if we render to file
		glamour.WithStandardStyle("dark"),
	)

	err = t.Execute(&sb, data)

	s := sb.String()
	sizeString := fmt.Sprintf("size: %dx%d\n", sz.Width, sz.Height)
	_ = sizeString

	out, err := r.Render(s)
	return out, err
}

type RenderOptions struct {
	Query                       *SectionQuery
	ShowAllSections             bool
	ShowShortTopic              bool
	HelpCommand                 string
	ListSections                bool
	ExplictInformationRequested bool
	SomeFlagSet                 bool
}

func (hs *HelpSystem) RenderTopicHelp(
	topicSection *Section,
	options *RenderOptions) (string, error) {
	hp := NewHelpPage(options.Query.FindSections(hs.Sections))

	// TODO(manuel, 2022-12-09): we should check if we found any sections here in case a flag was set
	// if that's the case, we should probably show a list

	t := template.New("topic")
	t.Funcs(helpers.TemplateFuncs)
	tmpl := HELP_TOPIC_TEMPLATE

	if options.ShowShortTopic {
		tmpl = HELP_SHORT_TOPIC_TEMPLATE
	}
	if options.ListSections {
		tmpl = HELP_SHORT_TOPIC_TEMPLATE + HELP_LIST_TEMPLATE
	} else {
		if options.ShowAllSections {
			tmpl += HELP_LONG_SECTION_TEMPLATE
		} else {
			tmpl += HELP_SHORT_SECTION_TEMPLATE
		}
	}

	template.Must(t.Parse(tmpl))

	data := map[string]interface{}{}
	data["Topic"] = topicSection
	data["Slug"] = topicSection.Slug
	data["HelpCommand"] = options.HelpCommand
	data["Help"] = hp

	s, err := RenderToMarkdown(t, data)
	return s, err
}

//go:embed templates/help-topic.tmpl
var HELP_TOPIC_TEMPLATE string

//go:embed templates/help-short-topic.tmpl
var HELP_SHORT_TOPIC_TEMPLATE string

//go:embed templates/help-short-section-list.tmpl
var HELP_SHORT_SECTION_TEMPLATE string

//go:embed templates/help-long-section-list.tmpl
var HELP_LONG_SECTION_TEMPLATE string

//go:embed templates/help-list.tmpl
var HELP_LIST_TEMPLATE string
