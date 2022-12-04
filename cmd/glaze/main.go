package main

import (
	"github.com/spf13/cobra"
	"github.com/wesen/glazed/cmd/glaze/cmds"
	"github.com/wesen/glazed/pkg/help"
)

var rootCmd = &cobra.Command{
	Use:   "glaze",
	Short: "glaze is a tool to format structured data",
}

func main() {
	_ = rootCmd.Execute()
}

func init() {
	templatesSection := &help.Section{
		Title:          "Using go templates",
		Slug:           "templates",
		Content:        `Information about templates`,
		SubSections:    []*help.Section{},
		Tags:           []string{"templates"},
		IsTemplate:     false,
		ShowPerDefault: true,
	}

	sections := []*help.Section{templatesSection}
	helpFunc, usageFunc := help.GetHelpUsageFuncs(sections)
	helpTemplate, usageTemplate := help.GetHelpUsageTemplates(sections)

	_ = usageFunc
	_ = usageTemplate

	rootCmd.SetHelpFunc(helpFunc)
	rootCmd.SetUsageFunc(usageFunc)
	rootCmd.SetHelpTemplate(helpTemplate)
	rootCmd.SetUsageTemplate(usageTemplate)

	helpCmd := help.NewCobraHelpCommand(sections)
	rootCmd.SetHelpCommand(helpCmd)

	rootCmd.AddCommand(cmds.JsonCmd)
	rootCmd.AddCommand(cmds.YamlCmd)
}
