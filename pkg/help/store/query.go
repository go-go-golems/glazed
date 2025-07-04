package store

import (
	"context"
	"fmt"
	"strings"

	"github.com/go-go-golems/glazed/pkg/help/model"
	"github.com/pkg/errors"
)

// Predicate represents a query predicate that can be compiled to SQL
type Predicate func(*QueryCompiler)

// QueryCompiler builds SQL queries from predicates
type QueryCompiler struct {
	whereClause []string
	args        []interface{}
	joins       []string
	orderBy     string
	limit       int
	offset      int
}

// NewQueryCompiler creates a new query compiler
func NewQueryCompiler() *QueryCompiler {
	return &QueryCompiler{
		whereClause: []string{},
		args:        []interface{}{},
		joins:       []string{},
	}
}

// AddWhere adds a WHERE clause condition
func (qc *QueryCompiler) AddWhere(condition string, args ...interface{}) {
	qc.whereClause = append(qc.whereClause, condition)
	qc.args = append(qc.args, args...)
}

// AddJoin adds a JOIN clause
func (qc *QueryCompiler) AddJoin(join string) {
	qc.joins = append(qc.joins, join)
}

// SetOrderBy sets the ORDER BY clause
func (qc *QueryCompiler) SetOrderBy(orderBy string) {
	qc.orderBy = orderBy
}

// SetLimit sets the LIMIT clause
func (qc *QueryCompiler) SetLimit(limit int) {
	qc.limit = limit
}

// SetOffset sets the OFFSET clause
func (qc *QueryCompiler) SetOffset(offset int) {
	qc.offset = offset
}

// BuildQuery builds the complete SQL query
func (qc *QueryCompiler) BuildQuery() (string, []interface{}) {
	baseQuery := `
		SELECT s.id, s.slug, s.section_type, s.title, s.sub_title, s.short, s.content,
			s.topics, s.flags, s.commands, s.is_top_level, s.is_template,
			s.show_per_default, s.order_num, s.created_at, s.updated_at
		FROM sections s
	`

	// Add joins
	for _, join := range qc.joins {
		baseQuery += " " + join
	}

	// Add WHERE clause
	if len(qc.whereClause) > 0 {
		baseQuery += " WHERE " + strings.Join(qc.whereClause, " AND ")
	}

	// Add ORDER BY
	if qc.orderBy != "" {
		baseQuery += " ORDER BY " + qc.orderBy
	}

	// Add LIMIT and OFFSET (SQLite requires LIMIT when using OFFSET)
	if qc.limit > 0 {
		baseQuery += fmt.Sprintf(" LIMIT %d", qc.limit)
		if qc.offset > 0 {
			baseQuery += fmt.Sprintf(" OFFSET %d", qc.offset)
		}
	} else if qc.offset > 0 {
		// If only offset is specified, use a very large limit
		baseQuery += fmt.Sprintf(" LIMIT -1 OFFSET %d", qc.offset)
	}

	return baseQuery, qc.args
}

// Find executes a query using the provided predicate
func (s *Store) Find(ctx context.Context, predicate Predicate) ([]*model.Section, error) {
	compiler := NewQueryCompiler()
	predicate(compiler)

	query, args := compiler.BuildQuery()
	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, errors.Wrap(err, "failed to execute query")
	}
	defer func() {
		_ = rows.Close()
	}()

	var sections []*model.Section
	for rows.Next() {
		section, err := s.scanSection(rows)
		if err != nil {
			return nil, errors.Wrap(err, "failed to scan section")
		}
		sections = append(sections, section)
	}

	return sections, nil
}

// Predicate constructors

// IsType filters sections by type
func IsType(sectionType model.SectionType) Predicate {
	return func(qc *QueryCompiler) {
		qc.AddWhere("s.section_type = ?", sectionType.ToInt())
	}
}

// HasTopic filters sections that have a specific topic
func HasTopic(topic string) Predicate {
	return func(qc *QueryCompiler) {
		qc.AddWhere("s.topics LIKE ?", "%"+topic+"%")
	}
}

// HasFlag filters sections that have a specific flag
func HasFlag(flag string) Predicate {
	return func(qc *QueryCompiler) {
		qc.AddWhere("s.flags LIKE ?", "%"+flag+"%")
	}
}

// HasCommand filters sections that have a specific command
func HasCommand(command string) Predicate {
	return func(qc *QueryCompiler) {
		qc.AddWhere("s.commands LIKE ?", "%"+command+"%")
	}
}

// IsTopLevel filters sections that are top level
func IsTopLevel() Predicate {
	return func(qc *QueryCompiler) {
		qc.AddWhere("s.is_top_level = ?", true)
	}
}

// ShownByDefault filters sections that are shown by default
func ShownByDefault() Predicate {
	return func(qc *QueryCompiler) {
		qc.AddWhere("s.show_per_default = ?", true)
	}
}

// NotShownByDefault filters sections that are not shown by default
func NotShownByDefault() Predicate {
	return func(qc *QueryCompiler) {
		qc.AddWhere("s.show_per_default = ?", false)
	}
}

