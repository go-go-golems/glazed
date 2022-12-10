package help

import "strings"

// SectionQuery represents a query to get different types of sections.
//
// This is used for example by the `help` command line function
// to render out the help sections for individual commands.
//
// It can however also be used on its own.
type SectionQuery struct {
	OnlyShownByDefault    bool
	OnlyNotShownByDefault bool
	OnlyTopLevel          bool

	// only these types will be returned
	Types []SectionType

	// if any of these is set, and they match each of the Only types,
	// the section will be return
	Topics   []string
	Flags    []string
	Commands []string
	Slugs    []string

	// this will return any section as long as it matches the Only strings
	All bool

	// a section will be returned only if it matches all
	// of the following criteria
	OnlyTopics   []string
	OnlyFlags    []string
	OnlyCommands []string
	OnlySlugs    []string

	// We often need to filter sections that have already been shown
	WithoutSections []*Section
}

func NewSectionQuery() *SectionQuery {
	return &SectionQuery{
		All: true,
	}
}

func (s *SectionQuery) ReturnAllTypes() *SectionQuery {
	return s.ReturnTopics().ReturnExamples().ReturnApplications().ReturnTutorials()
}

func (s *SectionQuery) ReturnTopics() *SectionQuery {
	s.Types = append(s.Types, SectionGeneralTopic)
	return s
}

func (s *SectionQuery) ReturnExamples() *SectionQuery {
	s.Types = append(s.Types, SectionExample)
	return s
}

func (s *SectionQuery) ReturnApplications() *SectionQuery {
	s.Types = append(s.Types, SectionApplication)
	return s
}

func (s *SectionQuery) ReturnTutorials() *SectionQuery {
	s.Types = append(s.Types, SectionTutorial)
	return s
}

func (s *SectionQuery) ReturnOnlyShownByDefault() *SectionQuery {
	s.OnlyShownByDefault = true
	return s
}

func (s *SectionQuery) ReturnOnlyNotShownByDefault() *SectionQuery {
	s.OnlyNotShownByDefault = true
	return s
}

func (s *SectionQuery) ReturnOnlyTopLevel() *SectionQuery {
	s.OnlyTopLevel = true
	return s
}

func (s *SectionQuery) FilterSections(sections ...*Section) *SectionQuery {
	s.WithoutSections = sections
	return s
}

func (s *SectionQuery) ReturnAnyOfTopics(topics ...string) *SectionQuery {
	s.All = false
	s.Topics = topics
	return s
}

func (s *SectionQuery) ReturnAnyOfFlags(flags ...string) *SectionQuery {
	s.All = false
	s.Flags = flags
	return s
}

func (s *SectionQuery) ReturnAnyOfCommands(commands ...string) *SectionQuery {
	s.All = false
	s.Commands = commands
	return s
}

func (s *SectionQuery) ReturnAnyOfSlugs(slugs ...string) *SectionQuery {
	s.All = false
	s.Slugs = slugs
	return s
}

func (s *SectionQuery) ReturnOnlyTopics(topics ...string) *SectionQuery {
	s.OnlyTopics = topics
	return s
}

func (s *SectionQuery) ReturnOnlyFlags(flags ...string) *SectionQuery {
	s.OnlyFlags = flags
	return s
}

func (s *SectionQuery) ReturnOnlyCommands(commands ...string) *SectionQuery {
	s.OnlyCommands = commands
	return s
}

func (s *SectionQuery) ReturnOnlySlugs(slugs ...string) *SectionQuery {
	s.OnlySlugs = slugs
	return s
}

func (s *SectionQuery) HasOnlyQueries() bool {
	return len(s.OnlyTopics) > 0 || len(s.OnlyFlags) > 0 || len(s.OnlyCommands) > 0 || len(s.OnlySlugs) > 0
}

