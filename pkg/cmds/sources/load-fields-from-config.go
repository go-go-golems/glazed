package sources

import (
	"encoding/json"
	fields "github.com/go-go-golems/glazed/pkg/cmds/fields"
	"github.com/go-go-golems/glazed/pkg/cmds/schema"
	"github.com/go-go-golems/glazed/pkg/cmds/values"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
	"os"
	"strings"
)

// ConfigFileMapper is a function that transforms an arbitrary config file structure
// into the standard section map format: map[sectionSlug]map[fieldName]value.
// The input is the raw unmarshaled config data (typically from JSON or YAML).
// The output should map section slugs to field name/value pairs.
//
// Example: A flat config file like {"api-key": "secret", "threshold": 5}
// might be mapped to {"demo": {"api-key": "secret", "threshold": 5}}
type ConfigFileMapper func(rawConfig interface{}) (map[string]map[string]interface{}, error)

// FromFile loads field definitions from a JSON or YAML file and applies them to the schema.
// By default, it expects the config file to have the structure:
//
//	section-slug:
//	  field-name: value
//
// To use a custom config file structure, provide a ConfigFileMapper via WithConfigFileMapper.
func FromFile(filename string, options ...ConfigFileOption) Middleware {
	return func(next HandlerFunc) HandlerFunc {
		return func(schema_ *schema.Schema, parsedValues *values.Values) error {
			err := next(schema_, parsedValues)
			if err != nil {
				return err
			}

			opts := &configFileOptions{}
			for _, opt := range options {
				opt(opts)
			}

			m, err := readConfigFileToSectionMap(filename, opts.Mapper)
			if err != nil {
				return err
			}
			return updateFromMap(schema_, parsedValues, m, opts.ParseOptions...)
		}
	}
}

// FromFiles applies a list of config files in order (low -> high precedence).
// Each file is applied as a separate step; callers may add metadata via options per-call.
// To use a custom config file structure, provide a ConfigFileMapper via WithConfigFileMapper.
func FromFiles(files []string, options ...ConfigFileOption) Middleware {
	return func(next HandlerFunc) HandlerFunc {
		return func(schema_ *schema.Schema, parsedValues *values.Values) error {
			if err := next(schema_, parsedValues); err != nil {
				return err
			}
			opts := &configFileOptions{}
			for _, opt := range options {
				opt(opts)
			}
			for i, f := range files {
				m, err := readConfigFileToSectionMap(f, opts.Mapper)
				if err != nil {
					return err
				}
				parseOpts := append(opts.ParseOptions,
					fields.WithSource("config"),
					fields.WithMetadata(map[string]interface{}{
						"config_file": f,
						"index":       i,
					}),
				)
				if err := updateFromMap(schema_, parsedValues, m, parseOpts...); err != nil {
					return err
				}
			}
			return nil
		}
	}
}

type configFileOptions struct {
	Mapper       ConfigMapper // Accept both ConfigFileMapper and pattern mapper
	ParseOptions []fields.ParseOption
}

// ConfigFileOption configures behavior for FromFile and FromFiles.
type ConfigFileOption func(*configFileOptions)

// WithConfigFileMapper provides a custom mapper function to transform arbitrary config file structures
// into the standard section map format. If not provided, the default behavior expects:
//
//	section-slug:
//	  field-name: value
func WithConfigFileMapper(mapper ConfigFileMapper) ConfigFileOption {
	return func(o *configFileOptions) {
		if mapper != nil {
			o.Mapper = mapper
		}
	}
}

// WithConfigMapper provides a ConfigMapper (pattern-based or function-based) to transform
// config file structures into the standard section map format.
// This is the same as WithConfigFileMapper but accepts the ConfigMapper interface directly.
func WithConfigMapper(mapper ConfigMapper) ConfigFileOption {
	return func(o *configFileOptions) {
		o.Mapper = mapper
	}
}

// WithParseOptions adds parse step options that will be applied when loading fields from the config file.
func WithParseOptions(options ...fields.ParseOption) ConfigFileOption {
	return func(o *configFileOptions) {
		o.ParseOptions = append(o.ParseOptions, options...)
	}
}

func readConfigFileToSectionMap(filename string, mapper ConfigMapper) (map[string]map[string]interface{}, error) {
	var rawData interface{}
	switch {
	case strings.HasSuffix(filename, ".json"):
		bytes, err := os.ReadFile(filename)
		if err != nil {
			return nil, err
		}
		if err := json.Unmarshal(bytes, &rawData); err != nil {
			return nil, err
		}
	case strings.HasSuffix(filename, ".yaml"), strings.HasSuffix(filename, ".yml"):
		bytes, err := os.ReadFile(filename)
		if err != nil {
			return nil, err
		}
		if err := yaml.Unmarshal(bytes, &rawData); err != nil {
			return nil, err
		}
	default:
		return nil, errors.New("unsupported file type")
	}

	// If a custom mapper is provided, use it to transform the config structure
	if mapper != nil {
		return mapper.Map(rawData)
	}

	// Default behavior: expect the standard section map structure
	if m, ok := rawData.(map[string]interface{}); ok {
		result := make(map[string]map[string]interface{})
		for sectionSlug, sectionData := range m {
			if sectionMap, ok := sectionData.(map[string]interface{}); ok {
				result[sectionSlug] = sectionMap
			} else {
				return nil, errors.Errorf("expected map[string]interface{} for section %s, got %T", sectionSlug, sectionData)
			}
		}
		return result, nil
	}

	return nil, errors.Errorf("expected map[string]interface{} for config file, got %T", rawData)
}
