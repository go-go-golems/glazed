package cli

import (
	_ "embed"
	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/formatters"
	"github.com/go-go-golems/glazed/pkg/middlewares"
	"github.com/go-go-golems/glazed/pkg/types"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"regexp"
	"strings"
)

// Helpers for cobra commands

type SelectFlagsDefaults struct {
	Select         string
	SelectTemplate string
}

func NewSelectFlagsDefaults() *SelectFlagsDefaults {
	return &SelectFlagsDefaults{
		Select:         "",
		SelectTemplate: "",
	}
}

func AddSelectFlags(cmd *cobra.Command, defaults *SelectFlagsDefaults) {
	cmd.Flags().String("select", defaults.Select, "Select a single field and output as a single line")
	cmd.Flags().String("select-template", defaults.SelectTemplate, "Output a single templated value for each row, on a single line")
}

func ParseSelectFlags(cmd *cobra.Command) (*SelectSettings, error) {
	selectField, _ := cmd.Flags().GetString("select")
	selectTemplate, _ := cmd.Flags().GetString("select-template")

	return &SelectSettings{
		SelectField:    selectField,
		SelectTemplate: selectTemplate,
	}, nil
}

type ReplaceFlagsDefaults struct {
	// currently, only support loading replacements from a file
	ReplaceFile string
}

func NewReplaceFlagsDefaults() *ReplaceFlagsDefaults {
	return &ReplaceFlagsDefaults{
		ReplaceFile: "",
	}
}

func AddReplaceFlags(cmd *cobra.Command, defaults *ReplaceFlagsDefaults) {
	cmd.Flags().String("replace-file", defaults.ReplaceFile, "File with replacements")
}

func ParseReplaceFlags(cmd *cobra.Command) (*ReplaceSettings, error) {
	replaceFile, _ := cmd.Flags().GetString("replace-file")

	return &ReplaceSettings{
		ReplaceFile: replaceFile,
	}, nil
}

type RenameFlagsDefaults struct {
	Rename       []string
	RenameRegexp []string
	RenameYaml   string
}

func NewRenameFlagsDefaults() *RenameFlagsDefaults {
	return &RenameFlagsDefaults{
		Rename:       []string{},
		RenameRegexp: []string{},
		RenameYaml:   "",
	}
}

func AddRenameFlags(cmd *cobra.Command, defaults *RenameFlagsDefaults) {
	cmd.Flags().StringSlice("rename", defaults.Rename, "Rename fields (list of oldName:newName)")
	cmd.Flags().StringSlice("rename-regexp", defaults.RenameRegexp, "Rename fields using regular expressions (list of regex:newName)")
	cmd.Flags().String("rename-yaml", defaults.RenameYaml, "Rename fields using a yaml file")
}

func ParseRenameFlags(cmd *cobra.Command) (*RenameSettings, error) {
	renameFields, _ := cmd.Flags().GetStringSlice("rename")
	renameRegexpFields, _ := cmd.Flags().GetStringSlice("rename-regexp")
	renameYaml, _ := cmd.Flags().GetString("rename-yaml")

	renamesFieldsMap := map[types.FieldName]types.FieldName{}
	for _, renameField := range renameFields {
		parts := strings.Split(renameField, ":")
		if len(parts) != 2 {
			return nil, errors.Errorf("Invalid rename field: %s", renameField)
		}
		renamesFieldsMap[types.FieldName(parts[0])] = types.FieldName(parts[1])
	}

	regexpReplacements := middlewares.RegexpReplacements{}
	for _, renameRegexpField := range renameRegexpFields {
		parts := strings.Split(renameRegexpField, ":")
		if len(parts) != 2 {
			return nil, errors.Errorf("Invalid rename-regexp field: %s", renameRegexpField)
		}
		re, err := regexp.Compile(parts[0])
		if err != nil {
			return nil, errors.Wrapf(err, "Invalid regexp: %s", parts[0])
		}
		regexpReplacements = append(regexpReplacements,
			&middlewares.RegexpReplacement{Regexp: re, Replacement: parts[1]})
	}

	return &RenameSettings{
		RenameFields:  renamesFieldsMap,
		RenameRegexps: regexpReplacements,
		YamlFile:      renameYaml,
	}, nil
}

// TODO(manuel, 2022-11-20) Make it easy for the developer to configure which flag they want
// and which they don't

type FieldsFilterFlagsDefaults struct {
	Fields      string
	Filter      string
	SortColumns bool
}

func NewFieldsFilterFlagsDefaults() *FieldsFilterFlagsDefaults {
	return &FieldsFilterFlagsDefaults{
		Fields:      "",
		Filter:      "",
		SortColumns: false,
	}
}

// AddFieldsFilterFlags adds the flags for the following middlewares to the cmd:
// - FieldsFilterMiddleware
// - SortColumnsMiddleware
// - ReorderColumnOrderMiddleware
func AddFieldsFilterFlags(cmd *cobra.Command, defaults *FieldsFilterFlagsDefaults) {
	defaultFieldHelp := defaults.Fields
	if defaultFieldHelp == "" {
		defaultFieldHelp = "all"
	}
	cmd.Flags().String("fields", defaults.Fields, "Fields to include in the output, default: "+defaultFieldHelp)
	cmd.Flags().String("filter", defaults.Filter, "Fields to remove from output")
	cmd.Flags().Bool("sort-columns", defaults.SortColumns, "Sort columns alphabetically")
}

