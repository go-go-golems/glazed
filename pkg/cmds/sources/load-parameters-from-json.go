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
// into the standard layer map format: map[layerSlug]map[parameterName]value.
// The input is the raw unmarshaled config data (typically from JSON or YAML).
// The output should map layer slugs to parameter name/value pairs.
//
// Example: A flat config file like {"api-key": "secret", "threshold": 5}
// might be mapped to {"demo": {"api-key": "secret", "threshold": 5}}
type ConfigFileMapper func(rawConfig interface{}) (map[string]map[string]interface{}, error)

// FromFile loads parameter definitions from a JSON or YAML file and applies them to the parameter layers.
// By default, it expects the config file to have the structure:
//
//	layer-slug:
//	  parameter-name: value
//
// To use a custom config file structure, provide a ConfigFileMapper via WithConfigFileMapper.
func FromFile(filename string, options ...ConfigFileOption) Middleware {
	return func(next HandlerFunc) HandlerFunc {
		return func(layers_ *schema.Schema, parsedLayers *values.Values) error {
			err := next(layers_, parsedLayers)
			if err != nil {
				return err
			}

			opts := &configFileOptions{}
			for _, opt := range options {
				opt(opts)
			}

			m, err := readConfigFileToLayerMap(filename, opts.Mapper)
			if err != nil {
				return err
			}
			return updateFromMap(layers_, parsedLayers, m, opts.ParseOptions...)
		}
	}
}

// FromFiles applies a list of config files in order (low -> high precedence).
// Each file is applied as a separate step; callers may add metadata via options per-call.
// To use a custom config file structure, provide a ConfigFileMapper via WithConfigFileMapper.
func FromFiles(files []string, options ...ConfigFileOption) Middleware {
	return func(next HandlerFunc) HandlerFunc {
		return func(layers_ *schema.Schema, parsedLayers *values.Values) error {
			if err := next(layers_, parsedLayers); err != nil {
				return err
			}
			opts := &configFileOptions{}
			for _, opt := range options {
				opt(opts)
			}
			for i, f := range files {
				m, err := readConfigFileToLayerMap(f, opts.Mapper)
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
				if err := updateFromMap(layers_, parsedLayers, m, parseOpts...); err != nil {
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
// into the standard layer map format. If not provided, the default behavior expects:
//
//	layer-slug:
//	  parameter-name: value
func WithConfigFileMapper(mapper ConfigFileMapper) ConfigFileOption {
	return func(o *configFileOptions) {
		if mapper != nil {
			o.Mapper = mapper
		}
	}
}

// WithConfigMapper provides a ConfigMapper (pattern-based or function-based) to transform
// config file structures into the standard layer map format.
// This is the same as WithConfigFileMapper but accepts the ConfigMapper interface directly.
func WithConfigMapper(mapper ConfigMapper) ConfigFileOption {
	return func(o *configFileOptions) {
		o.Mapper = mapper
	}
}

// WithParseOptions adds parse step options that will be applied when loading parameters from the config file.
func WithParseOptions(options ...fields.ParseOption) ConfigFileOption {
	return func(o *configFileOptions) {
		o.ParseOptions = append(o.ParseOptions, options...)
	}
}

func readConfigFileToLayerMap(filename string, mapper ConfigMapper) (map[string]map[string]interface{}, error) {
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

	// Default behavior: expect the standard layer map structure
	if m, ok := rawData.(map[string]interface{}); ok {
		result := make(map[string]map[string]interface{})
		for layerSlug, layerData := range m {
			if layerMap, ok := layerData.(map[string]interface{}); ok {
				result[layerSlug] = layerMap
			} else {
				return nil, errors.Errorf("expected map[string]interface{} for layer %s, got %T", layerSlug, layerData)
			}
		}
		return result, nil
	}

	return nil, errors.Errorf("expected map[string]interface{} for config file, got %T", rawData)
}
