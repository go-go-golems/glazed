package cli

import (
	"github.com/go-go-golems/glazed/pkg/formatters"
	"github.com/go-go-golems/glazed/pkg/helpers"
	"github.com/go-go-golems/glazed/pkg/middlewares"
	"github.com/go-go-golems/glazed/pkg/types"
	"github.com/pkg/errors"
	"strings"
)

type GlazeProcessor struct {
	of  formatters.OutputFormatter
	oms []middlewares.ObjectMiddleware
}

func (gp *GlazeProcessor) OutputFormatter() formatters.OutputFormatter {
	return gp.of
}

func NewGlazeProcessor(of formatters.OutputFormatter, oms []middlewares.ObjectMiddleware) *GlazeProcessor {
	ret := &GlazeProcessor{
		of:  of,
		oms: oms,
	}

	return ret
}

// TODO(2022-12-18, manuel) we should actually make it possible to order the columns
// https://github.com/wesen/glazed/issues/56
func (gp *GlazeProcessor) ProcessInputObject(obj map[string]interface{}) error {
	for _, om := range gp.oms {
		obj2, err := om.Process(obj)
		if err != nil {
			return err
		}
		obj = obj2
	}

	gp.of.AddRow(&types.SimpleRow{Hash: obj})
	return nil
}

func ParseCLIKeyValueData(keyValues []string) (map[string]interface{}, error) {
	templateData := map[string]interface{}{}

	for _, keyValue := range keyValues {
		// check if keyValues starts with @ and load as a file
		if strings.HasPrefix(keyValue, "@") {
			templateDataFile := keyValue[1:]

			if strings.HasSuffix(templateDataFile, ".json") {
				err := helpers.LoadJSONFile(templateDataFile, &templateData)
				if err != nil {
					return nil, errors.Wrapf(err, "Error loading template data from file %s", templateDataFile)
				}
			} else if strings.HasSuffix(templateDataFile, ".yaml") || strings.HasSuffix(templateDataFile, ".yml") {
				err := helpers.LoadYAMLFile(templateDataFile, &templateData)
				if err != nil {
					return nil, errors.Wrapf(err, "Error loading template data from file %s", templateDataFile)
				}
			} else {
				return nil, errors.Errorf("Unknown template data file format for file %s", templateDataFile)
			}

		} else {
			key, value, ok := strings.Cut(keyValue, ":")
			if !ok {
				return nil, errors.Errorf("Invalid template data %s", keyValue)
			}
			templateData[key] = value
		}
	}
	return templateData, nil
}