func (s *SectionQuery) GetOnlyQueryAsString() string {
	sb := strings.Builder{}
	if len(s.OnlySlugs) > 0 {
		sb.WriteString(" sections: ")
		sb.WriteString(strings.Join(s.OnlySlugs, ","))
	}
	if len(s.OnlyTopics) > 0 {
		sb.WriteString(" topics: ")
		sb.WriteString(strings.Join(s.OnlyTopics, ","))
	}
	if len(s.OnlyFlags) > 0 {
		sb.WriteString(" flags: ")
		sb.WriteString(strings.Join(s.OnlyFlags, ","))
	}
	if len(s.OnlyCommands) > 0 {
		sb.WriteString(" commands: ")
		sb.WriteString(strings.Join(s.OnlyCommands, ","))
	}

	return sb.String()
}

func (s *SectionQuery) ResetOnlyQueries() *SectionQuery {
	s.OnlyTopics = []string{}
	s.OnlyFlags = []string{}
	s.OnlyCommands = []string{}
	s.OnlySlugs = []string{}
	return s
}

func (q *SectionQuery) Clone() *SectionQuery {
	return &SectionQuery{
		OnlyShownByDefault:    q.OnlyShownByDefault,
		OnlyNotShownByDefault: q.OnlyNotShownByDefault,
		OnlyTopLevel:          q.OnlyTopLevel,
		Types:                 q.Types,
		Topics:                q.Topics,
		Flags:                 q.Flags,
		Commands:              q.Commands,
		Slugs:                 q.Slugs,
		All:                   q.All,
		OnlyTopics:            q.OnlyTopics,
		OnlyFlags:             q.OnlyFlags,
		OnlyCommands:          q.OnlyCommands,
		OnlySlugs:             q.OnlySlugs,
		WithoutSections:       q.WithoutSections,
	}
}

func (q *SectionQuery) FindSections(sections []*Section) []*Section {
	var result []*Section

sectionLoop:
	for _, section := range sections {
		if q.OnlyShownByDefault && !section.ShowPerDefault {
			continue
		}

		if q.OnlyNotShownByDefault && section.ShowPerDefault {
			continue
		}

		if q.OnlyTopLevel && !section.IsTopLevel {
			continue
		}

		for _, without := range q.WithoutSections {
			if without == section {
				continue sectionLoop
			}
		}

		foundMatchingType := false
		for _, t := range q.Types {
			if section.SectionType == t {
				foundMatchingType = true
				break
			}
		}
		if !foundMatchingType {
			continue
		}

		// filter out the Only*
		if len(q.OnlyTopics) > 0 {
			foundMatchingTopic := true
			for _, t := range q.OnlyTopics {
				if !section.IsForTopic(t) {
					foundMatchingTopic = false
					break
				}
			}
			if !foundMatchingTopic {
				continue sectionLoop
			}
		}

		if len(q.OnlyFlags) > 0 {
			foundMatchingFlag := true
			for _, f := range q.OnlyFlags {
				if !section.IsForFlag(f) {
					foundMatchingFlag = false
					break
				}
			}
			if !foundMatchingFlag {
				continue sectionLoop
			}
		}

		if len(q.OnlyCommands) > 0 {
			foundMatchingCommand := true
			for _, c := range q.OnlyCommands {
				if !section.IsForCommand(c) {
					foundMatchingCommand = false
					break
				}
			}
			if !foundMatchingCommand {
				continue sectionLoop
			}
		}

		if len(q.OnlySlugs) > 0 {
			foundMatchingSlug := true
			for _, s := range q.OnlySlugs {
				if section.Slug != s {
					foundMatchingSlug = true
					break
				}
			}
			if !foundMatchingSlug {
				continue sectionLoop
			}
		}

		if q.All {
			result = append(result, section)
			continue sectionLoop
		}

		for _, topic := range q.Topics {
			if section.IsForTopic(topic) {
				result = append(result, section)
				continue sectionLoop
			}
		}
		for _, flag := range q.Flags {
			if section.IsForFlag(flag) {
				result = append(result, section)
				continue sectionLoop
			}
		}
		for _, command := range q.Commands {
			if section.IsForCommand(command) {
				result = append(result, section)
				continue sectionLoop
			}
		}
		for _, slug := range q.Slugs {
			if section.Slug == slug {
				result = append(result, section)
				continue sectionLoop
			}
		}
	}

	return result
}
