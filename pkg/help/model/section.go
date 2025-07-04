package model

import "fmt"

// SectionType represents the type of a help section.
type SectionType string

const (
	SectionGeneralTopic SectionType = "GeneralTopic"
	SectionExample      SectionType = "Example"
	SectionApplication  SectionType = "Application"
	SectionTutorial     SectionType = "Tutorial"
)

func (t SectionType) String() string {
	return string(t)
}

// Section represents a help/documentation section.
type Section struct {
	ID          int64    `yaml:"id,omitempty"`
	Slug        string   `yaml:"slug,omitempty"`
	Title       string   `yaml:"title,omitempty"`
	Subtitle    string   `yaml:"subtitle,omitempty"`
	Short       string   `yaml:"short,omitempty"`
	Content     string   `yaml:"content,omitempty"`
	SectionType SectionType `yaml:"sectionType,omitempty"`
	IsTopLevel  bool     `yaml:"isTopLevel,omitempty"`
	IsTemplate  bool     `yaml:"isTemplate,omitempty"`
	ShowDefault bool     `yaml:"showDefault,omitempty"`
	Ord         int      `yaml:"ord,omitempty"`
	Topics      []string `yaml:"topics,omitempty"`
	Flags       []string `yaml:"flags,omitempty"`
	Commands    []string `yaml:"commands,omitempty"`
}

func (s *Section) String() string {
	return fmt.Sprintf("Section[%s]: %s", s.SectionType, s.Title)
} 