package query

import (
	"strings"
)

// compiler accumulates SQL fragments to build a complete query
type compiler struct {
	joins  []string
	wheres []string
	args   []any
}

// addWhere adds a WHERE condition with arguments
func (c *compiler) addWhere(cond string, args ...any) {
	c.wheres = append(c.wheres, cond)
	c.args = append(c.args, args...)
}

// addJoin adds a JOIN clause
func (c *compiler) addJoin(join string) {
	c.joins = append(c.joins, join)
}

// SQL generates the final SQL query and arguments
func (c *compiler) SQL() (string, []any) {
	sql := "SELECT DISTINCT s.* FROM sections s"
	
	if len(c.joins) > 0 {
		sql += " " + strings.Join(c.joins, " ")
	}
	
	if len(c.wheres) > 0 {
		sql += " WHERE " + strings.Join(c.wheres, " AND ")
	}
	
	sql += " ORDER BY s.ord"
	
	return sql, c.args
}

// deduplicateJoins removes duplicate JOIN clauses
func (c *compiler) deduplicateJoins() {
	seen := make(map[string]bool)
	var unique []string
	
	for _, join := range c.joins {
		if !seen[join] {
			seen[join] = true
			unique = append(unique, join)
		}
	}
	
	c.joins = unique
}
