package model

import (
	"github.com/pkg/errors"
)

// SectionType represents the type of a help section
type SectionType int

const (
	SectionGeneralTopic SectionType = iota
	SectionExample
	SectionApplication
	SectionTutorial
)

func SectionTypeFromString(s string) (SectionType, error) {
	switch s {
	case "GeneralTopic":
		return SectionGeneralTopic, nil
	case "Example":
		return SectionExample, nil
	case "Application":
		return SectionApplication, nil
	case "Tutorial":
		return SectionTutorial, nil
	}
	return SectionGeneralTopic, errors.Errorf("unknown section type %s", s)
}

func (s SectionType) String() string {
	switch s {
	case SectionGeneralTopic:
		return "GeneralTopic"
	case SectionExample:
		return "Example"
	case SectionApplication:
		return "Application"
	case SectionTutorial:
		return "Tutorial"
	}
	return "Unknown"
}

// Section represents a help documentation section
type Section struct {
	ID          int64
	Slug        string
	Title       string
	Subtitle    string
	Short       string
	Content     string
	SectionType SectionType
	IsTopLevel  bool
	IsTemplate  bool
	ShowDefault bool
	Order       int

	Topics   []string
	Flags    []string
	Commands []string
}

// IsForCommand checks if this section is relevant for a specific command
func (s *Section) IsForCommand(command string) bool {
	for _, c := range s.Commands {
		if c == command {
			return true
		}
	}
	return false
}

// IsForFlag checks if this section is relevant for a specific flag
func (s *Section) IsForFlag(flag string) bool {
	for _, f := range s.Flags {
		if f == flag {
			return true
		}
	}
	return false
}

// IsForTopic checks if this section is relevant for a specific topic
func (s *Section) IsForTopic(topic string) bool {
	for _, t := range s.Topics {
		if t == topic {
			return true
		}
	}
	return false
}
