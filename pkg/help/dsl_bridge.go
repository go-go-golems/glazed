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
func (hs *HelpSystem) QuerySections(query string) ([]*model.Section, error) {
	query = strings.TrimSpace(query)
	if query == "" {
		ctx := context.Background()
		return hs.Store.List(ctx, "order_num ASC")
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

	// Pass the predicate directly to the store — single query, no O(N) temp stores
	ctx := context.Background()
	return hs.Store.Find(ctx, predicate)
}

// queryLegacy provides backward compatibility with the old simple parser
func (hs *HelpSystem) queryLegacy(query string) ([]*model.Section, error) {
	// Handle field:value queries
	if strings.Contains(query, ":") {
		parts := strings.SplitN(query, ":", 2)
		if len(parts) == 2 {
			field := strings.ToLower(strings.TrimSpace(parts[0]))
			value := strings.TrimSpace(parts[1])

			switch field {
			case "type":
				return hs.filterByType(value)
			case "topic":
				return hs.filterByTopic(value)
			case "flag":
				return hs.filterByFlag(value)
			case "command":
				return hs.filterByCommand(value)
			case "slug":
				ctx := context.Background()
				return hs.Store.Find(ctx, store.SlugEquals(value))
			default:
				return nil, fmt.Errorf("unknown field '%s'. Valid fields: type, topic, flag, command, slug", field)
			}
		}
	}

	// Handle quoted text search
	searchTerm := query
	if strings.HasPrefix(query, "\"") && strings.HasSuffix(query, "\"") {
		searchTerm = query[1 : len(query)-1]
	}

	ctx := context.Background()
	return hs.Store.Find(ctx, store.TextSearch(searchTerm))
}

func (hs *HelpSystem) filterByType(typeValue string) ([]*model.Section, error) {
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
		return nil, fmt.Errorf("unknown section type: %s", typeValue)
	}

	return hs.Store.Find(ctx, predicate)
}

func (hs *HelpSystem) filterByTopic(topic string) ([]*model.Section, error) {
	ctx := context.Background()
	return hs.Store.Find(ctx, store.HasTopic(topic))
}

func (hs *HelpSystem) filterByFlag(flag string) ([]*model.Section, error) {
	// Normalize flag (remove leading dashes if present)
	flag = strings.TrimPrefix(flag, "--")
	flag = strings.TrimPrefix(flag, "-")

	ctx := context.Background()
	return hs.Store.Find(ctx, store.HasFlag(flag))
}

func (hs *HelpSystem) filterByCommand(command string) ([]*model.Section, error) {
	ctx := context.Background()
	return hs.Store.Find(ctx, store.HasCommand(command))
}
