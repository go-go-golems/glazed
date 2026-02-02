package schema

import (
	"github.com/go-go-golems/glazed/pkg/cmds/fields"
	"github.com/go-go-golems/glazed/pkg/cmds/values"
	"github.com/spf13/cobra"
)

type CobraSection interface {
	Section
	// AddLayerToCobraCommand adds all the flags and arguments defined in this layer to the given cobra command.
	AddLayerToCobraCommand(cmd *cobra.Command) error
	ParseLayerFromCobraCommand(cmd *cobra.Command, options ...fields.ParseOption) (*values.SectionValues, error)
}
