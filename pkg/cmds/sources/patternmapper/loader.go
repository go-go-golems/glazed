package patternmapper

import (
	"github.com/go-go-golems/glazed/pkg/cmds/schema"
	sources "github.com/go-go-golems/glazed/pkg/cmds/sources"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
	"io"
	"os"
)

// mappingFile represents the top-level structure of a YAML/JSON mapping file.
// It supports either an object with a top-level "mappings" array or a bare array of rules.
type mappingFile struct {
	Mappings []mappingRule `yaml:"mappings" json:"mappings"`
}

// mappingRule is an unmarshalling helper for YAML/JSON that mirrors MappingRule but
// uses snake_case keys as typically found in config files.
type mappingRule struct {
	Source        string        `yaml:"source" json:"source"`
	TargetSection string        `yaml:"target_section" json:"target_section"`
	TargetField   string        `yaml:"target_field" json:"target_field"`
	Required      bool          `yaml:"required" json:"required"`
	Rules         []mappingRule `yaml:"rules" json:"rules"`
}

func (mr mappingRule) toMappingRule() MappingRule {
	r := MappingRule{
		Source:        mr.Source,
		TargetSection: mr.TargetSection,
		TargetField:   mr.TargetField,
		Required:      mr.Required,
	}
	if len(mr.Rules) > 0 {
		r.Rules = make([]MappingRule, 0, len(mr.Rules))
		for _, cr := range mr.Rules {
			r.Rules = append(r.Rules, cr.toMappingRule())
		}
	}
	return r
}

// LoadRulesFromReader reads YAML/JSON mapping rules from an io.Reader.
// It accepts both a top-level object with a "mappings" array and a bare array of rules.
func LoadRulesFromReader(r io.Reader) ([]MappingRule, error) {
	var data interface{}
	dec := yaml.NewDecoder(r)
	if err := dec.Decode(&data); err != nil {
		return nil, errors.Wrap(err, "failed to decode mapping file")
	}

	// Re-encode and decode into expected structs to handle YAML/JSON flexibly
	// First attempt: object with "mappings"
	var mf mappingFile
	// yaml.v3 can marshal round-trip generic interfaces
	b, err := yaml.Marshal(data)
	if err != nil {
		return nil, errors.Wrap(err, "failed to marshal intermediate mapping data")
	}
	_ = yaml.Unmarshal(b, &mf)
	var rules []mappingRule
	if len(mf.Mappings) > 0 {
		rules = mf.Mappings
	} else {
		// Second attempt: bare array of rules
		var arr []mappingRule
		if err := yaml.Unmarshal(b, &arr); err != nil {
			return nil, errors.Errorf("mapping file must contain 'mappings' array or be an array of rules: %v", err)
		}
		rules = arr
	}

	out := make([]MappingRule, 0, len(rules))
	for _, rr := range rules {
		out = append(out, rr.toMappingRule())
	}
	return out, nil
}

// LoadRulesFromFile reads YAML/JSON mapping rules from a file path.
func LoadRulesFromFile(filename string) ([]MappingRule, error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to open mapping file %q", filename)
	}
	defer func() {
		if closeErr := f.Close(); closeErr != nil {
			// Log error if closing fails, but don't return it as we're in a defer
			_ = closeErr
		}
	}()
	return LoadRulesFromReader(f)
}

// LoadMapperFromFile loads a ConfigMapper from a YAML/JSON mapping file using the provided schema.
func LoadMapperFromFile(sectionSchema *schema.Schema, filename string) (sources.ConfigMapper, error) {
	rules, err := LoadRulesFromFile(filename)
	if err != nil {
		return nil, err
	}
	return NewConfigMapper(sectionSchema, rules...)
}
