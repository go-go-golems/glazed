package help

import (
	"context"
	"fmt"
	"strings"

	"github.com/go-go-golems/glazed/pkg/help/dsl"
	"github.com/go-go-golems/glazed/pkg/help/model"
	"github.com/go-go-golems/glazed/pkg/help/store"
)

// QuerySections performs a DSL query on the current help system with boolean logic support
func (hs *HelpSystem) QuerySections(query string) ([]*Section, error) {
	query = strings.TrimSpace(query)
	if query == "" {
		return hs.Sections, nil
	}

	// Try to parse the query using the DSL parser first
	predicate, err := dsl.ParseQuery(query)
	if err != nil {
		// Check if this looks like a DSL query (contains boolean operators)
		queryUpper := strings.ToUpper(query)
		if strings.Contains(queryUpper, " AND ") ||
			strings.Contains(queryUpper, " OR ") ||
			strings.Contains(queryUpper, "NOT ") ||
			strings.Contains(query, "(") || strings.Contains(query, ")") ||
			strings.HasSuffix(queryUpper, " AND") ||
			strings.HasSuffix(queryUpper, " OR") ||
			strings.HasPrefix(queryUpper, "NOT ") {
			// This is clearly a DSL query, return the error
			return nil, fmt.Errorf("invalid query syntax: %s", err.Error())
		}
		// Fall back to the legacy simple parser for backward compatibility
		return hs.queryLegacy(query)
	}

	// Convert legacy sections to model sections for DSL evaluation
	modelSections := make([]*model.Section, 0, len(hs.Sections))
	for _, section := range hs.Sections {
		modelSection := hs.convertToModelSection(section)
		modelSections = append(modelSections, modelSection)
	}

	// Use the DSL predicate to filter sections
	results := make([]*Section, 0)
	for _, modelSection := range modelSections {
		if hs.evaluatePredicate(predicate, modelSection) {
			// Convert back to legacy section for return
			legacySection := hs.convertFromModelSection(modelSection)
			results = append(results, legacySection)
		}
	}

	return results, nil
}

// convertToModelSection converts a legacy Section to a model.Section
func (hs *HelpSystem) convertToModelSection(section *Section) *model.Section {
	return &model.Section{
		Slug:           section.Slug,
		SectionType:    model.SectionType(section.SectionType),
		Title:          section.Title,
		SubTitle:       section.SubTitle,
		Short:          section.Short,
		Content:        section.Content,
		Topics:         section.Topics,
		Flags:          section.Flags,
		Commands:       section.Commands,
		IsTopLevel:     section.IsTopLevel,
		IsTemplate:     section.IsTemplate,
		ShowPerDefault: section.ShowPerDefault,
		Order:          section.Order,
	}
}

// convertFromModelSection converts a model.Section to a legacy Section
func (hs *HelpSystem) convertFromModelSection(modelSection *model.Section) *Section {
	return &Section{
		Slug:           modelSection.Slug,
		SectionType:    SectionType(modelSection.SectionType),
		Title:          modelSection.Title,
		SubTitle:       modelSection.SubTitle,
		Short:          modelSection.Short,
		Content:        modelSection.Content,
		Topics:         modelSection.Topics,
		Flags:          modelSection.Flags,
		Commands:       modelSection.Commands,
		IsTopLevel:     modelSection.IsTopLevel,
		IsTemplate:     modelSection.IsTemplate,
		ShowPerDefault: modelSection.ShowPerDefault,
		Order:          modelSection.Order,
		HelpSystem:     hs,
	}
}

// evaluatePredicate evaluates a predicate against a model section
func (hs *HelpSystem) evaluatePredicate(predicate store.Predicate, section *model.Section) bool {
	// Create a temporary in-memory store for evaluation
	memStore, err := store.NewInMemory()
	if err != nil {
		return false
	}
	defer memStore.Close()

	// Insert the section into the temporary store
	err = memStore.Insert(context.Background(), section)
	if err != nil {
		return false
	}

	// Find using the predicate
	results, err := memStore.Find(context.Background(), predicate)
	if err != nil {
		return false
	}

	return len(results) > 0
}

