package help

import (
	"bytes"
	"embed"
	"github.com/adrg/frontmatter"
	"github.com/pkg/errors"
	"glazed/pkg/helpers"
	"io"
	"path/filepath"
	"text/template"
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

// Section is a structure describing an actual documentation section.
// This can describe:
// - a general topic: think of this as an entry in a book
// - an example: a way to run a certain command
// - an application: a concrete use case for running a command. This can potentially
//   use additional external tools, multiple commands, etc. While it is nice to keep
//   these self-contained, it is not required.
// - a tutorial: a step-by-step guide to running a command.
//
// Each section has a title, subtitle, short description and a full content.
// The slug is similar to an id and used to reference the section internally.
//
// Each section can be related to a list of topics (this would be a list of slugs
// a set of flags, and a list of commands.
//
// Some sections are shown by default. For example, when calling up the help for a command,
// the general topics,examples, applications and tutorials related to that command and that
// have the ShowPerDefault flag will be shown without further flags.
//
// Sections that don't have the ShowPerDefault flag set however will only be shown when
// explicitly asked for using the --topics --flags --examples options.
type Section struct {
	Slug        string
	SectionType SectionType
	// TODO(manuel, 2022-12-04): Potentially we want to attach a different topic name here
	// as the slug is used to look things up and it might be prettier?
	// or maybe introduce a "related topics" that can be used to look up topics to
	// attach to this section? That sounds actually like a better idea.
	//
	// If we want to attach examples to a specific section, that might better  be done
	// over a separate sectionSlugs entry, actually, instead of mixing slug and topic.

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

// SectionQuery represents a query to get different types of sections that we can pass it from the top
// so that we can for example restrict the examples of a certain general topic to the context of the command
// in which it is rendered
type SectionQuery struct {
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
	sections := GetSectionsByTypeAndTopic(s.HelpSystem.Sections, SectionGeneralTopic, s.Slug)
	sections = GetSectionsShownByDefault(sections)
	sections = FilterOutSection(sections, s)
	return sections
}

func (s *Section) DefaultExamples() []*Section {
	sections := GetSectionsByTypeAndTopic(s.HelpSystem.Sections, SectionExample, s.Slug)
	sections = GetSectionsShownByDefault(sections)
	sections = FilterOutSection(sections, s)
	return sections
}

func (s *Section) OtherExamples() []*Section {
	sections := GetSectionsByTypeAndTopic(s.HelpSystem.Sections, SectionExample, s.Slug)
	sections = GetSectionsNotShownByDefault(sections)
	sections = FilterOutSection(sections, s)
	return sections
}

func (s *Section) DefaultTutorials() []*Section {
	sections := GetSectionsByTypeAndTopic(s.HelpSystem.Sections, SectionTutorial, s.Slug)
	sections = GetSectionsShownByDefault(sections)
	sections = FilterOutSection(sections, s)
	return sections
}

func (s *Section) OtherTutorials() []*Section {
	sections := GetSectionsByTypeAndTopic(s.HelpSystem.Sections, SectionTutorial, s.Slug)
	sections = GetSectionsNotShownByDefault(sections)
	sections = FilterOutSection(sections, s)
	return sections
}

func (s *Section) DefaultApplications() []*Section {
	sections := GetSectionsByTypeAndTopic(s.HelpSystem.Sections, SectionApplication, s.Slug)
	sections = GetSectionsShownByDefault(sections)
	sections = FilterOutSection(sections, s)
	return sections
}

func (s *Section) OtherApplications() []*Section {
	sections := GetSectionsByTypeAndTopic(s.HelpSystem.Sections, SectionApplication, s.Slug)
	sections = GetSectionsNotShownByDefault(sections)
	sections = FilterOutSection(sections, s)
	return sections
}

func GetSectionsTopics(sections []*Section) []string {
	// TODO(manuel, 2022-12-04): This should be a set, and maybe sorted at the end
	// Potentially we want to show a short line for each topic, which might exist already if
	// the topic is actually a slug. Otherwise we might need to keep that information in the helpsystem.
	// this topic system needs to be fleshed out a bit more, since it's an odd mix of toplevel
	// and topic/command/flag restricted topics.
	topics := []string{}
	for _, section := range sections {
		for _, topic := range section.Topics {
			if !helpers.StringInSlice(topic, topics) {
				topics = append(topics, topic)
			}
		}
	}
	return topics
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

	return section, nil
}

// GenericHelpPage contains all the sections related to a command
//
// TODO (manuel, 2022-12-04): Not sure if we really need this, as it is all done with queries in help/cobra.go
// for now, but it might be good to centralize it here. Also move the split in Default/Others as well
type GenericHelpPage struct {
	DefaultGeneralTopics []*Section
	OtherGeneralTopics   []*Section
	DefaultExamples      []*Section
	OtherExamples        []*Section
	DefaultApplications  []*Section
	OtherApplications    []*Section
	DefaultTutorials     []*Section
	OtherTutorials       []*Section
}

func (hs *HelpSystem) GetCommandHelpPage(command string) *GenericHelpPage {
	sections := GetSectionsForCommand(hs.Sections, command)
	return NewHelpPage(sections)
}

func NewHelpPage(sections []*Section) *GenericHelpPage {
	ret := &GenericHelpPage{}

	generalTopics := GetSectionsByType(sections, SectionGeneralTopic)
	ret.DefaultGeneralTopics = GetSectionsShownByDefault(generalTopics)
	ret.OtherGeneralTopics = GetSectionsNotShownByDefault(generalTopics)

	examples := GetSectionsByType(sections, SectionExample)
	ret.DefaultExamples = GetSectionsShownByDefault(examples)
	ret.OtherExamples = GetSectionsNotShownByDefault(examples)

	applications := GetSectionsByType(sections, SectionApplication)
	ret.DefaultApplications = GetSectionsShownByDefault(applications)
	ret.OtherApplications = GetSectionsNotShownByDefault(applications)

	tutorials := GetSectionsByType(sections, SectionTutorial)
	ret.DefaultTutorials = GetSectionsShownByDefault(tutorials)
	ret.OtherTutorials = GetSectionsNotShownByDefault(tutorials)

	return ret
}

func (hs *HelpSystem) GetTopLevelHelpPage() *GenericHelpPage {
	return NewHelpPage(GetTopLevelSections(hs.Sections))
}

func (hs *HelpSystem) RenderSectionSummaries(w io.Writer, sections []*Section) error {
	return nil
}

func (hs *HelpSystem) RenderTopic(w io.Writer, section *Section) error {
	return nil
}

type HelpSystem struct {
	Sections []*Section

	// TODO(manuel, 2022-12-04): I don't think this is needed actually
	SectionsByFlag    map[string][]*Section
	SectionsByCommand map[string][]*Section
}

func NewHelpSystem() *HelpSystem {
	return &HelpSystem{
		Sections:          []*Section{},
		SectionsByFlag:    map[string][]*Section{},
		SectionsByCommand: map[string][]*Section{},
	}
}

func (hs *HelpSystem) LoadSectionsFromEmbedFS(f embed.FS, dir string) error {
	entries, err := f.ReadDir(dir)
	if err != nil {
		return err
	}
	for _, entry := range entries {
		if entry.IsDir() {
			err = hs.LoadSectionsFromEmbedFS(f, filepath.Join(dir, entry.Name()))
			if err != nil {
				return err
			}
		} else {
			b, err := f.ReadFile(filepath.Join(dir, entry.Name()))
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

func (hs *HelpSystem) AddSection(section *Section) {
	hs.Sections = append(hs.Sections, section)
	for _, flag := range section.Flags {
		if hs.SectionsByFlag[flag] == nil {
			hs.SectionsByFlag[flag] = []*Section{}
		}
		hs.SectionsByFlag[flag] = append(hs.SectionsByFlag[flag], section)
	}
	for _, command := range section.Commands {
		if hs.SectionsByCommand[command] == nil {
			hs.SectionsByCommand[command] = []*Section{}
		}
		hs.SectionsByCommand[command] = append(hs.SectionsByCommand[command], section)
	}
	section.HelpSystem = hs
}

func FilterOutSection(sections []*Section, section *Section) []*Section {
	filtered := []*Section{}
	for _, s := range sections {
		if s != section {
			filtered = append(filtered, s)
		}
	}
	return filtered
}

func GetSectionsBySlug(sections []*Section, slug string) []*Section {
	filtered := []*Section{}
	for _, s := range sections {
		if s.Slug == slug {
			filtered = append(filtered, s)
		}
	}
	return filtered
}

func GetSectionsByType(sections []*Section, sectionType SectionType) []*Section {
	ret := []*Section{}
	for _, section := range sections {
		if section.SectionType == sectionType {
			ret = append(ret, section)
		}
	}
	return ret
}

func GetSectionsByTopic(sections []*Section, topic string) []*Section {
	ret := []*Section{}
	for _, section := range sections {
		if section.IsForTopic(topic) {
			ret = append(ret, section)
		}
	}
	return ret
}

func GetSectionsByTypeAndTopic(sections []*Section, sectiontype SectionType, topic string) []*Section {
	ret := []*Section{}
	for _, section := range sections {
		if section.SectionType == sectiontype && section.IsForTopic(topic) {
			ret = append(ret, section)
		}
	}
	return ret
}

func GetTopLevelSections(sections []*Section) []*Section {
	ret := []*Section{}
	for _, section := range sections {
		if section.IsTopLevel {
			ret = append(ret, section)
		}
	}
	return ret
}

func GetSectionsShownByDefault(sections []*Section) []*Section {
	ret := []*Section{}
	for _, section := range sections {
		if section.ShowPerDefault {
			ret = append(ret, section)
		}
	}
	return ret
}

func GetSectionsNotShownByDefault(sections []*Section) []*Section {
	ret := []*Section{}
	for _, section := range sections {
		if !section.ShowPerDefault {
			ret = append(ret, section)
		}
	}
	return ret
}

func GetSectionsForCommand(sections []*Section, command string) []*Section {
	ret := []*Section{}
	for _, section := range sections {
		if section.IsForCommand(command) {
			ret = append(ret, section)
		}
	}
	return ret
}

func GetSectionsForFlag(sections []*Section, flag string) []*Section {
	ret := []*Section{}
	for _, section := range sections {
		if section.IsForFlag(flag) {
			ret = append(ret, section)
		}
	}
	return ret
}

func GetSectionsByTypeAndCommand(sections []*Section, sectiontype SectionType, command string) []*Section {
	return GetSectionsByType(GetSectionsForCommand(sections, command), sectiontype)
}

func GetSectionsByTypeCommandAndFlag(sections []*Section, sectiontype SectionType, command string, flag string) []*Section {
	return GetSectionsByType(GetSectionsForFlag(GetSectionsForCommand(sections, command), flag), sectiontype)
}

// TODO(manuel, 2022-12-04): This is all a placeholder for now
func (s *Section) Render(w io.Writer, data interface{}) error {
	renderedTitle := s.Title
	if s.IsTemplate {
		t := template.New("title")
		template.Must(t.Parse(s.Title))
		var titleBuffer bytes.Buffer
		err := t.Execute(&titleBuffer, data)
		if err != nil {
			return err
		}
		renderedTitle = titleBuffer.String()
	}
	_, err := w.Write([]byte(renderedTitle))
	if err != nil {
		return err
	}

	if s.IsTemplate {
		t := template.New("content")
		template.Must(t.Parse(s.Content))
		err := t.Execute(w, data)
		if err != nil {
			return err
		}
	} else {
		_, err := w.Write([]byte(s.Content))
		if err != nil {
			return err
		}
	}

	return nil
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
