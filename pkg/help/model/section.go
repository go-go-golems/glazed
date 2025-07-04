package model

import (
	"fmt"
)

// Section represents a help/documentation section.
type Section struct {
	ID          int64       `yaml:"id,omitempty"`
	Slug        string      `yaml:"slug,omitempty"`
	Title       string      `yaml:"title,omitempty"`
	Subtitle    string      `yaml:"subtitle,omitempty"`
	Short       string      `yaml:"short,omitempty"`
	Content     string      `yaml:"content,omitempty"`
	SectionType SectionType `yaml:"sectionType,omitempty"`
	IsTopLevel  bool        `yaml:"isTopLevel,omitempty"`
	IsTemplate  bool        `yaml:"isTemplate,omitempty"`
	ShowDefault bool        `yaml:"showDefault,omitempty"`
	Ord         int         `yaml:"ord,omitempty"`
	Topics      []string    `yaml:"topics,omitempty"`
	Flags       []string    `yaml:"flags,omitempty"`
	Commands    []string    `yaml:"commands,omitempty"`
}

func (s *Section) String() string {
	return fmt.Sprintf("Section[%s]: %s", s.SectionType, s.Title)
}

// Add helper methods to model.Section to match the old Section methods (IsForCommand, IsForFlag, IsForTopic) if needed for rendering or logic.
func (s *Section) IsForCommand(command string) bool {
	for _, c := range s.Commands {
		if c == command {
			return true
		}
	}
	return false
}

func (s *Section) IsForFlag(flag string) bool {
	for _, f := range s.Flags {
		if f == flag {
			return true
		}
	}
	return false
}

func (s *Section) IsForTopic(topic string) bool {
	for _, t := range s.Topics {
		if t == topic {
			return true
		}
	}
	return false
}
