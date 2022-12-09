package help

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

type QueryBuilder struct {
	SectionQuery *SectionQuery
}

func NewQueryBuilder() *QueryBuilder {
	return &QueryBuilder{
		SectionQuery: &SectionQuery{
			All: true,
		},
	}
}

func (b *QueryBuilder) ReturnAllTypes() *QueryBuilder {
	return b.ReturnTopics().ReturnExamples().ReturnApplications().ReturnTutorials()
}

func (b *QueryBuilder) ReturnTopics() *QueryBuilder {
	b.SectionQuery.Types = append(b.SectionQuery.Types, SectionGeneralTopic)
	return b
}

func (b *QueryBuilder) ReturnExamples() *QueryBuilder {
	b.SectionQuery.Types = append(b.SectionQuery.Types, SectionExample)
	return b
}

func (b *QueryBuilder) ReturnApplications() *QueryBuilder {
	b.SectionQuery.Types = append(b.SectionQuery.Types, SectionApplication)
	return b
}

func (b *QueryBuilder) ReturnTutorials() *QueryBuilder {
	b.SectionQuery.Types = append(b.SectionQuery.Types, SectionTutorial)
	return b
}

func (b *QueryBuilder) OnlyShownByDefault() *QueryBuilder {
	b.SectionQuery.OnlyShownByDefault = true
	return b
}

func (b *QueryBuilder) OnlyNotShownByDefault() *QueryBuilder {
	b.SectionQuery.OnlyNotShownByDefault = true
	return b
}

func (b *QueryBuilder) OnlyTopLevel() *QueryBuilder {
	b.SectionQuery.OnlyTopLevel = true
	return b
}

func (b *QueryBuilder) WithoutSections(sections ...*Section) *QueryBuilder {
	b.SectionQuery.WithoutSections = sections
	return b
}

func (b *QueryBuilder) Topics(topics ...string) *QueryBuilder {
	b.SectionQuery.All = false
	b.SectionQuery.Topics = topics
	return b
}

func (b *QueryBuilder) Flags(flags ...string) *QueryBuilder {
	b.SectionQuery.All = false
	b.SectionQuery.Flags = flags
	return b
}

func (b *QueryBuilder) Commands(commands ...string) *QueryBuilder {
	b.SectionQuery.All = false
	b.SectionQuery.Commands = commands
	return b
}

func (b *QueryBuilder) Slugs(slugs ...string) *QueryBuilder {
	b.SectionQuery.All = false
	b.SectionQuery.Slugs = slugs
	return b
}

func (b *QueryBuilder) OnlyTopics(topics ...string) *QueryBuilder {
	b.SectionQuery.OnlyTopics = topics
	return b
}

func (b *QueryBuilder) OnlyFlags(flags ...string) *QueryBuilder {
	b.SectionQuery.OnlyFlags = flags
	return b
}

func (b *QueryBuilder) OnlyCommands(commands ...string) *QueryBuilder {
	b.SectionQuery.OnlyCommands = commands
	return b
}

func (b *QueryBuilder) OnlySlugs(slugs ...string) *QueryBuilder {
	b.SectionQuery.OnlySlugs = slugs
	return b
}

func (b *QueryBuilder) FindSections(sections []*Section) []*Section {
	return b.Build().FindSections(sections)
}

func (b *QueryBuilder) Build() *SectionQuery {
	return b.SectionQuery
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
