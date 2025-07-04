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
		ctx := context.Background()
		modelSections, err := hs.Store.List(ctx, "order_num ASC")
		if err != nil {
			return nil, err
		}
		results := make([]*Section, len(modelSections))
		for i, modelSection := range modelSections {
			results[i] = &Section{Section: modelSection, HelpSystem: hs}
		}
		return results, nil
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

	// Get all sections from store for DSL evaluation
	ctx := context.Background()
	modelSections, err := hs.Store.List(ctx, "order_num ASC")
	if err != nil {
		return nil, err
	}

	// Use the DSL predicate to filter sections
	results := make([]*Section, 0)
	for _, modelSection := range modelSections {
		if hs.evaluatePredicate(predicate, modelSection) {
			// Convert to legacy section for return
			section := &Section{Section: modelSection, HelpSystem: hs}
			results = append(results, section)
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
		Section:    modelSection,
		HelpSystem: hs,
	}
}

// evaluatePredicate evaluates a predicate against a model section
func (hs *HelpSystem) evaluatePredicate(predicate store.Predicate, section *model.Section) bool {
	// Create a temporary in-memory store for evaluation
	memStore, err := store.NewInMemory()
	if err != nil {
		return false
	}
	defer func() {
		_ = memStore.Close()
	}()

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
				ctx := context.Background()
				modelSections, err := hs.Store.Find(ctx, store.SlugEquals(value))
				if err != nil {
					return []*Section{}, nil
				}
				for _, modelSection := range modelSections {
					results = append(results, &Section{Section: modelSection, HelpSystem: hs})
				}
				return results, nil
			default:
				return nil, fmt.Errorf("unknown field '%s'. Valid fields: type, topic, flag, command, slug", field)
			}
		}
	}

	// Handle quoted text search
	if strings.HasPrefix(query, "\"") && strings.HasSuffix(query, "\"") {
		searchTerm := query[1 : len(query)-1]
		ctx := context.Background()
		modelSections, err := hs.Store.Find(ctx, store.TextSearch(searchTerm))
		if err != nil {
			return []*Section{}, nil
		}
		for _, modelSection := range modelSections {
			results = append(results, &Section{Section: modelSection, HelpSystem: hs})
		}
		return results, nil
	}

	// Default: text search without quotes
	ctx := context.Background()
	modelSections, err := hs.Store.Find(ctx, store.TextSearch(query))
	if err != nil {
		return []*Section{}, nil
	}
	for _, modelSection := range modelSections {
		results = append(results, &Section{Section: modelSection, HelpSystem: hs})
	}

	return results, nil
}

func (hs *HelpSystem) filterByType(typeValue string) []*Section {
	ctx := context.Background()
	var predicate store.Predicate

	switch strings.ToLower(typeValue) {
	case "example":
		predicate = store.IsExample()
	case "tutorial":
		predicate = store.IsTutorial()
	case "topic", "generaltopic":
		predicate = store.IsGeneralTopic()
	case "application":
		predicate = store.IsApplication()
	default:
		return []*Section{}
	}

	modelSections, err := hs.Store.Find(ctx, predicate)
	if err != nil {
		return []*Section{}
	}

	results := make([]*Section, len(modelSections))
	for i, modelSection := range modelSections {
		results[i] = &Section{Section: modelSection, HelpSystem: hs}
	}
	return results
}

func (hs *HelpSystem) filterByTopic(topic string) []*Section {
	ctx := context.Background()
	modelSections, err := hs.Store.Find(ctx, store.HasTopic(topic))
	if err != nil {
		return []*Section{}
	}

	results := make([]*Section, len(modelSections))
	for i, modelSection := range modelSections {
		results[i] = &Section{Section: modelSection, HelpSystem: hs}
	}
	return results
}

func (hs *HelpSystem) filterByFlag(flag string) []*Section {
	// Normalize flag (remove leading dashes if present)
	flag = strings.TrimPrefix(flag, "--")
	flag = strings.TrimPrefix(flag, "-")

	ctx := context.Background()
	modelSections, err := hs.Store.Find(ctx, store.HasFlag(flag))
	if err != nil {
		return []*Section{}
	}

	results := make([]*Section, len(modelSections))
	for i, modelSection := range modelSections {
		results[i] = &Section{Section: modelSection, HelpSystem: hs}
	}
	return results
}

func (hs *HelpSystem) filterByCommand(command string) []*Section {
	ctx := context.Background()
	modelSections, err := hs.Store.Find(ctx, store.HasCommand(command))
	if err != nil {
		return []*Section{}
	}

	results := make([]*Section, len(modelSections))
	for i, modelSection := range modelSections {
		results[i] = &Section{Section: modelSection, HelpSystem: hs}
	}
	return results
}
