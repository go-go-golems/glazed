package cli

import (
	"github.com/pkg/errors"
	"github.com/wesen/glazed/pkg/formatters"
	"github.com/wesen/glazed/pkg/middlewares"
	"github.com/wesen/glazed/pkg/types"
	"gopkg.in/yaml.v3"
	"os"
	"strings"
)

// ParseTemplateFieldArguments parses a slice of --template-field arguments from the CLI.
//
//	--template-field '$fieldName:$template'
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
	ret := map[types.FieldName]string{}
	ret2 := map[string]interface{}{}
	fileContent, err := os.ReadFile(fileName)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to read file %s", fileName)
	}
	err = yaml.Unmarshal(fileContent, ret2)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to parse file %s", fileName)
	}

	for key, v := range ret2 {
		ret[key] = v.(string)
	}

	return ret, nil
}

type GlazeProcessor struct {
	of  formatters.OutputFormatter
	oms []middlewares.ObjectMiddleware
	// we keep an explicit reference to the column rename middleware because
	// we need to transform column names for middlewares that get added later on
	renameMiddlewares []*middlewares.RenameColumnMiddleware
}

func (gp *GlazeProcessor) AddTableMiddlewares(mws ...middlewares.TableMiddleware) {
	for _, m := range mws {
		switch m := m.(type) {
		case *middlewares.RenameColumnMiddleware:
			gp.renameMiddlewares = append(gp.renameMiddlewares, m)
		}
		gp.of.AddTableMiddleware(m)
	}
}

func (gp *GlazeProcessor) RenameColumns(columns []types.FieldName) []types.FieldName {
	for _, m := range gp.renameMiddlewares {
		orderedColumns, _ := m.RenameColumns(columns)
		columns = orderedColumns
	}

	return columns
}

func (gp *GlazeProcessor) OutputFormatter() formatters.OutputFormatter {
	return gp.of
}

func NewGlazeProcessor(of formatters.OutputFormatter, oms []middlewares.ObjectMiddleware) *GlazeProcessor {
	ret := &GlazeProcessor{
		of:                of,
		oms:               oms,
		renameMiddlewares: []*middlewares.RenameColumnMiddleware{},
	}

	// FIXME this is a bit ugly, and will need revisiting when we clean up the API
	// since so many middlewares rely on column renaming and ordering,
	// we need to remember column renames for future ordering middlwares.
	// This was surfaced by sqleton, which in order to preserve the column order that was
	// sent from the database, adds a ReorderColumns middleware.
	//
	// We need to rename the columns previous to create the middleware, since by the time the columns reach the middleware,
	// they will have been renamed. I wonder in this case if it shouldn't be possible to insert a middleware up front.
	//
	// In fact, let's add that functionality instead.

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
