package middlewares

import (
	"encoding/json"
	"github.com/go-go-golems/glazed/pkg/cmds/layers"
	"github.com/go-go-golems/glazed/pkg/cmds/parameters"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
	"os"
	"strings"
)

// LoadParametersFromFile loads parameter definitions from a JSON file and applies them to the parameter layers.
func LoadParametersFromFile(filename string, options ...parameters.ParseStepOption) Middleware {
	return func(next HandlerFunc) HandlerFunc {
		return func(layers_ *layers.ParameterLayers, parsedLayers *layers.ParsedLayers) error {
			err := next(layers_, parsedLayers)
			if err != nil {
				return err
			}

			m, err := readConfigFileToLayerMap(filename)
			if err != nil {
				return err
			}
			return updateFromMap(layers_, parsedLayers, m, options...)
		}
	}
}

// LoadParametersFromFiles applies a list of config files in order (low -> high precedence).
// Each file is applied as a separate step; callers may add metadata via options per-call.
func LoadParametersFromFiles(files []string, options ...parameters.ParseStepOption) Middleware {
	return func(next HandlerFunc) HandlerFunc {
		return func(layers_ *layers.ParameterLayers, parsedLayers *layers.ParsedLayers) error {
			if err := next(layers_, parsedLayers); err != nil {
				return err
			}
			for i, f := range files {
				m, err := readConfigFileToLayerMap(f)
				if err != nil {
					return err
				}
				opts := append(options,
					parameters.WithParseStepSource("config"),
					parameters.WithParseStepMetadata(map[string]interface{}{
						"config_file": f,
						"index":       i,
					}),
				)
				if err := updateFromMap(layers_, parsedLayers, m, opts...); err != nil {
					return err
				}
			}
			return nil
		}
	}
}

func readConfigFileToLayerMap(filename string) (map[string]map[string]interface{}, error) {
	m := map[string]map[string]interface{}{}
	switch {
	case strings.HasSuffix(filename, ".json"):
		bytes, err := os.ReadFile(filename)
		if err != nil {
			return nil, err
		}
		if err := json.Unmarshal(bytes, &m); err != nil {
			return nil, err
		}
	case strings.HasSuffix(filename, ".yaml"), strings.HasSuffix(filename, ".yml"):
		bytes, err := os.ReadFile(filename)
		if err != nil {
			return nil, err
		}
		if err := yaml.Unmarshal(bytes, &m); err != nil {
			return nil, err
		}
	default:
		return nil, errors.New("unsupported file type")
	}
	return m, nil
}
