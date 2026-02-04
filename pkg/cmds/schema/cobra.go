package schema

import (
	"github.com/go-go-golems/glazed/pkg/cmds/fields"
	"github.com/go-go-golems/glazed/pkg/cmds/values"
	"github.com/spf13/cobra"
)

type CobraSection interface {
	Section
	// AddSectionToCobraCommand adds all the flags and arguments defined in this section to the given cobra command.
	AddSectionToCobraCommand(cmd *cobra.Command) error
	ParseSectionFromCobraCommand(cmd *cobra.Command, options ...fields.ParseOption) (*values.SectionValues, error)
}
