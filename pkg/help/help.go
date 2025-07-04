package help

import (
	"io/fs"
	"path/filepath"
	"sort"
	"strings"

	"context"
	"github.com/go-go-golems/glazed/pkg/help/model"
	"github.com/go-go-golems/glazed/pkg/help/query"
	"github.com/go-go-golems/glazed/pkg/help/store"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

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
	for _, section := range hs.Sections {
		if section.Slug == slug {
			return section, nil
		}
	}
	return nil, errors.Errorf("no section with slug %s found", slug)
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

func (hs *HelpSystem) GetTopLevelHelpPage() ([]*model.Section, error) {
	pred := query.And(
		query.IsTopLevel(),
		query.Or(
			query.IsType(model.SectionGeneralTopic),
			query.IsType(model.SectionExample),
			query.IsType(model.SectionApplication),
			query.IsType(model.SectionTutorial),
		),
	)

	return hs.Store.Find(context.Background(), pred)
}

type HelpSystem struct {
	Sections []*model.Section
	Store    *store.Store
}

func NewHelpSystem() *HelpSystem {
	return &HelpSystem{
		Sections: []*model.Section{},
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
			section, err := model.LoadSectionFromMarkdown(filePath)
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
	hs.Sections = append(hs.Sections, section)
	hs.Store.UpsertSection(section)
}

func (hs *HelpSystem) SetupCobraRootCommand(cmd *cobra.Command) {
	helpFunc, usageFunc := GetCobraHelpUsageFuncs(hs)
	helpTemplate, usageTemplate := GetCobraHelpUsageTemplates(hs)

	cmd.PersistentFlags().Bool("long-help", false, "Show long help")

	cmd.SetHelpFunc(helpFunc)
	cmd.SetUsageFunc(usageFunc)
	cmd.SetHelpTemplate(helpTemplate)
	cmd.SetUsageTemplate(usageTemplate)

	helpCmd := NewCobraHelpCommand(hs)
	cmd.SetHelpCommand(helpCmd)
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
