package help

import (
	"bytes"
	"context"
	"fmt"
	"io/fs"
	"path/filepath"
	"sort"
	"strings"

	"github.com/adrg/frontmatter"
	"github.com/go-go-golems/glazed/pkg/help/dsl"
	"github.com/go-go-golems/glazed/pkg/help/model"
	"github.com/go-go-golems/glazed/pkg/help/store"
	strings2 "github.com/go-go-golems/glazed/pkg/helpers/strings"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
)

// Re-export types from model package for backward compatibility
type SectionType = model.SectionType

const (
	SectionGeneralTopic = model.SectionGeneralTopic
	SectionExample      = model.SectionExample
	SectionApplication  = model.SectionApplication
	SectionTutorial     = model.SectionTutorial
)

// Re-export functions from model package
var SectionTypeFromString = model.SectionTypeFromString

// Section is a structure describing an actual documentation section.
//
// This can describe:
//   - a general topic: think of this as an entry in a book
//   - an example: a way to run a certain command
//   - an application: a concrete use case for running a command. This can potentially
//     use additional external tools, multiple commands, etc. While it is nice to keep
//     these self-contained, it is not required.
//   - a tutorial: a step-by-step guide to running a command.
//
// Run `glaze help help-system` for more information.
type Section struct {
	*model.Section
	HelpSystem *HelpSystem
}

func (s *Section) IsForCommand(command string) bool {
	return strings2.StringInSlice(command, s.Commands)
}

func (s *Section) IsForFlag(flag string) bool {
	return strings2.StringInSlice(flag, s.Flags)
}

func (s *Section) IsForTopic(topic string) bool {
	return strings2.StringInSlice(topic, s.Topics)
}

// these should potentially be scoped by command

func (s *Section) DefaultGeneralTopic() []*Section {
	query := NewSectionQuery().
		ReturnTopics().
		ReturnOnlyTopics(s.Slug).
		ReturnOnlyShownByDefault().
		FilterSections(s)

	ctx := context.Background()
	results, err := query.FindSections(ctx, s.HelpSystem.Store)
	if err != nil {
		log.Warn().Err(err).Msg("Failed to query sections from store")
		return []*Section{}
	}
	return results
}

func (s *Section) DefaultExamples() []*Section {
	query := NewSectionQuery().
		ReturnExamples().
		ReturnOnlyTopics(s.Slug).
		ReturnOnlyShownByDefault().
		FilterSections(s)

	ctx := context.Background()
	results, err := query.FindSections(ctx, s.HelpSystem.Store)
	if err != nil {
		log.Warn().Err(err).Msg("Failed to query sections from store")
		return []*Section{}
	}
	return results
}

func (s *Section) OtherExamples() []*Section {
	query := NewSectionQuery().
		ReturnExamples().
		ReturnOnlyTopics(s.Slug).
		ReturnOnlyNotShownByDefault().
		FilterSections(s)

	ctx := context.Background()
	results, err := query.FindSections(ctx, s.HelpSystem.Store)
	if err != nil {
		log.Warn().Err(err).Msg("Failed to query sections from store")
		return []*Section{}
	}
	return results
}

func (s *Section) DefaultTutorials() []*Section {
	query := NewSectionQuery().
		ReturnTutorials().
		ReturnOnlyTopics(s.Slug).
		ReturnOnlyShownByDefault().
		FilterSections(s)

	ctx := context.Background()
	results, err := query.FindSections(ctx, s.HelpSystem.Store)
	if err != nil {
		log.Warn().Err(err).Msg("Failed to query sections from store")
		return []*Section{}
	}
	return results
}

func (s *Section) OtherTutorials() []*Section {
	query := NewSectionQuery().
		ReturnTutorials().
		ReturnOnlyTopics(s.Slug).
		ReturnOnlyNotShownByDefault().
		FilterSections(s)

	ctx := context.Background()
	results, err := query.FindSections(ctx, s.HelpSystem.Store)
	if err != nil {
		log.Warn().Err(err).Msg("Failed to query sections from store")
		return []*Section{}
	}
	return results
}

func (s *Section) DefaultApplications() []*Section {
	query := NewSectionQuery().
		ReturnApplications().
		ReturnOnlyTopics(s.Slug).
		ReturnOnlyShownByDefault().
		FilterSections(s)

	ctx := context.Background()
	results, err := query.FindSections(ctx, s.HelpSystem.Store)
	if err != nil {
		log.Warn().Err(err).Msg("Failed to query sections from store")
		return []*Section{}
	}
	return results
}

func (s *Section) OtherApplications() []*Section {
	query := NewSectionQuery().
		ReturnApplications().
		ReturnOnlyTopics(s.Slug).
		ReturnOnlyNotShownByDefault().
		FilterSections(s)

	ctx := context.Background()
	results, err := query.FindSections(ctx, s.HelpSystem.Store)
	if err != nil {
		log.Warn().Err(err).Msg("Failed to query sections from store")
		return []*Section{}
	}
	return results
}

