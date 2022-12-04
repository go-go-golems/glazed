package help

import (
	"bytes"
	"fmt"
	"glazed/pkg/helpers"
	"io"
	"text/template"
)

type SectionType int

const (
	SectionGeneralTopic SectionType = iota
	SectionExample
	SectionApplication
	SectionTutorial
)

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

type CommandHelpPage struct {
	GeneralTopics []*Section
	Examples      []*Section
	Applications  []*Section
	Tutorials     []*Section
}

type HelpSystem struct {
	Sections []*Section

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

type RenderContext struct {
	Depth int
	Tags  []string
	Data  interface{}
}

func (rc *RenderContext) AddTag(tag string) *RenderContext {
	newrc := &RenderContext{
		Depth: rc.Depth,
		Tags:  append(rc.Tags, tag),
	}
	return newrc
}

func (rc *RenderContext) IsTaggedWithAny(tags []string) bool {
	for _, t := range tags {
		if rc.IsTagged(t) {
			return true
		}
	}
	return false
}

func (rc *RenderContext) IsTagged(tag string) bool {
	return rc.IsTaggedWithAny([]string{tag})
}

func (rc *RenderContext) IncreaseDepth() *RenderContext {
	newrc := &RenderContext{
		Depth: rc.Depth + 1,
		Tags:  rc.Tags,
	}
	return newrc
}

func makeMarkdownHeader(depth int, title string) string {
	return fmt.Sprintf("%s %s\n\n", bytes.Repeat([]byte("#"), depth), title)
}

type ContentSection interface {
	Render(w io.Writer, rc *RenderContext) error
}

// TODO(manuel, 2022-12-04): This is all a placeholder for now
func (s *Section) Render(w io.Writer, rc *RenderContext) error {
	renderedTitle := s.Title
	if s.IsTemplate {
		t := template.New("title")
		template.Must(t.Parse(s.Title))
		var titleBuffer bytes.Buffer
		err := t.Execute(&titleBuffer, rc.Data)
		if err != nil {
			return err
		}
		renderedTitle = titleBuffer.String()
	}
	_, err := w.Write([]byte(makeMarkdownHeader(rc.Depth, renderedTitle)))
	if err != nil {
		return err
	}

	if s.IsTemplate {
		t := template.New("content")
		template.Must(t.Parse(s.Content))
		err := t.Execute(w, rc.Data)
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
