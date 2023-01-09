package help

import (
	"bytes"
	"embed"
	"fmt"
	"github.com/adrg/frontmatter"
	"github.com/pkg/errors"
	"github.com/wesen/glazed/pkg/helpers"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type SectionType int

const (
	SectionGeneralTopic SectionType = iota
	SectionExample
	SectionApplication
	SectionTutorial
)

func SectionTypeFromString(s string) (SectionType, error) {
	switch s {
	case "GeneralTopic":
		return SectionGeneralTopic, nil
	case "Example":
		return SectionExample, nil
	case "Application":
		return SectionApplication, nil
	case "Tutorial":
		return SectionTutorial, nil
	}
	return SectionGeneralTopic, errors.Errorf("unknown section type %s", s)
}

func (s SectionType) String() string {
	switch s {
	case SectionGeneralTopic:
		return "GeneralTopic"
	case SectionExample:
		return "Example"
	case SectionApplication:
		return "Application"
	case SectionTutorial:
		return "Tutorial"
	}
	return "Unknown"
}

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
	Slug        string
	SectionType SectionType

	Title    string
	SubTitle string
	Short    string
	Content  string

	// metadata used to search and select sections to be shown
	Topics   []string
	Flags    []string
	Commands []string

	// Show this section in the toplevel help
	IsTopLevel bool

	IsTemplate bool

	// show this template as a default example
	ShowPerDefault bool

	// Used to give some rough sense of order, not sure how useful this is going to be
	Order int

	HelpSystem *HelpSystem
}

func (s *Section) IsForCommand(command string) bool {
	return helpers.StringInSlice(command, s.Commands)
}

func (s *Section) IsForFlag(flag string) bool {
	return helpers.StringInSlice(flag, s.Flags)
}

func (s *Section) IsForTopic(topic string) bool {
	return helpers.StringInSlice(topic, s.Topics)
}

// these should potentially be scoped by command

func (s *Section) DefaultGeneralTopic() []*Section {
	return NewSectionQuery().
		ReturnTopics().
		ReturnOnlyTopics(s.Slug).
		ReturnOnlyShownByDefault().
		FilterSections(s).
		FindSections(s.HelpSystem.Sections)
}

func (s *Section) DefaultExamples() []*Section {
	return NewSectionQuery().
		ReturnExamples().
		ReturnOnlyTopics(s.Slug).
		ReturnOnlyShownByDefault().
		FilterSections(s).
		FindSections(s.HelpSystem.Sections)
}

func (s *Section) OtherExamples() []*Section {
	return NewSectionQuery().
		ReturnExamples().
		ReturnOnlyTopics(s.Slug).
		ReturnOnlyNotShownByDefault().
		FilterSections(s).
		FindSections(s.HelpSystem.Sections)
}

func (s *Section) DefaultTutorials() []*Section {
	return NewSectionQuery().
		ReturnTutorials().
		ReturnOnlyTopics(s.Slug).
		ReturnOnlyShownByDefault().
		FilterSections(s).
		FindSections(s.HelpSystem.Sections)
}

func (s *Section) OtherTutorials() []*Section {
	return NewSectionQuery().
		ReturnTutorials().
		ReturnOnlyTopics(s.Slug).
		ReturnOnlyNotShownByDefault().
		FilterSections(s).
		FindSections(s.HelpSystem.Sections)
}

func (s *Section) DefaultApplications() []*Section {
	return NewSectionQuery().
		ReturnApplications().
		ReturnOnlyTopics(s.Slug).
		ReturnOnlyShownByDefault().
		FilterSections(s).
		FindSections(s.HelpSystem.Sections)
}

func (s *Section) OtherApplications() []*Section {
	return NewSectionQuery().
		ReturnApplications().
		ReturnOnlyTopics(s.Slug).
		ReturnOnlyNotShownByDefault().
		FilterSections(s).
		FindSections(s.HelpSystem.Sections)
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

	section := &Section{}

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
		section.Topics = helpers.InterfaceToStringList(topics)
	}

	if flags, ok := metaData["Flags"]; ok {
		section.Flags = helpers.InterfaceToStringList(flags)
	}

	if commands, ok := metaData["Commands"]; ok {
		section.Commands = helpers.InterfaceToStringList(commands)
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
		return nil, fmt.Errorf("missing slug or title")
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
	for _, section := range hs.Sections {
		if section.Slug == slug {
			return section, nil
		}
	}
	return nil, fmt.Errorf("no section with slug %s found", slug)
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
	sections := NewSectionQuery().
		ReturnOnlyTopLevel().
		ReturnAllTypes().
		FindSections(hs.Sections)
	return NewHelpPage(sections)
}

type HelpSystem struct {
	Sections []*Section
}

func NewHelpSystem() *HelpSystem {
	return &HelpSystem{
		Sections: []*Section{},
	}
}

func (hs *HelpSystem) LoadSectionsFromDirectory(dir string) error {
	files, err := os.ReadDir(dir)
	if err != nil {
		return err
	}
	for _, entry := range files {
		filePath := filepath.Join(dir, entry.Name())
		if entry.IsDir() {
			err := hs.LoadSectionsFromDirectory(filePath)
			if err != nil {
				return err
			}
		} else {
			if !strings.HasSuffix(entry.Name(), ".md") {
				continue
			}
			b, err := os.ReadFile(filePath)
			if err != nil {
				return err
			}
			section, err := LoadSectionFromMarkdown(b)
			if err != nil {
				return err
			}
			hs.AddSection(section)
		}
	}

	return nil
}

func (hs *HelpSystem) LoadSectionsFromEmbedFS(f embed.FS, dir string) error {
	entries, err := f.ReadDir(dir)
	if err != nil {
		return err
	}
	for _, entry := range entries {
		fileName := filepath.Join(dir, entry.Name())
		if entry.IsDir() {
			err = hs.LoadSectionsFromEmbedFS(f, fileName)
			if err != nil {
				return err
			}
		} else {
			if !strings.HasSuffix(fileName, ".md") {
				continue
			}
			b, err := f.ReadFile(fileName)
			if err != nil {
				return errors.Wrapf(err, "failed to read file %s", fileName)
			}
			section, err := LoadSectionFromMarkdown(b)
			if err != nil {
				return errors.Wrapf(err, "failed to load section from file %s", fileName)
			}
			hs.AddSection(section)
		}
	}

	return nil
}

func (hs *HelpSystem) AddSection(section *Section) {
	hs.Sections = append(hs.Sections, section)
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