func LoadSectionFromMarkdown(markdownBytes []byte) (*Section, error) {
	// get YAML metadata from markdown bytes
	//var matter struct {
	//	Name string   `yaml:"name"`
	//	Tags []string `yaml:"tags"`
	//}
	var metaData map[string]interface{}

	inputReader := bytes.NewReader(markdownBytes)
	rest, err := frontmatter.Parse(inputReader, &metaData)
	if err != nil {
		return nil, err
	}

	modelSection := &model.Section{}
	section := &Section{Section: modelSection}

	if title, ok := metaData["Title"]; ok {
		section.Title = title.(string)
	}
	if subTitle, ok := metaData["SubTitle"]; ok {
		section.SubTitle = subTitle.(string)
	}
	if short, ok := metaData["Short"]; ok {
		section.Short = short.(string)
	}

	if sectionType, ok := metaData["SectionType"]; ok {
		section.SectionType, err = SectionTypeFromString(sectionType.(string))
		if err != nil {
			return nil, err
		}
	} else {
		section.SectionType = SectionGeneralTopic
	}

	if slug := metaData["Slug"]; slug != nil {
		section.Slug = slug.(string)
	}
	section.Content = string(rest)

	if topics, ok := metaData["Topics"]; ok {
		section.Topics = strings2.InterfaceToStringList(topics)
	}

	if flags, ok := metaData["Flags"]; ok {
		section.Flags = strings2.InterfaceToStringList(flags)
	}

	if commands, ok := metaData["Commands"]; ok {
		section.Commands = strings2.InterfaceToStringList(commands)
	}

	if isTopLevel, ok := metaData["IsTopLevel"]; ok {
		section.IsTopLevel = isTopLevel.(bool)
	}

	if isTemplate, ok := metaData["IsTemplate"]; ok {
		section.IsTemplate = isTemplate.(bool)
	}

	if showPerDefault, ok := metaData["ShowPerDefault"]; ok {
		section.ShowPerDefault = showPerDefault.(bool)
	}

	if order, ok := metaData["Order"]; ok {
		section.Order = order.(int)
	}

	if section.Slug == "" || section.Title == "" {
		return nil, errors.New("missing slug or title")
	}

	return section, nil
}

// HelpPage contains all the sections related to a command
type HelpPage struct {
	DefaultGeneralTopics []*Section
	OtherGeneralTopics   []*Section
	// this is just the concatenation of default and others
	AllGeneralTopics []*Section

	DefaultExamples []*Section
	OtherExamples   []*Section
	AllExamples     []*Section

	DefaultApplications []*Section
	OtherApplications   []*Section
	AllApplications     []*Section

	DefaultTutorials []*Section
	OtherTutorials   []*Section
	AllTutorials     []*Section
}

func (hs *HelpSystem) GetSectionWithSlug(slug string) (*Section, error) {
	ctx := context.Background()
	modelSection, err := hs.Store.GetBySlug(ctx, slug)
	if err != nil {
		return nil, err
	}
	section := &Section{Section: modelSection, HelpSystem: hs}
	return section, nil
}

func NewHelpPage(sections []*Section) *HelpPage {
	ret := &HelpPage{}

	sort.Slice(sections, func(i, j int) bool {
		return sections[i].Order < sections[j].Order
	})

	for _, section := range sections {
		switch section.SectionType {
		case SectionGeneralTopic:
			if section.ShowPerDefault {
				ret.DefaultGeneralTopics = append(ret.DefaultGeneralTopics, section)
			} else {
				ret.OtherGeneralTopics = append(ret.OtherGeneralTopics, section)
			}
			ret.AllGeneralTopics = append(ret.DefaultGeneralTopics, ret.OtherGeneralTopics...)
		case SectionExample:
			if section.ShowPerDefault {
				ret.DefaultExamples = append(ret.DefaultExamples, section)
			} else {
				ret.OtherExamples = append(ret.OtherExamples, section)
			}
			ret.AllExamples = append(ret.DefaultExamples, ret.OtherExamples...)
		case SectionApplication:
			if section.ShowPerDefault {
				ret.DefaultApplications = append(ret.DefaultApplications, section)
			} else {
				ret.OtherApplications = append(ret.OtherApplications, section)
			}
			ret.AllApplications = append(ret.DefaultApplications, ret.OtherApplications...)
		case SectionTutorial:
			if section.ShowPerDefault {
				ret.DefaultTutorials = append(ret.DefaultTutorials, section)
			} else {
				ret.OtherTutorials = append(ret.OtherTutorials, section)
			}
			ret.AllTutorials = append(ret.DefaultTutorials, ret.OtherTutorials...)
		}
	}

	return ret
}

func (hs *HelpSystem) GetTopLevelHelpPage() *HelpPage {
	query := NewSectionQuery().
		ReturnOnlyTopLevel().
		ReturnAllTypes()

	ctx := context.Background()
	sections, err := query.FindSections(ctx, hs.Store)
	if err != nil {
		log.Warn().Err(err).Msg("Failed to query top level sections")
		return NewHelpPage([]*Section{})
	}
	return NewHelpPage(sections)
}

