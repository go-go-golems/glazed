package help

import (
	"fmt"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"glazed/pkg/helpers"
	"strings"
	"text/template"
)

type HelpFunc = func(c *cobra.Command, args []string)
type UsageFunc = func(c *cobra.Command) error

func GetHelpUsageFuncs(sections []*Section) (HelpFunc, UsageFunc) {
	helpFunc := func(c *cobra.Command, args []string) {
		t := template.New("top")
		t.Funcs(helpers.TemplateFuncs)
		template.Must(t.Parse(c.HelpTemplate()))

		// this is where we have to find the help sections we should show for this specific command
		data := map[string]interface{}{}
		data["Command"] = c

		err := t.Execute(c.OutOrStderr(), data)
		if err != nil {
			c.PrintErrln(err)
		}
	}

	usageFunc := func(c *cobra.Command) error {
		var tags []string

		tags = append(tags, fmt.Sprintf("command:%s", c.Name()))
		c.Flags().VisitAll(func(f *pflag.Flag) {
			tags = append(tags, fmt.Sprintf("flag:%s", f.Name))
		})

		// TODO(manuel, 2022-12-03) The commands matching a flag tag should probably be listed separately under the flag itself?
		// or maybe we should actually add some metadata about why the section was found
		// so that it can be used to render the help text
		cmdSections := FindSectionWithTags(sections, tags)

		t := template.New("top")

		renderedSections, err2 := renderHelpSections(cmdSections)
		if err2 != nil {
			return err2
		}

		// this is where we would have to find the help sections we should show for this specific command
		t.Funcs(helpers.TemplateFuncs)
		template.Must(t.Parse(c.UsageTemplate()))

		data := map[string]interface{}{}
		data["Command"] = c

		// TODO (manuel, 2021-12-03) - potentially we should also separate additional help sections, additional flag related usage sections, etc
		// TODO (manuel, 2021-12-03) - disable for actual usages for now, only shown on help
		data["Sections"] = renderedSections
		data["HasSections"] = len(renderedSections) > 0
		slugPadding := 7
		for _, section := range renderedSections {
			slug := section["Slug"].(string)
			if len(slug) > slugPadding {
				slugPadding = len(slug)
			}
		}
		for _, section := range renderedSections {
			section["SlugPadding"] = slugPadding + 1
		}

		err := t.Execute(c.OutOrStderr(), data)
		return err
	}

	return helpFunc, usageFunc
}

func renderHelpSections(sections []*Section) ([]map[string]interface{}, error) {
	var renderedSections []map[string]interface{}
	for _, section := range sections {
		buf := &strings.Builder{}
		err := section.Render(buf, NewRenderContext([]string{}, nil))
		if err != nil {
			return nil, err
		}
		renderedSections = append(renderedSections, map[string]interface{}{
			"Slug":    section.Slug,
			"Title":   section.Title,
			"Content": buf.String(),
			"Tags":    section.Tags,
		})
	}
	return renderedSections, nil
}

func GetHelpUsageTemplates(sections []*Section) (string, string) {
	_ = sections
	return HELP_TEMPLATE, USAGE_TEMPLATE
}