// SlugEquals filters sections by exact slug match
func SlugEquals(slug string) Predicate {
	return func(qc *QueryCompiler) {
		qc.AddWhere("s.slug = ?", slug)
	}
}

// SlugIn filters sections by slug list
func SlugIn(slugs []string) Predicate {
	return func(qc *QueryCompiler) {
		if len(slugs) == 0 {
			// Add a condition that will never match to return no results
			qc.AddWhere("1 = 0")
			return
		}
		placeholders := make([]string, len(slugs))
		args := make([]interface{}, len(slugs))
		for i, slug := range slugs {
			placeholders[i] = "?"
			args[i] = slug
		}
		qc.AddWhere("s.slug IN ("+strings.Join(placeholders, ",")+")", args...)
	}
}

// TitleContains filters sections where title contains the term
func TitleContains(term string) Predicate {
	return func(qc *QueryCompiler) {
		qc.AddWhere("s.title LIKE ?", "%"+term+"%")
	}
}

// ContentContains filters sections where content contains the term
func ContentContains(term string) Predicate {
	return func(qc *QueryCompiler) {
		qc.AddWhere("s.content LIKE ?", "%"+term+"%")
	}
}

// OrderByOrder sorts sections by their order field
func OrderByOrder() Predicate {
	return func(qc *QueryCompiler) {
		qc.SetOrderBy("s.order_num ASC")
	}
}

// OrderByTitle sorts sections by title
func OrderByTitle() Predicate {
	return func(qc *QueryCompiler) {
		qc.SetOrderBy("s.title ASC")
	}
}

// OrderByCreatedAt sorts sections by creation time
func OrderByCreatedAt() Predicate {
	return func(qc *QueryCompiler) {
		qc.SetOrderBy("s.created_at DESC")
	}
}

// Limit sets the maximum number of results
func Limit(limit int) Predicate {
	return func(qc *QueryCompiler) {
		qc.SetLimit(limit)
	}
}

// Offset sets the number of results to skip
func Offset(offset int) Predicate {
	return func(qc *QueryCompiler) {
		qc.SetOffset(offset)
	}
}

// Boolean combinators

// And combines multiple predicates with AND logic
func And(predicates ...Predicate) Predicate {
	return func(qc *QueryCompiler) {
		for _, predicate := range predicates {
			predicate(qc)
		}
	}
}

// Or combines multiple predicates with OR logic
func Or(predicates ...Predicate) Predicate {
	return func(qc *QueryCompiler) {
		if len(predicates) == 0 {
			return
		}

		// Build separate compiler for each predicate
		orClauses := make([]string, 0, len(predicates))
		allArgs := make([]interface{}, 0)

		for _, predicate := range predicates {
			subCompiler := NewQueryCompiler()
			predicate(subCompiler)

			if len(subCompiler.whereClause) > 0 {
				orClause := "(" + strings.Join(subCompiler.whereClause, " AND ") + ")"
				orClauses = append(orClauses, orClause)
				allArgs = append(allArgs, subCompiler.args...)
			}
		}

		if len(orClauses) > 0 {
			qc.AddWhere("("+strings.Join(orClauses, " OR ")+")", allArgs...)
		}
	}
}

// Not negates a predicate
func Not(predicate Predicate) Predicate {
	return func(qc *QueryCompiler) {
		subCompiler := NewQueryCompiler()
		predicate(subCompiler)

		if len(subCompiler.whereClause) > 0 {
			notClause := "NOT (" + strings.Join(subCompiler.whereClause, " AND ") + ")"
			qc.AddWhere(notClause, subCompiler.args...)
		}
	}
}

// Convenience predicates for common combinations

// IsTutorial filters for tutorial sections
func IsTutorial() Predicate {
	return IsType(model.SectionTutorial)
}

// IsExample filters for example sections
func IsExample() Predicate {
	return IsType(model.SectionExample)
}

// IsGeneralTopic filters for general topic sections
func IsGeneralTopic() Predicate {
	return IsType(model.SectionGeneralTopic)
}

// IsApplication filters for application sections
func IsApplication() Predicate {
	return IsType(model.SectionApplication)
}

// IsTemplate filters for template sections
func IsTemplate() Predicate {
	return func(qc *QueryCompiler) {
		qc.AddWhere("s.is_template = ?", true)
	}
}

// TopLevelDefaults filters for top-level sections shown by default
func TopLevelDefaults() Predicate {
	return And(IsTopLevel(), ShownByDefault())
}

// ExamplesForTopic filters for examples related to a specific topic
func ExamplesForTopic(topic string) Predicate {
	return And(IsExample(), HasTopic(topic))
}

// TutorialsForTopic filters for tutorials related to a specific topic
func TutorialsForTopic(topic string) Predicate {
	return And(IsTutorial(), HasTopic(topic))
}

// DefaultExamplesForTopic filters for default examples related to a specific topic
func DefaultExamplesForTopic(topic string) Predicate {
	return And(IsExample(), HasTopic(topic), ShownByDefault())
}

// DefaultTutorialsForTopic filters for default tutorials related to a specific topic
func DefaultTutorialsForTopic(topic string) Predicate {
	return And(IsTutorial(), HasTopic(topic), ShownByDefault())
}
