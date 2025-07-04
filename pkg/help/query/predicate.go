package query

import (
	"strings"

	"github.com/go-go-golems/glazed/pkg/help/model"
)

// Predicate is a function that adds conditions to a query compiler
type Predicate func(*compiler)

// Basic predicates for filtering sections

// IsType filters sections by their type
func IsType(t model.SectionType) Predicate {
	return func(c *compiler) {
		c.addWhere("s.sectionType = ?", t.String())
	}
}

// HasTopic filters sections that have a specific topic
func HasTopic(topic string) Predicate {
	return func(c *compiler) {
		c.addJoin("JOIN section_topics st ON st.section_id = s.id")
		c.addWhere("st.topic = ?", topic)
	}
}

// HasFlag filters sections that have a specific flag
func HasFlag(flag string) Predicate {
	return func(c *compiler) {
		c.addJoin("JOIN section_flags sf ON sf.section_id = s.id")
		c.addWhere("sf.flag = ?", flag)
	}
}

// HasCommand filters sections that have a specific command
func HasCommand(cmd string) Predicate {
	return func(c *compiler) {
		c.addJoin("JOIN section_commands sc ON sc.section_id = s.id")
		c.addWhere("sc.command = ?", cmd)
	}
}

// IsTopLevel filters sections that are top-level
func IsTopLevel() Predicate {
	return func(c *compiler) {
		c.addWhere("s.isTopLevel = ?", true)
	}
}

// ShownByDefault filters sections that are shown by default
func ShownByDefault() Predicate {
	return func(c *compiler) {
		c.addWhere("s.showDefault = ?", true)
	}
}

// NotShownByDefault filters sections that are not shown by default
func NotShownByDefault() Predicate {
	return func(c *compiler) {
		c.addWhere("s.showDefault = ?", false)
	}
}

// SlugEquals filters sections by exact slug match
func SlugEquals(slug string) Predicate {
	return func(c *compiler) {
		c.addWhere("s.slug = ?", slug)
	}
}

// TextSearch performs full-text search across section content
func TextSearch(term string) Predicate {
	return func(c *compiler) {
		c.addJoin("JOIN section_fts fts ON fts.rowid = s.id")
		c.addWhere("section_fts MATCH ?", term)
	}
}

// IsTemplate filters sections that are templates
func IsTemplate() Predicate {
	return func(c *compiler) {
		c.addWhere("s.isTemplate = ?", true)
	}
}

// Boolean combinators

// And combines multiple predicates with AND logic
func And(preds ...Predicate) Predicate {
	return func(c *compiler) {
		if len(preds) == 0 {
			return
		}
		
		var subWheres []string
		for _, p := range preds {
			sub := &compiler{}
			p(sub)
			
			// Add joins from subcompiler
			c.joins = append(c.joins, sub.joins...)
			
			// Add arguments from subcompiler
			c.args = append(c.args, sub.args...)
			
			// Combine WHERE clauses with AND
			if len(sub.wheres) > 0 {
				subWheres = append(subWheres, strings.Join(sub.wheres, " AND "))
			}
		}
		
		if len(subWheres) > 0 {
			c.wheres = append(c.wheres, "("+strings.Join(subWheres, " AND ")+")")
		}
		
		c.deduplicateJoins()
	}
}

// Or combines multiple predicates with OR logic
func Or(preds ...Predicate) Predicate {
	return func(c *compiler) {
		if len(preds) == 0 {
			return
		}
		
		var subWheres []string
		for _, p := range preds {
			sub := &compiler{}
			p(sub)
			
			// Add joins from subcompiler
			c.joins = append(c.joins, sub.joins...)
			
			// Add arguments from subcompiler
			c.args = append(c.args, sub.args...)
			
			// Combine WHERE clauses with OR
			if len(sub.wheres) > 0 {
				subWheres = append(subWheres, strings.Join(sub.wheres, " AND "))
			}
		}
		
		if len(subWheres) > 0 {
			c.wheres = append(c.wheres, "("+strings.Join(subWheres, " OR ")+")")
		}
		
		c.deduplicateJoins()
	}
}

// Not negates a predicate
func Not(pred Predicate) Predicate {
	return func(c *compiler) {
		sub := &compiler{}
		pred(sub)
		
		// Add joins from subcompiler
		c.joins = append(c.joins, sub.joins...)
		
		// Add arguments from subcompiler
		c.args = append(c.args, sub.args...)
		
		// Negate the WHERE clause
		if len(sub.wheres) > 0 {
			c.wheres = append(c.wheres, "NOT ("+strings.Join(sub.wheres, " AND ")+")")
		}
		
		c.deduplicateJoins()
	}
}

// Compile builds the final SQL query from a predicate
func Compile(pred Predicate) (string, []any) {
	c := &compiler{}
	pred(c)
	c.deduplicateJoins()
	return c.SQL()
}
