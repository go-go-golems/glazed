package query

import (
	"fmt"
	"github.com/go-go-golems/glazed/pkg/help/model"
	"strings"
)

var joinCounter int32

type Predicate func(*Compiler)

type Compiler struct {
	joins        []string
	wheres       []string
	args         []any
	aliasCounter int
}

func (c *Compiler) nextAlias(prefix string) string {
	c.aliasCounter++
	return fmt.Sprintf("%s%d", prefix, c.aliasCounter)
}

func (c *Compiler) addWhere(cond string, args ...any) {
	c.wheres = append(c.wheres, cond)
	c.args = append(c.args, args...)
}

func (c *Compiler) addJoin(join string) {
	c.joins = append(c.joins, join)
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
	return sql, c.args
}

func IsType(t model.SectionType) Predicate {
	return func(c *Compiler) {
		c.addWhere("s.sectionType = ?", t.String())
	}
}

func HasTopic(topic string) Predicate {
	return func(c *Compiler) {
		alias := c.nextAlias("st")
		c.addJoin(fmt.Sprintf("JOIN section_topics %s ON %s.section_id = s.id", alias, alias))
		c.addWhere(fmt.Sprintf("%s.topic = ?", alias), topic)
	}
}

func HasFlag(flag string) Predicate {
	return func(c *Compiler) {
		alias := c.nextAlias("sf")
		c.addJoin(fmt.Sprintf("JOIN section_flags %s ON %s.section_id = s.id", alias, alias))
		c.addWhere(fmt.Sprintf("%s.flag = ?", alias), flag)
	}
}

func HasCommand(cmd string) Predicate {
	return func(c *Compiler) {
		alias := c.nextAlias("sc")
		c.addJoin(fmt.Sprintf("JOIN section_commands %s ON %s.section_id = s.id", alias, alias))
		c.addWhere(fmt.Sprintf("%s.command = ?", alias), cmd)
	}
}

func IsTopLevel() Predicate {
	return func(c *Compiler) {
		c.addWhere("s.isTopLevel = 1")
	}
}

func ShownByDefault() Predicate {
	return func(c *Compiler) {
		c.addWhere("s.showDefault = 1")
	}
}

func SlugEquals(slug string) Predicate {
	return func(c *Compiler) {
		c.addWhere("s.slug = ?", slug)
	}
}

func TextSearch(term string) Predicate {
	return func(c *Compiler) {
		alias := c.nextAlias("fts")
		c.addJoin(fmt.Sprintf("JOIN section_fts %s ON %s.rowid = s.id", alias, alias))
		c.addWhere("section_fts MATCH ?", term)
	}
}

func And(preds ...Predicate) Predicate {
	return func(c *Compiler) {
		var subWheres []string
		var subJoins []string
		var subArgs []any
		for _, p := range preds {
			sub := &Compiler{aliasCounter: c.aliasCounter}
			p(sub)
			c.aliasCounter = sub.aliasCounter
			if len(sub.wheres) > 0 {
				subWheres = append(subWheres, "("+strings.Join(sub.wheres, " AND ")+")")
			}
			subJoins = append(subJoins, sub.joins...)
			subArgs = append(subArgs, sub.args...)
		}
		c.joins = append(c.joins, subJoins...)
		if len(subWheres) > 0 {
			c.wheres = append(c.wheres, "("+strings.Join(subWheres, " AND ")+")")
		}
		c.args = append(c.args, subArgs...)
	}
}

func Or(preds ...Predicate) Predicate {
	return func(c *Compiler) {
		var subWheres []string
		var subJoins []string
		var subArgs []any
		for _, p := range preds {
			sub := &Compiler{aliasCounter: c.aliasCounter}
			p(sub)
			c.aliasCounter = sub.aliasCounter
			if len(sub.wheres) > 0 {
				subWheres = append(subWheres, "("+strings.Join(sub.wheres, " AND ")+")")
			}
			subJoins = append(subJoins, sub.joins...)
			subArgs = append(subArgs, sub.args...)
		}
		c.joins = append(c.joins, subJoins...)
		if len(subWheres) > 0 {
			c.wheres = append(c.wheres, "("+strings.Join(subWheres, " OR ")+")")
		}
		c.args = append(c.args, subArgs...)
	}
}

func Not(pred Predicate) Predicate {
	return func(c *Compiler) {
		sub := &Compiler{aliasCounter: c.aliasCounter}
		pred(sub)
		c.aliasCounter = sub.aliasCounter
		c.joins = append(c.joins, sub.joins...)
		if len(sub.wheres) > 0 {
			c.wheres = append(c.wheres, "NOT ("+strings.Join(sub.wheres, " AND ")+")")
		}
		c.args = append(c.args, sub.args...)
	}
}

// Special Not predicate for HasFlag
func NotHasFlag(flag string) Predicate {
	return func(c *Compiler) {
		c.addWhere("NOT EXISTS (SELECT 1 FROM section_flags nf WHERE nf.section_id = s.id AND nf.flag = ?)", flag)
	}
}

// Special Not predicate for HasTopic
func NotHasTopic(topic string) Predicate {
	return func(c *Compiler) {
		c.addWhere("NOT EXISTS (SELECT 1 FROM section_topics nt WHERE nt.section_id = s.id AND nt.topic = ?)", topic)
	}
}

// Special Not predicate for HasCommand
func NotHasCommand(cmd string) Predicate {
	return func(c *Compiler) {
		c.addWhere("NOT EXISTS (SELECT 1 FROM section_commands nc WHERE nc.section_id = s.id AND nc.command = ?)", cmd)
	}
}
