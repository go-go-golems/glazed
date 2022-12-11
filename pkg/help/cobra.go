package help

import (
	_ "embed"
	"fmt"
	"github.com/spf13/cobra"
	"github.com/wesen/glazed/pkg/helpers"
	"strings"
	"text/template"
)

type HelpFunc = func(c *cobra.Command, args []string)
type UsageFunc = func(c *cobra.Command) error

func GetCobraHelpUsageFuncs(hs *HelpSystem) (HelpFunc, UsageFunc) {
	helpFunc := func(c *cobra.Command, args []string) {
		qb := NewSectionQuery().
			ReturnAllTypes()

		options := &RenderOptions{
			Query:           qb,
			ShowAllSections: false,
			ShowShortTopic:  false,
			HelpCommand:     c.Root().CommandPath() + " help",
		}

		cobra.CheckErr(renderCommandHelpPage(c, options, hs))
	}

	usageFunc := func(c *cobra.Command) error {
		qb := NewSectionQuery().
			ReturnExamples()

		options := &RenderOptions{
			Query:           qb,
			ShowAllSections: false,
			ShowShortTopic:  true,
			HelpCommand:     c.Root().CommandPath() + " help",
		}
		return renderCommandHelpPage(c, options, hs)
	}

	return helpFunc, usageFunc
}

func renderCommandHelpPage(c *cobra.Command, options *RenderOptions, hs *HelpSystem) error {
	t := template.New("commandUsage")

	isTopLevel := c.Parent() == nil

	userQuery := options.Query

	if isTopLevel {
		// we need to clone the userQuery because we need to modify it to restrict the search to the command
		// but however the initial data is computed from the incoming userQuery
		userQuery = userQuery.ReturnOnlyTopLevel()
	} else {
		userQuery = userQuery.SearchForCommand(c.Name())
	}

	data, noResultsFound := hs.ComputeRenderData(userQuery)

	t.Funcs(helpers.TemplateFuncs)
	var tmpl string
	if options.ListSections || noResultsFound {
		tmpl = COBRA_COMMAND_SHORT_HELP_TEMPLATE + HELP_LIST_TEMPLATE
	} else {
		if options.ShowShortTopic {
			tmpl = COBRA_COMMAND_SHORT_HELP_TEMPLATE
		} else {
			tmpl = COBRA_COMMAND_HELP_TEMPLATE
		}
		if !userQuery.HasOnlyQueries() && !userQuery.HasRestrictedReturnTypes() {
			tmpl += c.UsageTemplate()
		}
		if options.ShowAllSections {
			tmpl += HELP_LONG_SECTION_TEMPLATE
		} else {
			tmpl += HELP_SHORT_SECTION_TEMPLATE
		}
	}
	template.Must(t.Parse(tmpl))

	data["Command"] = c
	data["HelpCommand"] = options.HelpCommand
	data["Slug"] = c.Name()

	s, err := RenderToMarkdown(t, data)
	if err != nil {
		return err
	}
	_, _ = fmt.Fprintln(c.OutOrStderr(), s)

	return err
}

func GetCobraHelpUsageTemplates(hs *HelpSystem) (string, string) {
	_ = hs
	return COBRA_COMMAND_HELP_TEMPLATE, COBRA_COMMAND_USAGE_TEMPLATE
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

			for _, section := range hs.Sections {
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
			root := c.Root()

			qb := NewSectionQuery()

			topic := c.Flag("topic").Value.String()
			if topic != "" {
				qb = qb.ReturnOnlyTopics(topic)
			}
			flag := c.Flag("flag").Value.String()
			if flag != "" {
				qb = qb.ReturnOnlyFlags(flag)
			}

			command := c.Flag("command").Value.String()
			if command != "" {
				qb = qb.ReturnOnlyCommands(command)
			}

			showAllSections, _ := c.Flags().GetBool("all")
			showShortTopic, _ := c.Flags().GetBool("short")

			topics, _ := c.Flags().GetBool("topics")
			if topics {
				qb = qb.ReturnTopics()
				showAllSections = true
				showShortTopic = true
			}
			examples, _ := c.Flags().GetBool("examples")
			if examples {
				qb = qb.ReturnExamples()
				showAllSections = true
				showShortTopic = true
			}
			applications, _ := c.Flags().GetBool("applications")
			if applications {
				qb = qb.ReturnApplications()
				showAllSections = true
				showShortTopic = true
			}
			tutorials, _ := c.Flags().GetBool("tutorials")
			if tutorials {
				qb = qb.ReturnTutorials()
				showAllSections = true
				showShortTopic = true
			}

			if !topics && !examples && !applications && !tutorials {
				qb = qb.ReturnAllTypes()
			}

			list, _ := c.Flags().GetBool("list")

			options := &RenderOptions{
				Query:           qb,
				ShowAllSections: showAllSections,
				ShowShortTopic:  showShortTopic,
				ListSections:    list,
				HelpCommand:     root.CommandPath() + " help",
			}

			// first, we check if we can find an explicit help topic
			if len(args) >= 1 {
				topicSection, err := hs.GetSectionWithSlug(args[0])

				// if we found a topic with that slug, show it
				if err == nil {
					// TODO(manuel, 2022-12-10) Potentially allow subtopics search
					options.Query = options.Query.
						ReturnAnyOfTopics(args[0]).
						FilterSections(topicSection)

					s, err := hs.RenderTopicHelp(
						topicSection,
						options)
					if err != nil {
						// need to show the default error page here
						c.Printf("Unknown help topic: %s", args[0])
						cobra.CheckErr(root.Usage())
					}
					_, _ = fmt.Fprintln(c.OutOrStderr(), s)
					return
				}
			}

			// if we couldn't find an explicit help page, show command help
			cmd, _, e := root.Find(args)
			if cmd == nil || e != nil {
				c.Printf("Unknown help topic %#q\n", args)
				if list {
					// TODO(manuel, 2022-12-09): We could show a main help page if specified
					cobra.CheckErr(renderCommandHelpPage(root, options, hs))
				} else {
					cobra.CheckErr(renderCommandHelpPage(root, options, hs))
				}
			} else {
				cobra.CheckErr(renderCommandHelpPage(cmd, options, hs))

			}
		},
	}

	ret.Flags().String("topic", "", "Show help related to topic")
	ret.Flags().String("command", "", "Show help related to command")
	ret.Flags().String("flag", "", "Show help related to flag")

	ret.Flags().Bool("list", false, "List all sections")
	ret.Flags().Bool("topics", false, "Show all topics")
	ret.Flags().Bool("examples", false, "Show all examples")
	ret.Flags().Bool("applications", false, "Show all applications")
	ret.Flags().Bool("tutorials", false, "Show all tutorials")

	ret.Flags().Bool("all", false, "Show all sections, not just default")
	ret.Flags().Bool("short", false, "Show short version")

	// TODO(manuel, 2022-12-04): Additional verbs to build
	// - toc -- done with --list
	// - topics
	// - search
	// - serve

	return ret
}

// COBRA_COMMAND_USAGE_TEMPLATE - template used by the glazed library help cobra command.
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
var COBRA_COMMAND_USAGE_TEMPLATE string

//go:embed templates/cobra-help.tmpl
var COBRA_COMMAND_HELP_TEMPLATE string

//go:embed templates/cobra-short-help.tmpl
var COBRA_COMMAND_SHORT_HELP_TEMPLATE string
