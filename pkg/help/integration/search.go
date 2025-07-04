package integration

import (
	"context"

	"github.com/go-go-golems/glazed/pkg/help/model"
	"github.com/go-go-golems/glazed/pkg/help/query"
	"github.com/go-go-golems/glazed/pkg/help/search"
	"github.com/go-go-golems/glazed/pkg/help/store"
	"github.com/pkg/errors"
)

// SearchService provides high-level search functionality that integrates
// the text-based query DSL with the help store
type SearchService struct {
	store     *store.Store
	converter *search.Converter
	optimizer *search.QueryOptimizer
}

// NewSearchService creates a new search service
func NewSearchService(s *store.Store) *SearchService {
	return &SearchService{
		store:     s,
		converter: search.NewConverter(),
		optimizer: search.NewQueryOptimizer(),
	}
}

// Search executes a text-based query and returns matching sections
func (ss *SearchService) Search(ctx context.Context, textQuery string) ([]*model.Section, error) {
	if textQuery == "" {
		// Empty query returns all sections
		return ss.store.Find(ctx, func(comp *query.Compiler) {})
	}

	// Parse the text query
	query, err := search.ParseQuery(textQuery)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to parse query: %s", textQuery)
	}

	// Validate the query
	err = search.ValidateQuery(query)
	if err != nil {
		return nil, errors.Wrapf(err, "invalid query: %s", textQuery)
	}

	// Optimize the query
	optimized := ss.optimizer.Optimize(query)

	// Convert to predicate
	predicate, err := ss.converter.Convert(optimized)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to convert query: %s", textQuery)
	}

	// Execute the query
	sections, err := ss.store.Find(ctx, predicate)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to execute query: %s", textQuery)
	}

	return sections, nil
}

// SearchWithInfo executes a text-based query and returns both results and query information
func (ss *SearchService) SearchWithInfo(ctx context.Context, textQuery string) ([]*model.Section, *search.QueryInfo, error) {
	if textQuery == "" {
		sections, err := ss.store.Find(ctx, func(comp *query.Compiler) {})
		return sections, &search.QueryInfo{IsSimple: true}, err
	}

	// Parse the text query
	query, err := search.ParseQuery(textQuery)
	if err != nil {
		return nil, nil, errors.Wrapf(err, "failed to parse query: %s", textQuery)
	}

	// Analyze the query
	info := search.AnalyzeQuery(query)

	// Validate the query
	err = search.ValidateQuery(query)
	if err != nil {
		return nil, info, errors.Wrapf(err, "invalid query: %s", textQuery)
	}

	// Optimize the query
	optimized := ss.optimizer.Optimize(query)

	// Convert to predicate
	predicate, err := ss.converter.Convert(optimized)
	if err != nil {
		return nil, info, errors.Wrapf(err, "failed to convert query: %s", textQuery)
	}

	// Execute the query
	sections, err := ss.store.Find(ctx, predicate)
	if err != nil {
		return nil, info, errors.Wrapf(err, "failed to execute query: %s", textQuery)
	}

	return sections, info, nil
}

// ValidateQuery validates a text query without executing it
func (ss *SearchService) ValidateQuery(textQuery string) error {
	if textQuery == "" {
		return nil
	}

	// Parse the text query
	query, err := search.ParseQuery(textQuery)
	if err != nil {
		return errors.Wrapf(err, "failed to parse query: %s", textQuery)
	}

	// Validate the query
	err = search.ValidateQuery(query)
	if err != nil {
		return errors.Wrapf(err, "invalid query: %s", textQuery)
	}

	// Try to convert to predicate
	_, err = ss.converter.Convert(query)
	if err != nil {
		return errors.Wrapf(err, "failed to convert query: %s", textQuery)
	}

	return nil
}

// AnalyzeQuery parses and analyzes a text query without executing it
func (ss *SearchService) AnalyzeQuery(textQuery string) (*search.QueryInfo, error) {
	if textQuery == "" {
		return &search.QueryInfo{IsSimple: true}, nil
	}

	// Parse the text query
	query, err := search.ParseQuery(textQuery)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to parse query: %s", textQuery)
	}

	// Analyze the query
	info := search.AnalyzeQuery(query)
	return info, nil
}

