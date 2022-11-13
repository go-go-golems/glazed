package cli

import (
	"dd-cli/pkg/types"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
	"io/ioutil"
	"os"
	"strings"
)

// ParseTemplateFieldArguments parses a slice of --template-field arguments from the CLI.
//
//   --template-field '$fieldName:$template'
func ParseTemplateFieldArguments(templateArguments []string) (map[types.FieldName]string, error) {
	ret := map[types.FieldName]string{}
	for i, templateArgument := range templateArguments {
		if strings.HasPrefix(templateArgument, "@") {
			ret_, err := ParseTemplateFieldFileArgument(templateArgument[1:])
			if err != nil {
				return nil, errors.Wrapf(err, "failed to parse template field file argument %d", i)
			}

			for key, value := range ret_ {
				ret[key] = value
			}
		} else {
			fieldName, template, ok := strings.Cut(templateArgument, ":")
			if !ok {
				return nil, errors.Errorf("invalid template argument %d: %s", i, templateArgument)
			}
			ret[fieldName] = template
		}
	}
	return ret, nil
}

// ParseTemplateFieldFileArgument loads the given file, which must be a yaml file containing a string: string
// dictionary. The keys will be the resulting fields, while the values are the templates to be evaluated.
func ParseTemplateFieldFileArgument(fileName string) (map[types.FieldName]string, error) {
	// check file exists
	_, err := os.Stat(fileName)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to stat file %s", fileName)
	}

	// parse yaml file
	var ret map[types.FieldName]string
	fileContent, err := ioutil.ReadFile(fileName)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to read file %s", fileName)
	}
	err = yaml.Unmarshal(fileContent, ret)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to parse file %s", fileName)
	}

	return ret, nil
}
