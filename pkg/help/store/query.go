package store

import (
	"fmt"
	"strings"
)

// Predicate represents a query predicate that can be compiled to SQL
type Predicate func(*compiler)

// compiler builds SQL queries from predicates
type compiler struct {
	whereClause strings.Builder
	fromClause  strings.Builder
	joins       []string
	args        []interface{}
	joinCount   int
}

// newCompiler creates a new query compiler
func newCompiler() *compiler {
	c := &compiler{
		joins: make([]string, 0),
		args:  make([]interface{}, 0),
	}
	c.fromClause.WriteString("FROM sections s")
	return c
}

// addJoin adds a join clause if not already present
func (c *compiler) addJoin(join string) {
	for _, existingJoin := range c.joins {
		if existingJoin == join {
			return
		}
	}
	c.joins = append(c.joins, join)
}

// addWhere adds a WHERE condition
func (c *compiler) addWhere(condition string, args ...interface{}) {
	if c.whereClause.Len() > 0 {
		c.whereClause.WriteString(" AND ")
	}
	c.whereClause.WriteString("(")
	c.whereClause.WriteString(condition)
	c.whereClause.WriteString(")")
	c.args = append(c.args, args...)
}

// addWhereOr adds a WHERE condition with OR
func (c *compiler) addWhereOr(condition string, args ...interface{}) {
	if c.whereClause.Len() > 0 {
		c.whereClause.WriteString(" OR ")
	}
	c.whereClause.WriteString("(")
	c.whereClause.WriteString(condition)
	c.whereClause.WriteString(")")
	c.args = append(c.args, args...)
}

// compile builds the final SQL query
func (c *compiler) compile() (string, []interface{}) {
	var query strings.Builder
	
	// SELECT clause
	query.WriteString(`SELECT DISTINCT s.id, s.slug, s.section_type, s.title, s.sub_title, s.short, s.content, 
		s.is_top_level, s.is_template, s.show_per_default, s.order_index, s.created_at, s.updated_at `)
	
	// FROM clause
	query.WriteString(c.fromClause.String())
	
	// JOIN clauses
	for _, join := range c.joins {
		query.WriteString(" ")
		query.WriteString(join)
	}
	
	// WHERE clause
	if c.whereClause.Len() > 0 {
		query.WriteString(" WHERE ")
		query.WriteString(c.whereClause.String())
	}
	
	// ORDER BY clause
	query.WriteString(" ORDER BY s.order_index ASC, s.title ASC")
	
	return query.String(), c.args
}

// Predicate functions

// IsType filters by section type
func IsType(sectionType string) Predicate {
	return func(c *compiler) {
		c.addWhere("s.section_type = ?", sectionType)
	}
}

// HasTopic filters by topic
func HasTopic(topic string) Predicate {
	return func(c *compiler) {
		c.addJoin("JOIN section_topics st ON s.id = st.section_id")
		c.addJoin("JOIN topics t ON st.topic_id = t.id")
		c.addWhere("t.name = ?", topic)
	}
}

// HasFlag filters by flag
func HasFlag(flag string) Predicate {
	return func(c *compiler) {
		c.addJoin("JOIN section_flags sf ON s.id = sf.section_id")
		c.addJoin("JOIN flags f ON sf.flag_id = f.id")
		c.addWhere("f.name = ?", flag)
	}
}

// HasCommand filters by command
func HasCommand(command string) Predicate {
	return func(c *compiler) {
		c.addJoin("JOIN section_commands sc ON s.id = sc.section_id")
		c.addJoin("JOIN commands cmd ON sc.command_id = cmd.id")
		c.addWhere("cmd.name = ?", command)
	}
}

// IsTopLevel filters for top-level sections
func IsTopLevel() Predicate {
	return func(c *compiler) {
		c.addWhere("s.is_top_level = ?", true)
	}
}

// ShownByDefault filters for sections shown by default
func ShownByDefault() Predicate {
	return func(c *compiler) {
		c.addWhere("s.show_per_default = ?", true)
	}
}

// NotShownByDefault filters for sections not shown by default
func NotShownByDefault() Predicate {
	return func(c *compiler) {
		c.addWhere("s.show_per_default = ?", false)
	}
}

// SlugEquals filters by exact slug match
func SlugEquals(slug string) Predicate {
	return func(c *compiler) {
		c.addWhere("s.slug = ?", slug)
	}
}

// TextSearch performs full-text search using FTS5
func TextSearch(term string) Predicate {
	return func(c *compiler) {
		c.addJoin("JOIN sections_fts fts ON s.id = fts.rowid")
		c.addWhere("sections_fts MATCH ?", term)
	}
}

// IsTemplate filters for template sections
func IsTemplate() Predicate {
	return func(c *compiler) {
		c.addWhere("s.is_template = ?", true)
	}
}

// All creates a predicate that matches all sections
func All() Predicate {
	return func(c *compiler) {
		// No conditions - matches all sections
	}
}

// Boolean combinators

// And combines predicates with AND logic
func And(predicates ...Predicate) Predicate {
	return func(c *compiler) {
		if len(predicates) == 0 {
			return
		}
		
		// Create a sub-compiler for the AND group
		subCompiler := newCompiler()
		for _, pred := range predicates {
			pred(subCompiler)
		}
		
		// Merge joins
		for _, join := range subCompiler.joins {
			c.addJoin(join)
		}
		
		// Add the combined WHERE clause
		if subCompiler.whereClause.Len() > 0 {
			c.addWhere(subCompiler.whereClause.String(), subCompiler.args...)
		}
	}
}

// Or combines predicates with OR logic
func Or(predicates ...Predicate) Predicate {
	return func(c *compiler) {
		if len(predicates) == 0 {
			return
		}
		
		var orClauses []string
		var orArgs []interface{}
		
		for _, pred := range predicates {
			subCompiler := newCompiler()
			pred(subCompiler)
			
			// Merge joins
			for _, join := range subCompiler.joins {
				c.addJoin(join)
			}
			
			if subCompiler.whereClause.Len() > 0 {
				orClauses = append(orClauses, subCompiler.whereClause.String())
				orArgs = append(orArgs, subCompiler.args...)
			}
		}
		
		if len(orClauses) > 0 {
			orCondition := strings.Join(orClauses, " OR ")
			c.addWhere(orCondition, orArgs...)
		}
	}
}

// Not negates a predicate
func Not(predicate Predicate) Predicate {
	return func(c *compiler) {
		subCompiler := newCompiler()
		predicate(subCompiler)
		
		// Merge joins
		for _, join := range subCompiler.joins {
			c.addJoin(join)
		}
		
		if subCompiler.whereClause.Len() > 0 {
			negatedCondition := fmt.Sprintf("NOT (%s)", subCompiler.whereClause.String())
			c.addWhere(negatedCondition, subCompiler.args...)
		}
	}
}

// Helper predicates for section types

// IsGeneralTopic filters for general topic sections
func IsGeneralTopic() Predicate {
	return IsType("GeneralTopic")
}

// IsExample filters for example sections
func IsExample() Predicate {
	return IsType("Example")
}

// IsApplication filters for application sections
func IsApplication() Predicate {
	return IsType("Application")
}

// IsTutorial filters for tutorial sections
func IsTutorial() Predicate {
	return IsType("Tutorial")
}