type HelpSystem struct {
	Store *store.Store // Store backend
}

func NewHelpSystem() *HelpSystem {
	st, err := store.NewInMemory()
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create in-memory store")
	}
	return &HelpSystem{
		Store: st,
	}
}

// NewHelpSystemWithStore creates a HelpSystem with store backend support
func NewHelpSystemWithStore(st *store.Store) *HelpSystem {
	return &HelpSystem{
		Store: st,
	}
}

func (hs *HelpSystem) LoadSectionsFromFS(f fs.FS, dir string) error {
	entries, err := fs.ReadDir(f, dir)
	if err != nil {
		log.Warn().Err(err).Str("dir", dir).Msg("Failed to read directory")
		return nil
	}
	for _, entry := range entries {
		filePath := filepath.Join(dir, entry.Name())
		if entry.IsDir() {
			err = hs.LoadSectionsFromFS(f, filePath)
			if err != nil {
				log.Warn().Err(err).Str("dir", filePath).Msg("Failed to load sections from directory")
				continue
			}
		} else {
			// make an explicit exception for readme.md
			if !strings.HasSuffix(entry.Name(), ".md") || strings.ToLower(entry.Name()) == "readme.md" {
				continue
			}
			b, err := fs.ReadFile(f, filePath)
			if err != nil {
				log.Warn().Err(err).Str("file", filePath).Msg("Failed to read file")
				continue
			}
			section, err := LoadSectionFromMarkdown(b)
			if err != nil {
				log.Debug().Err(err).Str("file", filePath).Msg("Failed to load section from file")
				continue
			}
			hs.AddSection(section)
		}
	}

	return nil
}

func (hs *HelpSystem) AddSection(section *Section) {
	ctx := context.Background()
	err := hs.Store.Upsert(ctx, section.Section)
	if err != nil {
		log.Warn().Err(err).Msg("Failed to store section")
	}
	section.HelpSystem = hs
}

type HelpError int

const (
	ErrSectionNotFound HelpError = iota
)

func (e HelpError) Error() string {
	switch e {
	case ErrSectionNotFound:
		return "Section not found"
	default:
		return "Unknown error"
	}
}

// PrintQueryDebug prints debug information about a query
func (hs *HelpSystem) PrintQueryDebug(queryDSL string, printQuery, printSQL bool) error {
	if printQuery {
		err := hs.printQueryAST(queryDSL)
		if err != nil {
			return err
		}
	}

	if printSQL {
		err := hs.printQuerySQL(queryDSL)
		if err != nil {
			return err
		}
	}

	return nil
}

// printQueryAST prints the parsed query AST in a readable format
func (hs *HelpSystem) printQueryAST(queryDSL string) error {
	fmt.Printf("Query: %s\n", queryDSL)

	// Parse the query into AST
	expr, err := dsl.Parse(queryDSL)
	if err != nil {
		return fmt.Errorf("failed to parse query: %w", err)
	}

	fmt.Printf("AST:\n")
	hs.printExpressionTree(expr, 0)

	return nil
}

// printQuerySQL prints the generated SQL query for debugging
func (hs *HelpSystem) printQuerySQL(queryDSL string) error {
	// Parse the query using the DSL parser
	predicate, err := dsl.ParseQuery(queryDSL)
	if err != nil {
		return fmt.Errorf("failed to parse query: %w", err)
	}

	// Create a query compiler to generate SQL
	compiler := store.NewQueryCompiler()
	predicate(compiler)

	// Build the SQL query
	sqlQuery, args := compiler.BuildQuery()

	fmt.Printf("SQL Query:\n")
	fmt.Printf("%s\n", sqlQuery)

	if len(args) > 0 {
		fmt.Printf("Parameters: %v\n", args)
	}

	return nil
}

// printExpressionTree prints the AST in a tree format
func (hs *HelpSystem) printExpressionTree(expr dsl.Expression, depth int) {
	indent := strings.Repeat("  ", depth)

	switch e := expr.(type) {
	case *dsl.BinaryExpression:
		fmt.Printf("%s%s\n", indent, e.Operator)
		fmt.Printf("%s├── ", indent)
		hs.printExpressionTree(e.Left, depth+1)
		fmt.Printf("%s└── ", indent)
		hs.printExpressionTree(e.Right, depth+1)
	case *dsl.UnaryExpression:
		fmt.Printf("%s%s\n", indent, e.Operator)
		fmt.Printf("%s└── ", indent)
		hs.printExpressionTree(e.Right, depth+1)
	case *dsl.FieldExpression:
		fmt.Printf("Field: %s = \"%s\"\n", e.Field, e.Value)
	case *dsl.TextExpression:
		fmt.Printf("Text: \"%s\"\n", e.Text)

	default:
		fmt.Printf("Unknown: %s\n", expr.String())
	}
}
