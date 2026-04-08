package help

import (
	"context"
	"fmt"
	"io/fs"
	"path/filepath"
	"sort"
	"strings"

	"github.com/go-go-golems/glazed/pkg/help/dsl"
	"github.com/go-go-golems/glazed/pkg/help/model"
	"github.com/go-go-golems/glazed/pkg/help/store"
	"github.com/rs/zerolog/log"
)

func LoadSectionFromMarkdown(markdownBytes []byte) (*model.Section, error) {
	return model.ParseSectionFromMarkdown(markdownBytes)
}

// HelpPage contains all the sections related to a command
type HelpPage struct {
	DefaultGeneralTopics []*model.Section
	OtherGeneralTopics   []*model.Section
	// this is just the concatenation of default and others
	AllGeneralTopics []*model.Section

	DefaultExamples []*model.Section
	OtherExamples   []*model.Section
	AllExamples     []*model.Section

	DefaultApplications []*model.Section
	OtherApplications   []*model.Section
	AllApplications     []*model.Section

	DefaultTutorials []*model.Section
	OtherTutorials   []*model.Section
	AllTutorials     []*model.Section
}

func (hs *HelpSystem) GetSectionWithSlug(slug string) (*model.Section, error) {
	ctx := context.Background()
	return hs.Store.GetBySlug(ctx, slug)
}

func NewHelpPage(sections []*model.Section) *HelpPage {
	ret := &HelpPage{}

	sort.Slice(sections, func(i, j int) bool {
		return sections[i].Order < sections[j].Order
	})

	for _, section := range sections {
		switch section.SectionType {
		case model.SectionGeneralTopic:
			if section.ShowPerDefault {
				ret.DefaultGeneralTopics = append(ret.DefaultGeneralTopics, section)
			} else {
				ret.OtherGeneralTopics = append(ret.OtherGeneralTopics, section)
			}
			ret.AllGeneralTopics = append(ret.DefaultGeneralTopics, ret.OtherGeneralTopics...)
		case model.SectionExample:
			if section.ShowPerDefault {
				ret.DefaultExamples = append(ret.DefaultExamples, section)
			} else {
				ret.OtherExamples = append(ret.OtherExamples, section)
			}
			ret.AllExamples = append(ret.DefaultExamples, ret.OtherExamples...)
		case model.SectionApplication:
			if section.ShowPerDefault {
				ret.DefaultApplications = append(ret.DefaultApplications, section)
			} else {
				ret.OtherApplications = append(ret.OtherApplications, section)
			}
			ret.AllApplications = append(ret.DefaultApplications, ret.OtherApplications...)
		case model.SectionTutorial:
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
	ctx := context.Background()
	sections, err := hs.Store.Find(ctx, store.IsTopLevel())
	if err != nil {
		log.Warn().Err(err).Msg("Failed to query top level sections")
		return NewHelpPage([]*model.Section{})
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
			section, err := model.ParseSectionFromMarkdown(b)
			if err != nil {
				log.Debug().Err(err).Str("file", filePath).Msg("Failed to load section from file")
				continue
			}
			hs.AddSection(section)
		}
	}

	return nil
}

func (hs *HelpSystem) AddSection(section *model.Section) {
	ctx := context.Background()
	err := hs.Store.Upsert(ctx, section)
	if err != nil {
		log.Warn().Err(err).Msg("Failed to store section")
	}
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
		fmt.Printf("Arguments: %v\n", args)
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
