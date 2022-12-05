package main

import (
	"embed"
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

//go:embed doc/*
var docFS embed.FS

func init() {
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
	err := helpSystem.LoadSectionsFromEmbedFS(docFS, ".")
	if err != nil {
		panic(err)
	}
	helpSystem.AddSection(obscureJsonSection)
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
