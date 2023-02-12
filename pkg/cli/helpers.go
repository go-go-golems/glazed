package cli

import (
	"github.com/go-go-golems/glazed/pkg/helpers"
	"github.com/pkg/errors"
	"strings"
)

// TODO(manuel, 2023-02-11) This should be its distinct ParameterType
//
// See https://github.com/go-go-golems/glazed/issues/128

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
