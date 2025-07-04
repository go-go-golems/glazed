package help

import (
	"context"
	"strings"

	"github.com/go-go-golems/glazed/pkg/help/model"
	"github.com/go-go-golems/glazed/pkg/help/store"
)

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
	Types map[SectionType]bool

	// if any of these is set, and they match each of the Only types,
	// the section will be return
	Topics   []string
	Flags    []string
	Commands []string
	Slugs    []string

	// this will return any section as long as it matches the Only strings
	All bool

	SearchedCommand string
	SearchedSlug    string

	// a section will be returned only if it matches all
	// of the following criteria
	OnlyTopics   []string
	OnlyFlags    []string
	OnlyCommands []string

	// We often need to filter sections that have already been shown
	WithoutSections []*Section
}

func NewSectionQuery() *SectionQuery {
	ret := &SectionQuery{
		Types: make(map[SectionType]bool),
		All:   true,
	}

	// per default, we don't return any kind of section
	return ret.DontReturnTopics().DontReturnExamples().DontReturnApplications().DontReturnTutorials()
}

func (s *SectionQuery) ReturnAllTypes() *SectionQuery {
	return s.ReturnTopics().ReturnExamples().ReturnApplications().ReturnTutorials()
}

func (s *SectionQuery) ReturnTopics() *SectionQuery {
	s.Types[SectionGeneralTopic] = true
	return s
}

func (s *SectionQuery) DontReturnTopics() *SectionQuery {
	s.Types[SectionGeneralTopic] = false
	return s
}

func (s *SectionQuery) ReturnExamples() *SectionQuery {
	s.Types[SectionExample] = true
	return s
}

func (s *SectionQuery) DontReturnExamples() *SectionQuery {
	s.Types[SectionExample] = false
	return s
}

func (s *SectionQuery) ReturnApplications() *SectionQuery {
	s.Types[SectionApplication] = true
	return s
}

func (s *SectionQuery) DontReturnApplications() *SectionQuery {
	s.Types[SectionApplication] = false
	return s
}

func (s *SectionQuery) ReturnTutorials() *SectionQuery {
	s.Types[SectionTutorial] = true
	return s
}

func (s *SectionQuery) SearchForSlug(slug string) *SectionQuery {
	s.SearchedSlug = slug
	return s
}

func (s *SectionQuery) SearchForCommand(command string) *SectionQuery {
	s.SearchedCommand = command
	return s
}

func (s *SectionQuery) DontReturnTutorials() *SectionQuery {
	s.Types[SectionTutorial] = false
	return s
}