func ParseFieldsFilterFlags(cmd *cobra.Command) (*FieldsFilterSettings, error) {
	fieldStr := cmd.Flag("fields").Value.String()
	filters := []string{}
	fields := []string{}
	if fieldStr != "" {
		fields = strings.Split(fieldStr, ",")
	}
	if cmd.Flag("fields").Changed && !cmd.Flag("filter").Changed {
		filters = []string{}
	} else {
		filterStr := cmd.Flag("filter").Value.String()
		if filterStr != "" {
			filters = strings.Split(filterStr, ",")
		}
	}

	sortColumns, err := cmd.Flags().GetBool("sort-columns")
	if err != nil {
		return nil, err
	}

	return &FieldsFilterSettings{
		Fields:         fields,
		Filters:        filters,
		SortColumns:    sortColumns,
		ReorderColumns: fields,
	}, nil
}

type FlagsDefaults struct {
	Output       *OutputFlagsDefaults
	Select       *SelectFlagsDefaults
	Rename       *RenameFlagsDefaults
	Template     *TemplateFlagsDefaults
	FieldsFilter *FieldsFilterFlagsDefaults
	Replace      *ReplaceFlagsDefaults
}

func NewFlagsDefaults() *FlagsDefaults {
	return &FlagsDefaults{
		Output:       NewOutputFlagsDefaults(),
		Select:       NewSelectFlagsDefaults(),
		Rename:       NewRenameFlagsDefaults(),
		Template:     NewTemplateFlagsDefaults(),
		FieldsFilter: NewFieldsFilterFlagsDefaults(),
		Replace:      NewReplaceFlagsDefaults(),
	}
}

func AddFlags(cmd *cobra.Command, defaults *FlagsDefaults) {
	AddOutputFlags(cmd)
	AddSelectFlags(cmd, defaults.Select)
	AddRenameFlags(cmd, defaults.Rename)
	AddTemplateFlags(cmd)
	AddFieldsFilterFlags(cmd, defaults.FieldsFilter)
	AddReplaceFlags(cmd, defaults.Replace)
}

func SetupProcessor(cmd *cobra.Command) (*cmds.GlazeProcessor, formatters.OutputFormatter, error) {
	outputSettings, err := ParseOutputFlags(cmd)
	if err != nil {
		return nil, nil, errors.Wrapf(err, "Error parsing output flags")
	}

	templateSettings, err := ParseTemplateFlags(cmd)
	if err != nil {
		return nil, nil, errors.Wrapf(err, "Error parsing template flags")
	}

	fieldsFilterSettings, err := ParseFieldsFilterFlags(cmd)
	if err != nil {
		return nil, nil, errors.Wrapf(err, "Error parsing fields filter flags")
	}

	selectSettings, err := ParseSelectFlags(cmd)
	if err != nil {
		return nil, nil, errors.Wrapf(err, "Error parsing select flags")
	}
	outputSettings.UpdateWithSelectSettings(selectSettings)
	fieldsFilterSettings.UpdateWithSelectSettings(selectSettings)
	templateSettings.UpdateWithSelectSettings(selectSettings)

	renameSettings, err := ParseRenameFlags(cmd)
	if err != nil {
		return nil, nil, errors.Wrapf(err, "Error parsing rename flags")
	}

	replaceSettings, err := ParseReplaceFlags(cmd)
	if err != nil {
		return nil, nil, errors.Wrapf(err, "Error parsing replace flags")
	}

	of, err := outputSettings.CreateOutputFormatter()
	if err != nil {
		return nil, nil, errors.Wrapf(err, "Error creating output formatter")
	}

	// rename middlewares run first because they are used to clean up column names
	// for the following middlewares too.
	// these following middlewares can create proper column names on their own
	// when needed
	err = renameSettings.AddMiddlewares(of)
	if err != nil {
		return nil, nil, errors.Wrapf(err, "Error adding rename middlewares")
	}

	err = templateSettings.AddMiddlewares(of)
	if err != nil {
		return nil, nil, errors.Wrapf(err, "Error adding template middlewares")
	}

	if (outputSettings.Output == "json" || outputSettings.Output == "yaml") && outputSettings.FlattenObjects {
		mw := middlewares.NewFlattenObjectMiddleware()
		of.AddTableMiddleware(mw)
	}
	fieldsFilterSettings.AddMiddlewares(of)

	err = replaceSettings.AddMiddlewares(of)
	if err != nil {
		return nil, nil, errors.Wrapf(err, "Error adding replace middlewares")
	}

	var middlewares_ []middlewares.ObjectMiddleware
	if !templateSettings.UseRowTemplates && len(templateSettings.Templates) > 0 {
		ogtm, err := middlewares.NewObjectGoTemplateMiddleware(templateSettings.Templates)
		if err != nil {
			return nil, nil, errors.Wrapf(err, "Could not process template argument")
		}
		middlewares_ = append(middlewares_, ogtm)
	}

	gp := cmds.NewGlazeProcessor(of, middlewares_)
	return gp, of, nil
}
