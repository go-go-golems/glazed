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
		Short:          "Using go templates",
		Content:        `Information about templates`,
		Topics:         []string{"templates"},
		Flags:          []string{"template", "template-field"},
		Commands:       []string{"yaml", "json"},
		SectionType:    help.SectionGeneralTopic,
		IsTopLevel:     true,
		IsTemplate:     false,
		ShowPerDefault: true,
	}
	templatesExample1 := &help.Section{
		Title:          "Example 1",
		Slug:           "templates-example-1",
		Short:          "glaze json foo.json --template '{{.foo}}'",
		Commands:       []string{"json"},
		Topics:         []string{"templates"},
		Flags:          []string{"template"},
		SectionType:    help.SectionExample,
		ShowPerDefault: true,
		IsTemplate:     false,
		IsTopLevel:     false,
	}

	templatesExample2 := &help.Section{
		Title:          "Example 2",
		Slug:           "templates-example-2",
		Short:          "glaze json foo2.json --template '{{.foo}}' --template-field foo",
		Commands:       []string{"json"},
		Topics:         []string{"templates"},
		Flags:          []string{"template", "template-field"},
		SectionType:    help.SectionExample,
		ShowPerDefault: false,
		IsTemplate:     false,
		IsTopLevel:     false,
	}

	templatesExample3 := &help.Section{
		Title:          "Example 3",
		Slug:           "templates-example-3",
		Short:          "glaze json foo3.json --template '{{.foo}}' --template-field foo",
		Commands:       []string{"json"},
		Topics:         []string{"templates"},
		Flags:          []string{"template", "template-field"},
		SectionType:    help.SectionExample,
		ShowPerDefault: true,
		IsTemplate:     false,
		IsTopLevel:     false,
	}

	jsonInfoSection := &help.Section{
		Title:          "JSON",
		Slug:           "json-information",
		Content:        `Information about JSON long long long`,
		Short:          "Information about JSON short",
		Topics:         []string{"json-information"},
		Commands:       []string{"json"},
		SectionType:    help.SectionGeneralTopic,
		IsTemplate:     false,
		IsTopLevel:     true,
		ShowPerDefault: true,
	}

	helpSystem := help.NewHelpSystem()
	helpSystem.AddSection(templatesSection)
	helpSystem.AddSection(templatesExample1)
	helpSystem.AddSection(templatesExample2)
	helpSystem.AddSection(templatesExample3)
	helpSystem.AddSection(jsonInfoSection)

	helpFunc, usageFunc := help.GetHelpUsageFuncs(helpSystem)
	helpTemplate, usageTemplate := help.GetHelpUsageTemplates(helpSystem)

	_ = usageFunc
	_ = usageTemplate

	rootCmd.SetHelpFunc(helpFunc)
	rootCmd.SetUsageFunc(usageFunc)
	rootCmd.SetHelpTemplate(helpTemplate)
	rootCmd.SetUsageTemplate(usageTemplate)

	helpCmd := help.NewCobraHelpCommand(helpSystem)
	rootCmd.SetHelpCommand(helpCmd)

	rootCmd.AddCommand(cmds.JsonCmd)
	rootCmd.AddCommand(cmds.YamlCmd)
}
