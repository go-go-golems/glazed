package query

import (
	"github.com/go-go-golems/glazed/pkg/help/model"
)

// IsType filters sections by type
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

// IsTopLevel filters sections that are top level
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

// TextSearch performs full-text search using FTS5
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

// NotTemplate filters sections that are not templates
func NotTemplate() Predicate {
	return func(c *compiler) {
		c.addWhere("s.isTemplate = ?", false)
	}
}
