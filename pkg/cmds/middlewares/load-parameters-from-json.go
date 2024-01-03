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

			m := map[string]map[string]interface{}{}
			switch {

			case strings.HasSuffix(filename, ".json"):
				bytes, err := os.ReadFile(filename)
				if err != nil {
					return err
				}
				err = json.Unmarshal(bytes, &m)
				if err != nil {
					return err
				}

			case strings.HasSuffix(filename, ".yaml"),
				strings.HasSuffix(filename, ".yml"):
				bytes, err := os.ReadFile(filename)
				if err != nil {
					return err
				}
				err = yaml.Unmarshal(bytes, &m)
				if err != nil {
					return err
				}

			default:
				return errors.New("unsupported file type")
			}

			return updateFromMap(layers_, parsedLayers, m, options...)
		}
	}
}
