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

	obscureJsonSection := &help.Section{
		Title:          "Obscuring JSON data",
		Slug:           "obscure-json",
		Short:          "Obscuring JSON data",
		Content:        `Information about obscuring JSON data`,
		Topics:         []string{"obscure-json"},
		Flags:          []string{"template", "template-field"},
		Commands:       []string{"json", "yaml"},
		SectionType:    help.SectionGeneralTopic,
		IsTopLevel:     true,
		IsTemplate:     false,
		ShowPerDefault: false,
	}

	templatesExample1 := &help.Section{
		Title:          "Example 1",
		Slug:           "templates-example-1",
		Short:          "glaze json foo.json --template '{{.foo}}'",
		Commands:       []string{"json"},
		Topics:         []string{"templates", "json-information"},
		Flags:          []string{"template"},
		SectionType:    help.SectionExample,
		ShowPerDefault: true,
		IsTemplate:     false,
		IsTopLevel:     false,
	}
	templatesExampleYAML1 := &help.Section{
		Title:          "Example 1",
		Slug:           "templates-yaml-example-1",
		Short:          "glaze yaml foo.yaml --template '{{.foo}}'",
		Commands:       []string{"yaml"},
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

	templatesExampleYAML2 := &help.Section{
		Title:          "Example 2",
		Slug:           "templates-yaml-example-2",
		Short:          "glaze yaml foo2.yaml --template '{{.foo}}' --template-field foo",
		Commands:       []string{"yaml"},
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
		IsTopLevel:     true,
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

	yamlInfoSection := &help.Section{
		Title:          "YAML",
		Slug:           "yaml-information",
		Content:        `Information about YAML long long long`,
		Short:          "Information about YAML short",
		Topics:         []string{"yaml-information"},
		Commands:       []string{"yaml"},
		SectionType:    help.SectionGeneralTopic,
		IsTemplate:     false,
		IsTopLevel:     true,
		ShowPerDefault: true,
	}

	dataDogApplicationSection := &help.Section{
		Title:          "DataDog Application",
		Slug:           "datadog-application",
		Content:        `Information about DataDog Application long long long`,
		Short:          "Information about DataDog Application short",
		Topics:         []string{"datadog-application", "templates"},
		Commands:       []string{"json"},
		SectionType:    help.SectionApplication,
		IsTemplate:     false,
		IsTopLevel:     true,
		ShowPerDefault: true,
	}

	jsonCleanupTutorialSection := &help.Section{
		Title:          "JSON Cleanup Tutorial",
		Slug:           "json-cleanup-tutorial",
		Content:        `Information about JSON Cleanup Tutorial long long long`,
		Short:          "Information about JSON Cleanup Tutorial short",
		Topics:         []string{"json-cleanup-tutorial", "templates"},
		Commands:       []string{"json"},
		SectionType:    help.SectionTutorial,
		IsTemplate:     false,
		IsTopLevel:     true,
		ShowPerDefault: true,
	}

	helpSystem := help.NewHelpSystem()
	helpSystem.AddSection(templatesSection)
	helpSystem.AddSection(obscureJsonSection)
	helpSystem.AddSection(templatesExample1)
	helpSystem.AddSection(templatesExample2)
	helpSystem.AddSection(templatesExample3)
	helpSystem.AddSection(templatesExampleYAML1)
	helpSystem.AddSection(templatesExampleYAML2)
	helpSystem.AddSection(jsonInfoSection)
	helpSystem.AddSection(yamlInfoSection)
	helpSystem.AddSection(dataDogApplicationSection)
	helpSystem.AddSection(jsonCleanupTutorialSection)

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
	rootCmd.AddCommand(cmds.DocsCmd)
}
