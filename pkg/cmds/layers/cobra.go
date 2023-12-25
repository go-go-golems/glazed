package layers

import (
	"github.com/go-go-golems/glazed/pkg/cmds/parameters"
	"github.com/spf13/cobra"
)

type CobraParameterLayer interface {
	ParameterLayer
	// AddLayerToCobraCommand adds all the flags and arguments defined in this layer to the given cobra command.
	AddLayerToCobraCommand(cmd *cobra.Command) error
	ParseLayerFromCobraCommand(cmd *cobra.Command, options ...parameters.ParseStepOption) (*ParsedLayer, error)
}
