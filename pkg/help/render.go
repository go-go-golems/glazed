package help

import (
	"context"
	_ "embed"
	"io"
	"os"
	"strings"
	"text/template"

	"github.com/charmbracelet/glamour"
	"github.com/go-go-golems/glazed/pkg/help/model"
	"github.com/go-go-golems/glazed/pkg/help/store"
	"github.com/go-go-golems/glazed/pkg/helpers/templating"
	tsize "github.com/kopoli/go-terminal-size"
	"github.com/mattn/go-isatty"
)

type fdWriter interface {
	Fd() uintptr
}

func RenderToMarkdown(t *template.Template, data interface{}, output io.Writer) (string, error) {
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
	} else {
		if f, ok := output.(fdWriter); ok {
			if !isatty.IsTerminal(f.Fd()) {
				options = append(options, glamour.WithStandardStyle("notty"))
			}
		} else {
			options = append(options, glamour.WithStandardStyle("notty"))
		}
	}

	var sb strings.Builder
	r, _ := glamour.NewTermRenderer(options...)

	err = t.Execute(&sb, data)
	if err != nil {
		return "", err
	}

	s := sb.String()
	out, err := r.Render(s)
	return out, err
}

type RenderOptions struct {
	Predicate             store.Predicate
	RelaxNoQueryPredicate store.Predicate
	RelaxNoTypesPredicate store.Predicate
	RelaxBroadPredicate   store.Predicate
	HasOnlyQueries        bool
	HasRestrictedTypes    bool
	QueryString           string
	RequestedTypes        string
	ShowAllSections       bool
	ShowShortTopic        bool
	HelpCommand           string
	LongHelp              bool
	ListSections          bool
	OnlyTopLevel          bool
	ShowDocumentationList bool
}

func predicateOrDefault(pred store.Predicate) store.Predicate {
	if pred == nil {
		return store.OrderByOrder()
	}
	return store.And(pred, store.OrderByOrder())
}

func (hs *HelpSystem) findWithPredicate(pred store.Predicate) ([]*model.Section, error) {
	ctx := context.Background()
	return hs.Store.Find(ctx, predicateOrDefault(pred))
}

func (hs *HelpSystem) ComputeRenderData(options *RenderOptions) (map[string]interface{}, bool) {
	sections, err := hs.findWithPredicate(options.Predicate)
	if err != nil {
		sections = []*model.Section{}
	}
	data := map[string]interface{}{}

	if len(sections) == 0 {
		var alternativeSections []*model.Section

		if options.HasOnlyQueries && options.RelaxNoQueryPredicate != nil {
			alternativeSections, err = hs.findWithPredicate(options.RelaxNoQueryPredicate)
			if err != nil {
				alternativeSections = []*model.Section{}
			}
		}

		if len(alternativeSections) == 0 && options.HasRestrictedTypes && options.RelaxNoTypesPredicate != nil {
			alternativeSections, err = hs.findWithPredicate(options.RelaxNoTypesPredicate)
			if err != nil {
				alternativeSections = []*model.Section{}
			}
		}

		if len(alternativeSections) == 0 && options.RelaxBroadPredicate != nil {
			alternativeSections, err = hs.findWithPredicate(options.RelaxBroadPredicate)
			if err != nil {
				alternativeSections = []*model.Section{}
			}
		}

		data["Help"] = NewHelpPage(alternativeSections)
	} else {
		data["Help"] = NewHelpPage(sections)
	}

	noResultsFound := len(sections) == 0 && (options.HasOnlyQueries || options.HasRestrictedTypes)

	data["NoResultsFound"] = noResultsFound
	data["QueryString"] = options.QueryString
	data["RequestedTypes"] = options.RequestedTypes
	data["IsTopLevel"] = options.OnlyTopLevel
	return data, noResultsFound
}

func (hs *HelpSystem) RenderTopicHelp(
	topicSection *model.Section,
	options *RenderOptions) (string, error) {
	return hs.RenderTopicHelpWithWriter(topicSection, options, os.Stdout)
}

// RenderTopicHelpWithWriter renders a topic's help content using the provided writer
// to detect terminal characteristics when applying Glamour styles.
func (hs *HelpSystem) RenderTopicHelpWithWriter(
	topicSection *model.Section,
	options *RenderOptions,
	output io.Writer,
) (string, error) {
	data, noResultsFound := hs.ComputeRenderData(options)

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

	s, err := RenderToMarkdown(t, data, output)
	return s, err
}

//go:embed cmd/templates/help-topic.tmpl
var HELP_TOPIC_TEMPLATE string

//go:embed cmd/templates/help-short-topic.tmpl
var HELP_SHORT_TOPIC_TEMPLATE string

//go:embed cmd/templates/help-short-section-list.tmpl
var HELP_SHORT_SECTION_TEMPLATE string

//go:embed cmd/templates/help-long-section-list.tmpl
var HELP_LONG_SECTION_TEMPLATE string

//go:embed cmd/templates/help-list.tmpl
var HELP_LIST_TEMPLATE string
