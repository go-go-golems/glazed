package cli

import (
	"github.com/go-go-golems/glazed/pkg/types"
	"github.com/spf13/cobra"
)

type SelectSettings struct {
	SelectField    string `glazed.parameter:"select-field"`
	SelectTemplate string `glazed.parameter:"select-template"`
}

func (ofs *OutputFormatterSettings) UpdateWithSelectSettings(ss *SelectSettings) {
	if ss.SelectField != "" || ss.SelectTemplate != "" {
		ofs.Output = "table"
		ofs.TableFormat = "tsv"
		ofs.FlattenObjects = true
		ofs.WithHeaders = false
	}
}

func (ffs *FieldsFilterSettings) UpdateWithSelectSettings(ss *SelectSettings) {
	if ss.SelectField != "" {
		ffs.Fields = []string{ss.SelectField}
	}
}

func (tf *TemplateSettings) UpdateWithSelectSettings(ss *SelectSettings) {
	if ss.SelectTemplate != "" {
		tf.Templates = map[types.FieldName]string{
			"_0": ss.SelectTemplate,
		}
	}
}

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

func AddSelectFlags(cmd *cobra.Command, defaults *SelectFlagsDefaults) error {
	cmd.Flags().String("select", defaults.Select, "Select a single field and output as a single line")
	cmd.Flags().String("select-template", defaults.SelectTemplate, "Output a single templated value for each row, on a single line")
	return nil
}

func ParseSelectFlags(cmd *cobra.Command) (*SelectSettings, error) {
	selectField, _ := cmd.Flags().GetString("select")
	selectTemplate, _ := cmd.Flags().GetString("select-template")

	return &SelectSettings{
		SelectField:    selectField,
		SelectTemplate: selectTemplate,
	}, nil
}
