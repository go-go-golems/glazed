package store

import (
	"fmt"
	"os"
	"regexp"
	"strings"
	"gopkg.in/yaml.v3"
	"github.com/go-go-golems/glazed/pkg/help/model"
	ghyaml "github.com/go-go-golems/glazed/pkg/helpers/yaml"
)

// LoadSectionFromMarkdown loads a Section from a markdown file with YAML front-matter.
func LoadSectionFromMarkdown(path string) (*model.Section, error) {
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
	sec := &model.Section{}
	yamlStr = ghyaml.Clean(yamlStr, false)
	if err := yaml.Unmarshal([]byte(yamlStr), sec); err != nil {
		return nil, fmt.Errorf("unmarshal yaml: %w", err)
	}
	sec.Content = strings.TrimSpace(body)
	return sec, nil
} 