func (s *SectionQuery) HasRestrictedReturnTypes() bool {
	for _, v := range s.Types {
		if !v {
			return true
		}
	}

	return false
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

func (s *SectionQuery) IsOnlyTopLevel() bool {
	return s.OnlyTopLevel
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

func (s *SectionQuery) HasOnlyQueries() bool {
	return len(s.OnlyTopics) > 0 || len(s.OnlyFlags) > 0 || len(s.OnlyCommands) > 0
}

func (s *SectionQuery) GetOnlyQueryAsString() string {
	ret := []string{}
	if len(s.OnlyTopics) > 0 {
		ret = append(ret, "topics "+strings.Join(s.OnlyTopics, ", "))
	}
	if len(s.OnlyFlags) > 0 {
		ret = append(ret, "flags "+strings.Join(s.OnlyFlags, ", "))
	}
	if len(s.OnlyCommands) > 0 {
		ret = append(ret, "commands "+strings.Join(s.OnlyCommands, ", "))
	}

	return strings.Join(ret, " and ")
}

func (s *SectionQuery) ResetOnlyQueries() *SectionQuery {
	s.OnlyTopics = []string{}
	s.OnlyFlags = []string{}
	s.OnlyCommands = []string{}
	return s
}

func (s *SectionQuery) Clone() *SectionQuery {
	ret := &SectionQuery{
		OnlyShownByDefault:    s.OnlyShownByDefault,
		OnlyNotShownByDefault: s.OnlyNotShownByDefault,
		OnlyTopLevel:          s.OnlyTopLevel,
		All:                   s.All,
		SearchedCommand:       s.SearchedCommand,
		SearchedSlug:          s.SearchedSlug,
	}

	// gotta love go, let's do a deep copy

	ret.Types = make(map[SectionType]bool)
	for k, v := range s.Types {
		ret.Types[k] = v
	}
	ret.Topics = make([]string, len(s.Topics))
	copy(ret.Topics, s.Topics)
	ret.Flags = make([]string, len(s.Flags))
	copy(ret.Flags, s.Flags)
	ret.Commands = make([]string, len(s.Commands))
	copy(ret.Commands, s.Commands)
	ret.Slugs = make([]string, len(s.Slugs))
	copy(ret.Slugs, s.Slugs)
	ret.OnlyTopics = make([]string, len(s.OnlyTopics))
	copy(ret.OnlyTopics, s.OnlyTopics)
	ret.OnlyFlags = make([]string, len(s.OnlyFlags))
	copy(ret.OnlyFlags, s.OnlyFlags)
	ret.OnlyCommands = make([]string, len(s.OnlyCommands))
	copy(ret.OnlyCommands, s.OnlyCommands)
	ret.WithoutSections = make([]*Section, len(s.WithoutSections))
	copy(ret.WithoutSections, s.WithoutSections)

	return ret
}

func (s *SectionQuery) FindSections(sections []*Section) []*Section {
	var result []*Section

sectionLoop:
	for _, section := range sections {
		if s.OnlyShownByDefault && !section.ShowPerDefault {
			continue
		}

		if s.OnlyNotShownByDefault && section.ShowPerDefault {
			continue
		}

		if s.OnlyTopLevel && !section.IsTopLevel {
			continue
		}

		for _, without := range s.WithoutSections {
			if without == section {
				continue sectionLoop
			}
		}

		if s.SearchedSlug != "" {
			if section.Slug != s.SearchedSlug {
				continue sectionLoop
			}
		}

		if s.SearchedCommand != "" {
			foundMatchingCommand := false
			for _, command := range section.Commands {
				if command == s.SearchedCommand {
					foundMatchingCommand = true
					break
				}
			}
			if !foundMatchingCommand {
				continue sectionLoop
			}
		}

		foundMatchingType, ok := s.Types[section.SectionType]
		if !ok || !foundMatchingType {
			continue
		}

		// filter out the Only*
		if len(s.OnlyTopics) > 0 {
			foundMatchingTopic := true
			for _, t := range s.OnlyTopics {
				if !section.IsForTopic(t) {
					foundMatchingTopic = false
					break
				}
			}
			if !foundMatchingTopic {
				continue sectionLoop
			}
		}

		if len(s.OnlyFlags) > 0 {
			foundMatchingFlag := true
			for _, f := range s.OnlyFlags {
				if !section.IsForFlag(f) {
					foundMatchingFlag = false
					break
				}
			}
			if !foundMatchingFlag {
				continue sectionLoop
			}
		}

		if len(s.OnlyCommands) > 0 {
			foundMatchingCommand := true
			for _, c := range s.OnlyCommands {
				if !section.IsForCommand(c) {
					foundMatchingCommand = false
					break
				}
			}
			if !foundMatchingCommand {
				continue sectionLoop
			}
		}

		if s.All {
			result = append(result, section)
			continue sectionLoop
		}

		for _, topic := range s.Topics {
			if section.IsForTopic(topic) {
				result = append(result, section)
				continue sectionLoop
			}
		}
		for _, flag := range s.Flags {
			if section.IsForFlag(flag) {
				result = append(result, section)
				continue sectionLoop
			}
		}
		for _, command := range s.Commands {
			if section.IsForCommand(command) {
				result = append(result, section)
				continue sectionLoop
			}
		}
		for _, slug := range s.Slugs {
			if section.Slug == slug {
				result = append(result, section)
				continue sectionLoop
			}
		}
	}

	return result
}

// FindSectionsWithStore finds sections using the store backend
func (s *SectionQuery) FindSectionsWithStore(ctx context.Context, st *store.Store) ([]*Section, error) {
	predicate := s.toPredicate()
	
	storeSections, err := st.Find(ctx, predicate)
	if err != nil {
		return nil, err
	}
	
	// Convert store sections to help sections
	sections := make([]*Section, len(storeSections))
	for i, storeSection := range storeSections {
		sections[i] = convertStoreSection(storeSection)
	}
	
	return sections, nil
}

// toPredicate converts the SectionQuery to a store predicate
func (s *SectionQuery) toPredicate() store.Predicate {
	var predicates []store.Predicate
	
	// Handle type filtering
	if s.HasRestrictedReturnTypes() {
		var typePredicates []store.Predicate
		for sectionType, enabled := range s.Types {
			if enabled {
				typePredicates = append(typePredicates, store.IsType(model.SectionType(sectionType)))
			}
		}
		if len(typePredicates) > 0 {
			predicates = append(predicates, store.Or(typePredicates...))
		}
	}
	
	// Handle visibility filters
	if s.OnlyShownByDefault {
		predicates = append(predicates, store.ShownByDefault())
	}
	if s.OnlyNotShownByDefault {
		predicates = append(predicates, store.NotShownByDefault())
	}
	
	// Handle top level filter
	if s.OnlyTopLevel {
		predicates = append(predicates, store.IsTopLevel())
	}
	
	// Handle slug search
	if s.SearchedSlug != "" {
		predicates = append(predicates, store.SlugEquals(s.SearchedSlug))
	}
	
	// Handle command search
	if s.SearchedCommand != "" {
		predicates = append(predicates, store.HasCommand(s.SearchedCommand))
	}
	
	// Handle Only* filters (all must match)
	if len(s.OnlyTopics) > 0 {
		for _, topic := range s.OnlyTopics {
			predicates = append(predicates, store.HasTopic(topic))
		}
	}
	if len(s.OnlyFlags) > 0 {
		for _, flag := range s.OnlyFlags {
			predicates = append(predicates, store.HasFlag(flag))
		}
	}
	if len(s.OnlyCommands) > 0 {
		for _, command := range s.OnlyCommands {
			predicates = append(predicates, store.HasCommand(command))
		}
	}
	
	// Handle Any* filters (any can match)
	if !s.All {
		var anyPredicates []store.Predicate
		
		for _, topic := range s.Topics {
			anyPredicates = append(anyPredicates, store.HasTopic(topic))
		}
		for _, flag := range s.Flags {
			anyPredicates = append(anyPredicates, store.HasFlag(flag))
		}
		for _, command := range s.Commands {
			anyPredicates = append(anyPredicates, store.HasCommand(command))
		}
		for _, slug := range s.Slugs {
			anyPredicates = append(anyPredicates, store.SlugEquals(slug))
		}
		
		if len(anyPredicates) > 0 {
			predicates = append(predicates, store.Or(anyPredicates...))
		}
	}
	
	// Handle excluded sections - this requires special handling since we need to exclude by ID
	if len(s.WithoutSections) > 0 {
		var excludedSlugs []string
		for _, section := range s.WithoutSections {
			excludedSlugs = append(excludedSlugs, section.Slug)
		}
		predicates = append(predicates, store.Not(store.SlugIn(excludedSlugs)))
	}
	
	// Add default ordering
	predicates = append(predicates, store.OrderByOrder())
	
	if len(predicates) == 0 {
		return func(qc *store.QueryCompiler) {} // Empty predicate
	}
	
	return store.And(predicates...)
}

// convertStoreSection converts a model.Section to a help.Section
func convertStoreSection(storeSection *model.Section) *Section {
	return &Section{
		Slug:           storeSection.Slug,
		SectionType:    SectionType(storeSection.SectionType),
		Title:          storeSection.Title,
		SubTitle:       storeSection.SubTitle,
		Short:          storeSection.Short,
		Content:        storeSection.Content,
		Topics:         storeSection.Topics,
		Flags:          storeSection.Flags,
		Commands:       storeSection.Commands,
		IsTopLevel:     storeSection.IsTopLevel,
		IsTemplate:     storeSection.IsTemplate,
		ShowPerDefault: storeSection.ShowPerDefault,
		Order:          storeSection.Order,
	}
}

func (s *SectionQuery) GetRequestedTypesAsString() string {
	requestedTypes := []string{}
	for sectionType, requested := range s.Types {
		if requested {
			switch sectionType {
			case SectionGeneralTopic:
				requestedTypes = append(requestedTypes, "general topics")
			case SectionExample:
				requestedTypes = append(requestedTypes, "examples")
			case SectionApplication:
				requestedTypes = append(requestedTypes, "applications")
			case SectionTutorial:
				requestedTypes = append(requestedTypes, "tutorials")

			}

		}
	}

	return strings.Join(requestedTypes, ", ")
}
