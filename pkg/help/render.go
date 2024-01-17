package help

import (
	_ "embed"
	"fmt"
	"github.com/go-go-golems/glazed/pkg/helpers/templating"
	"os"
	"strings"
	"text/template"

	"github.com/charmbracelet/glamour"
	tsize "github.com/kopoli/go-terminal-size"
	"github.com/mattn/go-isatty"
)

func RenderToMarkdown(t *template.Template, data interface{}, output *os.File) (string, error) {
	sz, err := tsize.GetSize()
	if err != nil {
		sz.Width = 80
	}

	options := [](glamour.TermRendererOption){
		glamour.WithWordWrap(sz.Width),
		// If this isn't set before we set WithEnvironmentConfig or WithStandardStyle,
		// printing with colors does not appear to work.
		glamour.WithAutoStyle(),
	}

	if os.Getenv("GLAMOUR_STYLE") != "" {
		options = append(options, glamour.WithEnvironmentConfig())
	} else if !isatty.IsTerminal(output.Fd()) {
		options = append(options, glamour.WithStandardStyle("notty"))
	}

	// get markdown output
	var sb strings.Builder
	r, _ := glamour.NewTermRenderer(options...)

	err = t.Execute(&sb, data)
	if err != nil {
		return "", err
	}

	s := sb.String()
	sizeString := fmt.Sprintf("size: %dx%d\n", sz.Width, sz.Height)
	_ = sizeString

	out, err := r.Render(s)
	return out, err
}

type RenderOptions struct {
	Query           *SectionQuery
	ShowAllSections bool
	ShowShortTopic  bool
	HelpCommand     string
	LongHelp        bool
	ListSections    bool
}

func (hs *HelpSystem) ComputeRenderData(userQuery *SectionQuery) (map[string]interface{}, bool) {
	sections := userQuery.FindSections(hs.Sections)
	data := map[string]interface{}{}

	// check if the user has restricted the help to only specific commands, flags or topics
	// (this is before adding our own restriction based on the command or toplevel we are
	// going to show the help for)
	hasUserRestrictedQuery := userQuery.HasOnlyQueries()
	// Check if the user has restricted the userQuery to only certain return types
	hasUserRestrictedTypes := userQuery.HasRestrictedReturnTypes()

	if len(sections) == 0 {
		var alternativeSections []*Section

		if hasUserRestrictedQuery {
			// in this case, we should widen our userQuery to not have restrictions on commands, flags, topics
			alternativeQuery := userQuery.Clone().ResetOnlyQueries()
			alternativeSections = alternativeQuery.FindSections(hs.Sections)
		}

		if len(alternativeSections) == 0 && hasUserRestrictedTypes {
			// in this case, we should widen our userQuery to not have restrictions on return types
			alternativeQuery := userQuery.Clone().ReturnAllTypes()
			alternativeSections = alternativeQuery.FindSections(hs.Sections)
		}

		if len(alternativeSections) == 0 {
			// in this case, both the userQuery relaxation and the type relaxation don't return anything,
			// so we should show all possible options for the command / topLevel
			alternativeQuery := userQuery.Clone().ResetOnlyQueries().ReturnAllTypes()
			alternativeSections = alternativeQuery.FindSections(hs.Sections)
		}

		alternativeHelpPage := NewHelpPage(alternativeSections)
		data["Help"] = alternativeHelpPage
	} else {
		hp := NewHelpPage(sections)
		data["Help"] = hp
	}

	noResultsFound := len(sections) == 0 && (userQuery.HasOnlyQueries() || userQuery.HasRestrictedReturnTypes())

	data["NoResultsFound"] = noResultsFound
	data["QueryString"] = userQuery.GetOnlyQueryAsString()
	data["RequestedTypes"] = userQuery.GetRequestedTypesAsString()
	data["Query"] = userQuery
	data["IsTopLevel"] = userQuery.IsOnlyTopLevel()
	return data, noResultsFound
}

func (hs *HelpSystem) RenderTopicHelp(
	topicSection *Section,
	options *RenderOptions) (string, error) {
	userQuery := options.Query

	data, noResultsFound := hs.ComputeRenderData(userQuery)

	t := template.New("topic")
	t.Funcs(templating.TemplateFuncs)
	tmpl := HELP_TOPIC_TEMPLATE

	if options.ShowShortTopic {
		tmpl = HELP_SHORT_TOPIC_TEMPLATE
	}
	if options.ListSections || noResultsFound {
		tmpl = HELP_SHORT_TOPIC_TEMPLATE + HELP_LIST_TEMPLATE
	} else {
		if options.ShowAllSections {
			tmpl += HELP_LONG_SECTION_TEMPLATE
		} else {
			tmpl += HELP_SHORT_SECTION_TEMPLATE
		}
	}

	template.Must(t.Parse(tmpl))

	data["Topic"] = topicSection
	data["Slug"] = topicSection.Slug
	data["HelpCommand"] = options.HelpCommand

	s, err := RenderToMarkdown(t, data, os.Stderr)
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
