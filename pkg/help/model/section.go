package model

import (
	"fmt"
	"github.com/go-go-golems/glazed/pkg/helpers/yaml"
	yaml2 "gopkg.in/yaml.v3"
	"os"
	"regexp"
	"strings"
)

// Section represents a help/documentation section.
type Section struct {
	ID             int64       `yaml:"id,omitempty"`
	Slug           string      `yaml:"slug,omitempty"`
	Title          string      `yaml:"title,omitempty"`
	Subtitle       string      `yaml:"subtitle,omitempty"`
	Short          string      `yaml:"short,omitempty"`
	Content        string      `yaml:"content,omitempty"`
	SectionType    SectionType `yaml:"sectionType,omitempty"`
	IsTopLevel     bool        `yaml:"isTopLevel,omitempty"`
	IsTemplate     bool        `yaml:"isTemplate,omitempty"`
	ShowPerDefault bool        `yaml:"showDefault,omitempty"`
	Order          int         `yaml:"ord,omitempty"`
	Topics         []string    `yaml:"topics,omitempty"`
	Flags          []string    `yaml:"flags,omitempty"`
	Commands       []string    `yaml:"commands,omitempty"`
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

// LoadSectionFromMarkdown loads a Section from a markdown file with YAML front-matter.
func LoadSectionFromMarkdown(path string) (*Section, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read file: %w", err)
	}
	content := string(b)

	// Extract YAML front-matter
	re := regexp.MustCompile(`(?s)^---\n(.*?)\n---\n(.*)$`)
	matches := re.FindStringSubmatch(content)
	if len(matches) != 3 {
		return nil, fmt.Errorf("no YAML front-matter found in %s", path)
	}
	yamlStr := matches[1]
	body := matches[2]

	// Parse YAML into Section
	sec := &Section{}
	yamlStr = yaml.Clean(yamlStr, false)
	if err := yaml2.Unmarshal([]byte(yamlStr), sec); err != nil {
		return nil, fmt.Errorf("unmarshal yaml: %w", err)
	}
	sec.Content = strings.TrimSpace(body)
	return sec, nil
}
