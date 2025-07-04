package query

import (
	"fmt"
	"strings"
)

// compiler accumulates SQL fragments and arguments
type compiler struct {
	joins   []string
	wheres  []string
	args    []interface{}
	joinSet map[string]bool // for deduplication
}

func newCompiler() *compiler {
	return &compiler{
		joins:   []string{},
		wheres:  []string{},
		args:    []interface{}{},
		joinSet: make(map[string]bool),
	}
}

func (c *compiler) addWhere(cond string, args ...interface{}) {
	c.wheres = append(c.wheres, cond)
	c.args = append(c.args, args...)
}

func (c *compiler) addJoin(join string) {
	if !c.joinSet[join] {
		c.joins = append(c.joins, join)
		c.joinSet[join] = true
	}
}

func (c *compiler) SQL() (string, []interface{}) {
	sql := "SELECT DISTINCT s.id, s.slug, s.title, s.subtitle, s.short, s.content, s.sectionType, s.isTopLevel, s.isTemplate, s.showDefault, s.ord FROM sections s"
	
	if len(c.joins) > 0 {
		sql += " " + strings.Join(c.joins, " ")
	}
	
	if len(c.wheres) > 0 {
		sql += " WHERE " + strings.Join(c.wheres, " AND ")
	}
	
	sql += " ORDER BY s.ord"
	return sql, c.args
}

func (c *compiler) merge(other *compiler) {
	for _, join := range other.joins {
		c.addJoin(join)
	}
	c.args = append(c.args, other.args...)
}

// Predicate represents a query predicate that can be compiled to SQL
type Predicate func(*compiler)

// Compile compiles a predicate to SQL and arguments
func Compile(pred Predicate) (string, []interface{}) {
	c := newCompiler()
	pred(c)
	return c.SQL()
}

// And combines predicates with AND logic
func And(preds ...Predicate) Predicate {
	return func(c *compiler) {
		if len(preds) == 0 {
			return
		}
		
		subWheres := make([]string, 0, len(preds))
		for _, p := range preds {
			sub := newCompiler()
			p(sub)
			c.merge(sub)
			if len(sub.wheres) > 0 {
				subWheres = append(subWheres, fmt.Sprintf("(%s)", strings.Join(sub.wheres, " AND ")))
			}
		}
		
		if len(subWheres) > 0 {
			c.wheres = append(c.wheres, fmt.Sprintf("(%s)", strings.Join(subWheres, " AND ")))
		}
	}
}

// Or combines predicates with OR logic  
func Or(preds ...Predicate) Predicate {
	return func(c *compiler) {
		if len(preds) == 0 {
			return
		}
		
		subWheres := make([]string, 0, len(preds))
		for _, p := range preds {
			sub := newCompiler()
			p(sub)
			c.merge(sub)
			if len(sub.wheres) > 0 {
				subWheres = append(subWheres, fmt.Sprintf("(%s)", strings.Join(sub.wheres, " AND ")))
			}
		}
		
		if len(subWheres) > 0 {
			c.wheres = append(c.wheres, fmt.Sprintf("(%s)", strings.Join(subWheres, " OR ")))
		}
	}
}

// Not negates a predicate
func Not(pred Predicate) Predicate {
	return func(c *compiler) {
		sub := newCompiler()
		pred(sub)
		c.merge(sub)
		if len(sub.wheres) > 0 {
			c.wheres = append(c.wheres, fmt.Sprintf("NOT (%s)", strings.Join(sub.wheres, " AND ")))
		}
	}
}
