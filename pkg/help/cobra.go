package help

import (
	_ "embed"
	"fmt"
	"github.com/charmbracelet/glamour"
	tsize "github.com/kopoli/go-terminal-size"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"glazed/pkg/helpers"
	"strings"
	"text/template"
)

type HelpFunc = func(c *cobra.Command, args []string)
type UsageFunc = func(c *cobra.Command) error

func GetHelpUsageFuncs(hs *HelpSystem) (HelpFunc, UsageFunc) {
	calledFromHelp := false
	helpFunc := func(c *cobra.Command, args []string) {
		t := template.New("top")
		t.Funcs(helpers.TemplateFuncs)
		helpTemplate := c.HelpTemplate()
		template.Must(t.Parse(helpTemplate))

		// this is where we have to find the help sections we should show for this specific command
		data := map[string]interface{}{}
		data["Command"] = c

		prevCalledFromHelp := calledFromHelp
		calledFromHelp = true
		out, err := renderToMarkdown(t, data)

		calledFromHelp = prevCalledFromHelp

		_, _ = fmt.Fprintln(c.OutOrStderr(), out)
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

		t := template.New("commandUsage")

		// this is where we would have to find the help sections we should show for this specific command
		t.Funcs(helpers.TemplateFuncs)
		templateString := c.UsageTemplate() + HELP_SHORT_SECTION_TEMPLATE
		template.Must(t.Parse(templateString))

		data := map[string]interface{}{}
		data["Command"] = c
		data["HelpCommand"] = c.CommandPath() + " help"
		data["Slug"] = c.Name()

		isTopLevel := c.Parent() == nil
		if isTopLevel {
			data["Help"] = hs.GetTopLevelHelpPage()
		} else {
			data["Help"] = hs.GetCommandHelpPage(c.Name())
		}

		var err error

		// TODO this is a hack to get the error handling to show the help as markdown
		// Not sure if we should bypass this entirely, and do all the markdown rendering in Usage itself
		if calledFromHelp {
			err = t.Execute(c.OutOrStderr(), data)
		} else {
			// if we are not called from help, then maybe we should render the markdown
			// here?
			s, err := renderToMarkdown(t, data)
			if err != nil {
				return err
			}
			_, _ = fmt.Fprintln(c.OutOrStderr(), s)
		}

		return err
	}

	return helpFunc, usageFunc
}

func renderToMarkdown(t *template.Template, data map[string]interface{}) (string, error) {
	sz, err := tsize.GetSize()
	if err != nil {
		sz.Width = 80
	}

	// get markdown output
	var sb strings.Builder
	r, _ := glamour.NewTermRenderer(
		glamour.WithWordWrap(sz.Width),
		//glamour.WithAutoStyle(),
		// TODO(manuel, 2022-12-04): We need to check if we can use colors here,
		// which is not the case if we render things out to a file / pipe,
		// in the context of a redirect, or if we render to file
		glamour.WithStandardStyle("dark"),
	)

	err = t.Execute(&sb, data)

	s := sb.String()
	sizeString := fmt.Sprintf("size: %dx%d\n", sz.Width, sz.Height)
	_ = sizeString

	out, err := r.Render(s)
	return out, err
}

func GetHelpUsageTemplates(hs *HelpSystem) (string, string) {
	_ = hs
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
// 2022-12-04 - Manuel Odendahl - Significantly reworked to support markdown sections
func NewCobraHelpCommand(hs *HelpSystem) *cobra.Command {
	var ret *cobra.Command
	ret = &cobra.Command{
		Use:   "help [topic/command]",
		Short: "Help about any command or topic",
		Long:  `Help provides help for any command and topic in the application.`,
		ValidArgsFunction: func(c *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			// copied from cobra itself
			var completions []string

			generalTopics := GetSectionsByType(GetTopLevelSections(hs.Sections), SectionGeneralTopic)
			for _, section := range generalTopics {
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
			// TODO(manuel, 2022-12-04): Handle properly showing all the options we have
			// and which sections to render out how. The cobra-usage.tmpl has been a crutch until
			// now
			if len(args) == 1 {
				// we need to integrate those into the standard help command template
				topicSections := GetSectionsBySlug(hs.Sections, args[0])
				if len(topicSections) > 1 {
					// if we have multiple topics we should show the short section (kind of table of contents for the whole thing)
					fmt.Println("XXX we should show a toplevel topic index page")
					_ = hs.RenderSectionSummaries(c.OutOrStdout(), topicSections)
				} else if len(topicSections) == 1 {
					// TODO(manuel, 2022-12-04): Markdown rendering of help topics is not working yet
					_ = hs.RenderTopic(c.OutOrStdout(), topicSections[0])
					return
				}
			}

			root := c.Root()
			cmd, _, e := root.Find(args)
			if cmd == root {
				// we got asked to just `help`, so we need to output all the additional topics as part of the help command too,
				// not just the root command usage, or maybe we should just move the root command help to `help glaze`
				// TODO - we need to add the short section list but only for toplevel topics
				cobra.CheckErr(cmd.Help())
			} else if cmd == nil || e != nil {
				c.Printf("Unknown help topic %#q\n", args)
				cobra.CheckErr(root.Usage())
			} else {
				cmd.InitDefaultHelpFlag()    // make possible 'help' flag to be shown
				cmd.InitDefaultVersionFlag() // make possible 'version' flag to be shown
				cobra.CheckErr(cmd.Help())
			}
		},
	}

	ret.Flags().StringSlice("topics", []string{}, "Show help for topics")
	ret.Flags().StringSlice("commands", []string{}, "Show help for commands")
	ret.Flags().StringSlice("flags", []string{}, "Show help for flags")
	ret.Flags().StringSlice("applications", []string{}, "Show help for applications")
	ret.Flags().StringSlice("tutorials", []string{}, "Show help for tutorials")
	ret.Flags().Bool("examples", false, "Show examples")

	// TODO(manuel, 2022-12-04): Additional verbs to build
	// - toc
	// - topics
	// - search
	// - serve

	return ret
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
// 2022-12-04 - Manuel Odendahl - Significantly reworked to support markdown sections
//go:embed templates/cobra-usage.tmpl
var USAGE_TEMPLATE string

//go:embed templates/help-short-section-list.tmpl
var HELP_SHORT_SECTION_TEMPLATE string

const HELP_TEMPLATE = `{{with .Command}}{{with (or .Long .Short)}}# {{. | trimTrailingWhitespaces}}

{{end}}{{if or .Runnable .HasSubCommands}}{{.UsageString}}{{end}}{{end}}`