// FormatQuery parses and formats a text query for display
func (ss *SearchService) FormatQuery(textQuery string) (string, error) {
	if textQuery == "" {
		return "", nil
	}

	// Parse the text query
	query, err := search.ParseQuery(textQuery)
	if err != nil {
		return "", errors.Wrapf(err, "failed to parse query: %s", textQuery)
	}

	// Format the query
	formatted := search.FormatQuery(query)
	return formatted, nil
}

// GetSupportedFields returns the list of supported filter fields
func (ss *SearchService) GetSupportedFields() []string {
	return search.GetSupportedFields()
}

// GetSupportedTypes returns the list of supported section types
func (ss *SearchService) GetSupportedTypes() []string {
	return search.GetSupportedTypes()
}

// GetFieldDescription returns a description of a filter field
func (ss *SearchService) GetFieldDescription(field string) string {
	return search.GetFieldDescription(field)
}

// BuildQuery provides a fluent interface for building queries programmatically
func (ss *SearchService) BuildQuery() *search.QueryBuilder {
	return search.NewQueryBuilder()
}

// ExecuteBuilder executes a query built with the query builder
func (ss *SearchService) ExecuteBuilder(ctx context.Context, builder *search.QueryBuilder) ([]*model.Section, error) {
	predicate := builder.Build()
	return ss.store.Find(ctx, predicate)
}

// CompileQuery compiles a text query to SQL for debugging
func (ss *SearchService) CompileQuery(textQuery string) (string, []interface{}, error) {
	if textQuery == "" {
		return "SELECT DISTINCT s.* FROM sections s ORDER BY s.ord", []interface{}{}, nil
	}

	// Parse the text query
	query, err := search.ParseQuery(textQuery)
	if err != nil {
		return "", nil, errors.Wrapf(err, "failed to parse query: %s", textQuery)
	}

	// Validate the query
	err = search.ValidateQuery(query)
	if err != nil {
		return "", nil, errors.Wrapf(err, "invalid query: %s", textQuery)
	}

	// Optimize the query
	optimized := ss.optimizer.Optimize(query)

	// Convert to predicate
	predicate, err := ss.converter.Convert(optimized)
	if err != nil {
		return "", nil, errors.Wrapf(err, "failed to convert query: %s", textQuery)
	}

	// Compile to SQL
	sql, args := query.Compile(predicate)
	return sql, args, nil
}

// SearchOptions provides configuration for search operations
type SearchOptions struct {
	Limit  int  // Maximum number of results to return (0 = no limit)
	Offset int  // Number of results to skip
	Debug  bool // Enable debug information
}

// SearchWithOptions executes a search with additional options
func (ss *SearchService) SearchWithOptions(ctx context.Context, textQuery string, opts SearchOptions) ([]*model.Section, error) {
	sections, err := ss.Search(ctx, textQuery)
	if err != nil {
		return nil, err
	}

	// Apply offset and limit
	if opts.Offset > 0 && opts.Offset < len(sections) {
		sections = sections[opts.Offset:]
	} else if opts.Offset >= len(sections) {
		sections = []*model.Section{}
	}

	if opts.Limit > 0 && opts.Limit < len(sections) {
		sections = sections[:opts.Limit]
	}

	return sections, nil
}

// SearchResult contains search results with metadata
type SearchResult struct {
	Sections  []*model.Section
	QueryInfo *search.QueryInfo
	SQL       string        // Debug information
	Args      []interface{} // Debug information
	Count     int          // Total number of results before limit/offset
}

// SearchWithMetadata executes a search and returns detailed results
func (ss *SearchService) SearchWithMetadata(ctx context.Context, textQuery string, opts SearchOptions) (*SearchResult, error) {
	result := &SearchResult{}

	// Get query info
	info, err := ss.AnalyzeQuery(textQuery)
	if err != nil {
		return nil, err
	}
	result.QueryInfo = info

	// Get debug information if requested
	if opts.Debug {
		sql, args, err := ss.CompileQuery(textQuery)
		if err != nil {
			return nil, err
		}
		result.SQL = sql
		result.Args = args
	}

	// Execute search
	sections, err := ss.Search(ctx, textQuery)
	if err != nil {
		return nil, err
	}

	result.Count = len(sections)

	// Apply offset and limit
	if opts.Offset > 0 && opts.Offset < len(sections) {
		sections = sections[opts.Offset:]
	} else if opts.Offset >= len(sections) {
		sections = []*model.Section{}
	}

	if opts.Limit > 0 && opts.Limit < len(sections) {
		sections = sections[:opts.Limit]
	}

	result.Sections = sections
	return result, nil
}