// queryLegacy provides backward compatibility with the old simple parser
func (hs *HelpSystem) queryLegacy(query string) ([]*Section, error) {
	var results []*Section

	// Handle simple cases first
	switch strings.ToLower(query) {
	case "examples":
		for _, section := range hs.Sections {
			if section.SectionType == SectionExample {
				results = append(results, section)
			}
		}
		return results, nil
	case "tutorials":
		for _, section := range hs.Sections {
			if section.SectionType == SectionTutorial {
				results = append(results, section)
			}
		}
		return results, nil
	case "topics":
		for _, section := range hs.Sections {
			if section.SectionType == SectionGeneralTopic {
				results = append(results, section)
			}
		}
		return results, nil
	case "applications":
		for _, section := range hs.Sections {
			if section.SectionType == SectionApplication {
				results = append(results, section)
			}
		}
		return results, nil
	case "toplevel":
		for _, section := range hs.Sections {
			if section.IsTopLevel {
				results = append(results, section)
			}
		}
		return results, nil
	case "defaults":
		for _, section := range hs.Sections {
			if section.ShowPerDefault {
				results = append(results, section)
			}
		}
		return results, nil
	}

	// Handle field:value queries
	if strings.Contains(query, ":") {
		parts := strings.SplitN(query, ":", 2)
		if len(parts) == 2 {
			field := strings.ToLower(strings.TrimSpace(parts[0]))
			value := strings.TrimSpace(parts[1])

			switch field {
			case "type":
				return hs.filterByType(value), nil
			case "topic":
				return hs.filterByTopic(value), nil
			case "flag":
				return hs.filterByFlag(value), nil
			case "command":
				return hs.filterByCommand(value), nil
			case "slug":
				for _, section := range hs.Sections {
					if strings.EqualFold(section.Slug, value) {
						results = append(results, section)
					}
				}
				return results, nil
			default:
				return nil, fmt.Errorf("unknown field '%s'. Valid fields: type, topic, flag, command, slug", field)
			}
		}
	}

	// Handle quoted text search
	if strings.HasPrefix(query, "\"") && strings.HasSuffix(query, "\"") {
		searchTerm := strings.ToLower(query[1 : len(query)-1])
		for _, section := range hs.Sections {
			content := strings.ToLower(section.Title + " " + section.Content + " " + section.Short)
			if strings.Contains(content, searchTerm) {
				results = append(results, section)
			}
		}
		return results, nil
	}

	// Default: text search without quotes
	searchTerm := strings.ToLower(query)
	for _, section := range hs.Sections {
		content := strings.ToLower(section.Title + " " + section.Content + " " + section.Short)
		if strings.Contains(content, searchTerm) {
			results = append(results, section)
		}
	}

	return results, nil
}

func (hs *HelpSystem) filterByType(typeValue string) []*Section {
	var results []*Section
	switch strings.ToLower(typeValue) {
	case "example":
		for _, section := range hs.Sections {
			if section.SectionType == SectionExample {
				results = append(results, section)
			}
		}
	case "tutorial":
		for _, section := range hs.Sections {
			if section.SectionType == SectionTutorial {
				results = append(results, section)
			}
		}
	case "topic", "generaltopic":
		for _, section := range hs.Sections {
			if section.SectionType == SectionGeneralTopic {
				results = append(results, section)
			}
		}
	case "application":
		for _, section := range hs.Sections {
			if section.SectionType == SectionApplication {
				results = append(results, section)
			}
		}
	}
	return results
}

func (hs *HelpSystem) filterByTopic(topic string) []*Section {
	var results []*Section
	for _, section := range hs.Sections {
		for _, sectionTopic := range section.Topics {
			if strings.EqualFold(sectionTopic, topic) {
				results = append(results, section)
				break
			}
		}
	}
	return results
}

func (hs *HelpSystem) filterByFlag(flag string) []*Section {
	var results []*Section
	// Normalize flag (remove leading dashes if present)
	flag = strings.TrimPrefix(flag, "--")
	flag = strings.TrimPrefix(flag, "-")

	for _, section := range hs.Sections {
		for _, sectionFlag := range section.Flags {
			cleanSectionFlag := strings.TrimPrefix(sectionFlag, "--")
			cleanSectionFlag = strings.TrimPrefix(cleanSectionFlag, "-")
			if strings.EqualFold(cleanSectionFlag, flag) {
				results = append(results, section)
				break
			}
		}
	}
	return results
}

func (hs *HelpSystem) filterByCommand(command string) []*Section {
	var results []*Section
	for _, section := range hs.Sections {
		for _, sectionCommand := range section.Commands {
			if strings.EqualFold(sectionCommand, command) {
				results = append(results, section)
				break
			}
		}
	}
	return results
}
