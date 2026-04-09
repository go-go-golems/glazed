package model

import (
	"strings"

	strings2 "github.com/go-go-golems/glazed/pkg/helpers/strings"
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

// SectionTypeFromString converts a string to a SectionType
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

// String returns the string representation of a SectionType
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

// ToInt returns the integer value of a SectionType
func (s SectionType) ToInt() int {
	return int(s)
}

// Section represents a help section with all its metadata
type Section struct {
	ID          int64       `json:"id,omitempty"`
	Slug        string      `json:"slug"`
	SectionType SectionType `json:"section_type"`

	Title    string `json:"title"`
	SubTitle string `json:"sub_title"`
	Short    string `json:"short"`
	Content  string `json:"content"`

	// Metadata for searching and filtering
	Topics   []string `json:"topics"`
	Flags    []string `json:"flags"`
	Commands []string `json:"commands"`

	// Display options
	IsTopLevel     bool `json:"is_top_level"`
	IsTemplate     bool `json:"is_template"`
	ShowPerDefault bool `json:"show_per_default"`
	Order          int  `json:"order"`

	CreatedAt string `json:"created_at,omitempty"`
	UpdatedAt string `json:"updated_at,omitempty"`
}

// IsForCommand checks if the section is related to a specific command
func (s *Section) IsForCommand(command string) bool {
	return strings2.StringInSlice(command, s.Commands)
}

// IsForFlag checks if the section is related to a specific flag
func (s *Section) IsForFlag(flag string) bool {
	return strings2.StringInSlice(flag, s.Flags)
}

// IsForTopic checks if the section is related to a specific topic
func (s *Section) IsForTopic(topic string) bool {
	return strings2.StringInSlice(topic, s.Topics)
}

// Validate ensures the section has required fields
func (s *Section) Validate() error {
	if s.Slug == "" {
		return errors.New("section slug is required")
	}
	if s.Title == "" {
		return errors.New("section title is required")
	}
	return nil
}

// TopicsAsString returns topics as a comma-separated string
func (s *Section) TopicsAsString() string {
	return strings.Join(s.Topics, ",")
}

// FlagsAsString returns flags as a comma-separated string
func (s *Section) FlagsAsString() string {
	return strings.Join(s.Flags, ",")
}

// CommandsAsString returns commands as a comma-separated string
func (s *Section) CommandsAsString() string {
	return strings.Join(s.Commands, ",")
}

// SetTopicsFromString sets topics from a comma-separated string
func (s *Section) SetTopicsFromString(topics string) {
	if topics == "" {
		s.Topics = []string{}
		return
	}
	s.Topics = strings.Split(topics, ",")
}

// SetFlagsFromString sets flags from a comma-separated string
func (s *Section) SetFlagsFromString(flags string) {
	if flags == "" {
		s.Flags = []string{}
		return
	}
	s.Flags = strings.Split(flags, ",")
}

// SetCommandsFromString sets commands from a comma-separated string
func (s *Section) SetCommandsFromString(commands string) {
	if commands == "" {
		s.Commands = []string{}
		return
	}
	s.Commands = strings.Split(commands, ",")
}
