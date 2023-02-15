package cli

import (
	"github.com/go-go-golems/glazed/pkg/formatters"
	"github.com/go-go-golems/glazed/pkg/middlewares"
	"github.com/spf13/cobra"
	"os"
)

type ReplaceSettings struct {
	ReplaceFile string
}

func (rs *ReplaceSettings) AddMiddlewares(of formatters.OutputFormatter) error {
	if rs.ReplaceFile != "" {
		b, err := os.ReadFile(rs.ReplaceFile)
		if err != nil {
			return err
		}

		mw, err := middlewares.NewReplaceMiddlewareFromYAML(b)
		if err != nil {
			return err
		}

		of.AddTableMiddleware(mw)
	}

	return nil
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

func AddReplaceFlags(cmd *cobra.Command, defaults *ReplaceFlagsDefaults) error {
	cmd.Flags().String("replace-file", defaults.ReplaceFile, "File with replacements")
	return nil
}

func ParseReplaceFlags(cmd *cobra.Command) (*ReplaceSettings, error) {
	replaceFile, _ := cmd.Flags().GetString("replace-file")

	return &ReplaceSettings{
		ReplaceFile: replaceFile,
	}, nil
}
