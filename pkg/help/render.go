package help

import (
	_ "embed"
	"fmt"
	"os"
	"strings"
	"text/template"

	"github.com/go-go-golems/glazed/pkg/helpers/templating"

	"context"
	"github.com/charmbracelet/glamour"
	"github.com/go-go-golems/glazed/pkg/help/model"
	"github.com/go-go-golems/glazed/pkg/help/query"
	"github.com/go-go-golems/glazed/pkg/help/store"
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
	Predicate       query.Predicate
	Store           *store.Store
	ShowAllSections bool
	ShowShortTopic  bool
	HelpCommand     string
	LongHelp        bool
	ListSections    bool
	OnlyTopLevel    bool
}

// Add a helper function to convert []*model.Section to []*Section
func modelSectionsToSections(modelSections []*model.Section) []*Section {
	sections := make([]*Section, len(modelSections))
	for i, ms := range modelSections {
		sections[i] = &Section{
			Slug:           ms.Slug,
			SectionType:    model.SectionTypeFromStringUnsafe(string(ms.SectionType)),
			Title:          ms.Title,
			SubTitle:       ms.Subtitle,
			Short:          ms.Short,
			Content:        ms.Content,
			Topics:         ms.Topics,
			Flags:          ms.Flags,
			Commands:       ms.Commands,
			IsTopLevel:     ms.IsTopLevel,
			IsTemplate:     ms.IsTemplate,
			ShowPerDefault: ms.ShowDefault,
			Order:          ms.Ord,
		}
	}
	return sections
}

// In ComputeRenderData, use the helper to convert sections before calling NewHelpPage
func ComputeRenderData(ctx context.Context, s *store.Store, pred query.Predicate) (map[string]interface{}, bool, error) {
	modelSections, err := s.Find(ctx, pred)
	if err != nil {
		return nil, false, err
	}
	sections := modelSectionsToSections(modelSections)
	data := map[string]interface{}{}

	noResultsFound := len(sections) == 0
	// TODO: Add more query string info if needed
	data["NoResultsFound"] = noResultsFound
	data["Help"] = NewHelpPage(sections)
	return data, noResultsFound, nil
}

func RenderTopicHelp(
	ctx context.Context,
	topicSection *model.Section,
	options *RenderOptions,
) (string, error) {
	data, noResultsFound, err := ComputeRenderData(ctx, options.Store, options.Predicate)
	if err != nil {
		return "", err
	}

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