// NewCobraHelpCommand uses the InitDefaultHelpCommand code from cobra.
// This code is lifted from cobra and modified to accommodate help sections
//
// Copyright 2013-2022 The Cobra Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
// 2022-12-03 - Manuel Odendahl - Added support for help sections
func NewCobraHelpCommand(sections []*Section) *cobra.Command {
	var ret *cobra.Command
	ret = &cobra.Command{
		Use:   "help [topic/command]",
		Short: "Help about any command or topic",
		Long:  `Help provides help for any command and topic in the application.`,
		ValidArgsFunction: func(c *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			// copied from cobra itself
			var completions []string

			for _, section := range sections {
				completions = append(completions, fmt.Sprintf("%s\t%s", section.Slug, section.Title))
			}

			cmd, _, e := c.Root().Find(args)
			if e != nil {
				// couldn't find a command by that name
				return nil, cobra.ShellCompDirectiveNoFileComp
			}
			if cmd == nil {
				// Root help command.
				cmd = c.Root()
			}
			for _, subCmd := range cmd.Commands() {
				if subCmd.IsAvailableCommand() || subCmd == ret {
					if strings.HasPrefix(subCmd.Name(), toComplete) {
						completions = append(completions, fmt.Sprintf("%s\t%s", subCmd.Name(), subCmd.Short))
					}
				}
			}
			return completions, cobra.ShellCompDirectiveNoFileComp
		},

		Run: func(c *cobra.Command, args []string) {
			// we need to integrate those into the standard help command template
			section, _ := FindSection(sections, args)
			if section != nil {
				// we need to parse the tags here
				rc := NewRenderContext([]string{}, nil)
				err := section.Render(c.OutOrStdout(), rc)
				cobra.CheckErr(err)
			} else {
				root := c.Root()
				cmd, _, e := root.Find(args)
				if cmd == root {
					// we got asked to just `help`, so we need to output all the additional topics as part of the help command too,
					// not just the root command usage, or maybe we should just move the root command help to `help glaze`
					cobra.CheckErr(cmd.Help())
				} else if cmd == nil || e != nil {
					c.Printf("Unknown help topic %#q\n", args)
					cobra.CheckErr(root.Usage())
				} else {
					cmd.InitDefaultHelpFlag()    // make possible 'help' flag to be shown
					cmd.InitDefaultVersionFlag() // make possible 'version' flag to be shown
					cobra.CheckErr(cmd.Help())
				}
			}
		},
	}

	return ret
}

func NewRenderContext(tags []string, data interface{}) *RenderContext {
	return &RenderContext{
		Depth: 0,
		Tags:  tags,
		Data:  data,
	}
}

// USAGE_TEMPLATE - template used by the glazed library help cobra command.
// This template has been adapted from the cobra usage command template.
//
// Original: https://github.com/spf13/cobra
//
// Copyright 2013-2022 The Cobra Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
//
// 2022-12-03 - Manuel Odendahl - Augmented template with sections
const USAGE_TEMPLATE string = `{{with .Command}}Usage:{{if .Runnable}}
  {{.UseLine}}{{end}}{{if .HasAvailableSubCommands}}
  {{.CommandPath}} [command]{{end}}{{if gt (len .Aliases) 0}}

Aliases:
  {{.NameAndAliases}}{{end}}{{if .HasExample}}

Examples:
{{.Example}}{{end}}{{if .HasAvailableSubCommands}}{{$cmds := .Commands}}{{if eq (len .Groups) 0}}

Available Commands:{{range $cmds}}{{if (or .IsAvailableCommand (eq .Name "help"))}}
  {{rpad .Name .NamePadding }} {{.Short}}{{end}}{{end}}{{else}}{{range $group := .Groups}}

{{.Title}}{{range $cmds}}{{if (and (eq .GroupID $group.ID) (or .IsAvailableCommand (eq .Name "help")))}}
  {{rpad .Name .NamePadding }} {{.Short}}{{end}}{{end}}{{end}}{{if not .AllChildCommandsHaveGroup}}

Additional Commands:{{range $cmds}}{{if (and (eq .GroupID "") (or .IsAvailableCommand (eq .Name "help")))}}
  {{rpad .Name .NamePadding }} {{.Short}}{{end}}{{end}}{{end}}{{end}}{{end}}{{end}}{{if .HasSections}}

Additional Topics:{{$sections := .Sections}}{{range $sections}}
  {{rpad .Slug .SlugPadding }} {{.Title}}{{end}}{{end}}{{with .Command}}{{if .HasAvailableLocalFlags}}

Flags:
{{.LocalFlags.FlagUsages | trimTrailingWhitespaces}}{{end}}{{if .HasAvailableInheritedFlags}}

Global Flags:
{{.InheritedFlags.FlagUsages | trimTrailingWhitespaces}}{{end}}{{if .HasHelpSubCommands}}

Additional help topics:{{range .Commands}}{{if .IsAdditionalHelpTopicCommand}}
  {{rpad .CommandPath .CommandPathPadding}} {{.Short}}{{end}}{{end}}{{end}}{{if .HasAvailableSubCommands}}

Use "{{.CommandPath}} [command] --help" for more information about a command.{{end}}{{end}}
`

const HELP_TEMPLATE = `{{with .Command}}{{with (or .Long .Short)}}{{. | trimTrailingWhitespaces}}

{{end}}{{if or .Runnable .HasSubCommands}}{{.UsageString}}{{end}}{{end}}`
