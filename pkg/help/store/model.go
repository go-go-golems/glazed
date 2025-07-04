package store

import (
	"github.com/go-go-golems/glazed/pkg/help"
)

// Section represents a section in the SQLite database
type Section struct {
	ID            int64              `db:"id" json:"id"`
	Slug          string             `db:"slug" json:"slug"`
	SectionType   help.SectionType   `db:"section_type" json:"section_type"`
	Title         string             `db:"title" json:"title"`
	SubTitle      *string            `db:"sub_title" json:"sub_title,omitempty"`
	Short         *string            `db:"short" json:"short,omitempty"`
	Content       string             `db:"content" json:"content"`
	IsTopLevel    bool               `db:"is_top_level" json:"is_top_level"`
	IsTemplate    bool               `db:"is_template" json:"is_template"`
	ShowPerDefault bool               `db:"show_per_default" json:"show_per_default"`
	OrderIndex    int                `db:"order_index" json:"order_index"`
	CreatedAt     string             `db:"created_at" json:"created_at"`
	UpdatedAt     string             `db:"updated_at" json:"updated_at"`
	
	// These are loaded separately via joins
	Topics   []string `json:"topics,omitempty"`
	Flags    []string `json:"flags,omitempty"`
	Commands []string `json:"commands,omitempty"`
}

// ToHelpSection converts a store.Section to a help.Section
func (s *Section) ToHelpSection() *help.Section {
	section := &help.Section{
		Slug:           s.Slug,
		SectionType:    s.SectionType,
		Title:          s.Title,
		Content:        s.Content,
		IsTopLevel:     s.IsTopLevel,
		IsTemplate:     s.IsTemplate,
		ShowPerDefault: s.ShowPerDefault,
		Order:          s.OrderIndex,
		Topics:         s.Topics,
		Flags:          s.Flags,
		Commands:       s.Commands,
	}
	
	if s.SubTitle != nil {
		section.SubTitle = *s.SubTitle
	}
	
	if s.Short != nil {
		section.Short = *s.Short
	}
	
	return section
}

// FromHelpSection converts a help.Section to a store.Section
func FromHelpSection(section *help.Section) *Section {
	s := &Section{
		Slug:           section.Slug,
		SectionType:    section.SectionType,
		Title:          section.Title,
		Content:        section.Content,
		IsTopLevel:     section.IsTopLevel,
		IsTemplate:     section.IsTemplate,
		ShowPerDefault: section.ShowPerDefault,
		OrderIndex:     section.Order,
		Topics:         section.Topics,
		Flags:          section.Flags,
		Commands:       section.Commands,
	}
	
	if section.SubTitle != "" {
		s.SubTitle = &section.SubTitle
	}
	
	if section.Short != "" {
		s.Short = &section.Short
	}
	
	return s
}

// Topic represents a topic in the database
type Topic struct {
	ID   int64  `db:"id" json:"id"`
	Name string `db:"name" json:"name"`
}

// Flag represents a flag in the database
type Flag struct {
	ID   int64  `db:"id" json:"id"`
	Name string `db:"name" json:"name"`
}

// Command represents a command in the database
type Command struct {
	ID   int64  `db:"id" json:"id"`
	Name string `db:"name" json:"name"`
}
