package query

import (
	"fmt"
	"strings"

	"github.com/go-go-golems/glazed/pkg/help/model"
)

// Predicate represents a query predicate that can be compiled to SQL
type Predicate func(*Compiler)

// Compiler accumulates SQL fragments for building queries
type Compiler struct {
	joins       []string
	wheres      []string
	args        []any
	aliasCount  map[string]int // Track alias usage for unique naming
}

func (c *Compiler) AddWhere(cond string, args ...any) {
	c.wheres = append(c.wheres, cond)
	c.args = append(c.args, args...)
}

func (c *Compiler) AddJoin(join string) {
	// Simple deduplication - avoid adding the same join twice
	for _, existing := range c.joins {
		if existing == join {
			return
		}
	}
	c.joins = append(c.joins, join)
}

func (c *Compiler) GetUniqueAlias(base string) string {
	if c.aliasCount == nil {
		c.aliasCount = make(map[string]int)
	}
	c.aliasCount[base]++
	if c.aliasCount[base] == 1 {
		return base
	}
	return fmt.Sprintf("%s%d", base, c.aliasCount[base])
}

func (c *Compiler) SQL() (string, []any) {
	sql := "SELECT DISTINCT s.* FROM sections s"
	if len(c.joins) > 0 {
		sql += " " + strings.Join(c.joins, " ")
	}
	if len(c.wheres) > 0 {
		sql += " WHERE " + strings.Join(c.wheres, " AND ")
	}
	sql += " ORDER BY s.ord"
	
	// Ensure args is never nil
	if c.args == nil {
		c.args = []any{}
	}
	return sql, c.args
}

// Compile compiles a predicate tree into SQL
func Compile(pred Predicate) (string, []any) {
	c := &Compiler{}
	pred(c)
	return c.SQL()
}

// Basic predicates
func IsType(t model.SectionType) Predicate {
	return func(c *Compiler) {
		c.AddWhere("s.sectionType = ?", t.String())
	}
}

func HasTopic(topic string) Predicate {
	return func(c *Compiler) {
		alias := c.GetUniqueAlias("st")
		c.AddJoin(fmt.Sprintf("JOIN section_topics %s ON %s.section_id = s.id", alias, alias))
		c.AddWhere(fmt.Sprintf("%s.topic = ?", alias), topic)
	}
}

func HasFlag(flag string) Predicate {
	return func(c *Compiler) {
		alias := c.GetUniqueAlias("sf")
		c.AddJoin(fmt.Sprintf("JOIN section_flags %s ON %s.section_id = s.id", alias, alias))
		c.AddWhere(fmt.Sprintf("%s.flag = ?", alias), flag)
	}
}

func HasCommand(cmd string) Predicate {
	return func(c *Compiler) {
		alias := c.GetUniqueAlias("sc")
		c.AddJoin(fmt.Sprintf("JOIN section_commands %s ON %s.section_id = s.id", alias, alias))
		c.AddWhere(fmt.Sprintf("%s.command = ?", alias), cmd)
	}
}

func IsTopLevel() Predicate {
	return func(c *Compiler) {
		c.AddWhere("s.isTopLevel = 1")
	}
}

func ShownByDefault() Predicate {
	return func(c *Compiler) {
		c.AddWhere("s.showDefault = 1")
	}
}

func SlugEquals(slug string) Predicate {
	return func(c *Compiler) {
		c.AddWhere("s.slug = ?", slug)
	}
}

func TextSearch(term string) Predicate {
	return func(c *Compiler) {
		c.AddJoin("JOIN section_fts fts ON fts.rowid = s.id")
		c.AddWhere("section_fts MATCH ?", term)
	}
}

// Boolean combinators
func And(preds ...Predicate) Predicate {
	return func(c *Compiler) {
		if len(preds) == 0 {
			return
		}
		if len(preds) == 1 {
			preds[0](c)
			return
		}

		// Execute all predicates directly on the same compiler
		// to share alias state
		for _, p := range preds {
			p(c)
		}
	}
}

func Or(preds ...Predicate) Predicate {
	return func(c *Compiler) {
		if len(preds) == 0 {
			return
		}
		if len(preds) == 1 {
			preds[0](c)
			return
		}

		// For OR, we need to collect WHERE clauses separately but share JOINs
		var subWheres []string
		startWheres := len(c.wheres)
		
		for _, p := range preds {
			// Execute predicate on main compiler to share JOIN state
			beforeWheres := len(c.wheres)
			p(c)
			
			// Collect the WHERE clauses added by this predicate
			if len(c.wheres) > beforeWheres {
				newWheres := c.wheres[beforeWheres:]
				if len(newWheres) == 1 {
					subWheres = append(subWheres, newWheres[0])
				} else {
					subWheres = append(subWheres, "("+strings.Join(newWheres, " AND ")+")")
				}
				// Remove the WHERE clauses from main compiler - we'll add them as OR
				c.wheres = c.wheres[:beforeWheres]
			}
		}
		
		if len(subWheres) > 0 {
			c.wheres = append(c.wheres[:startWheres], "("+strings.Join(subWheres, " OR ")+")")
		}
	}
}

func Not(pred Predicate) Predicate {
	return func(c *Compiler) {
		sub := &Compiler{}
		pred(sub)
		c.joins = append(c.joins, sub.joins...)
		if len(sub.wheres) > 0 {
			c.wheres = append(c.wheres, "NOT ("+strings.Join(sub.wheres, " AND ")+")")
		}
		c.args = append(c.args, sub.args...)
	}
}
