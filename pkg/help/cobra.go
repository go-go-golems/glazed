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
		qb := NewQueryBuilder().
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
		qb := NewQueryBuilder().
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

	// this is where we would have to find the help sections we should show for this specific command
	t.Funcs(helpers.TemplateFuncs)
	tmpl := COBRA_COMMAND_HELP_TEMPLATE + c.UsageTemplate()
	if options.ShowShortTopic {
		tmpl = COBRA_COMMAND_SHORT_HELP_TEMPLATE + c.UsageTemplate()
	}
	if options.ShowAllSections {
		tmpl += HELP_LONG_SECTION_TEMPLATE
	} else {
		tmpl += HELP_SHORT_SECTION_TEMPLATE
	}
	template.Must(t.Parse(tmpl))

	data := map[string]interface{}{}
	data["Command"] = c
	data["HelpCommand"] = options.HelpCommand
	data["Slug"] = c.Name()

	isTopLevel := c.Parent() == nil
	if isTopLevel {
		hp := NewHelpPage(options.Query.OnlyTopLevel().FindSections(hs.Sections))
		data["Help"] = hp
	} else {
		hp := NewHelpPage(options.Query.OnlyCommands(c.Name()).FindSections(hs.Sections))
		data["Help"] = hp
	}

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

			generalTopics := NewQueryBuilder().
				OnlyTopLevel().
				ReturnTopics().
				FindSections(hs.Sections)

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
			root := c.Root()

			// TODO(manuel, 2022-12-09): What we want to do here is warn if a certain
			// request (for a flag, a command, etc...) wasn't found, and to give a list
			// of the possible alternatives instead. Say, if looking for the flag --templates,
			// we can provide a list of the possible flags.

			// TODO(manuel, 2022-12-09): Furthermore, when searching for a flag, we should
			// make it easier on the user by stripping -- in case the option was passed with
			// dashes. It's not clear that we expect the flag name without --
			explicitInformationRequested := false
			someFlagSet := false
			qb := NewQueryBuilder()

			topic := c.Flag("topic").Value.String()
			if topic != "" {
				qb = qb.OnlyTopics(topic)
				explicitInformationRequested = true
				someFlagSet = true
			}
			flag := c.Flag("flag").Value.String()
			if flag != "" {
				qb = qb.OnlyFlags(flag)
				explicitInformationRequested = true
				someFlagSet = true
			}

			command := c.Flag("command").Value.String()
			if command != "" {
				qb = qb.OnlyCommands(command)
				explicitInformationRequested = true
				someFlagSet = true
			}

			showAllSections, _ := c.Flags().GetBool("all")
			showShortTopic, _ := c.Flags().GetBool("short")

			topics, _ := c.Flags().GetBool("topics")
			if topics {
				qb = qb.ReturnTopics()
				showAllSections = true
				showShortTopic = true
				someFlagSet = true
			}
			examples, _ := c.Flags().GetBool("examples")
			if examples {
				qb = qb.ReturnExamples()
				showAllSections = true
				showShortTopic = true
				someFlagSet = true
			}
			applications, _ := c.Flags().GetBool("applications")
			if applications {
				qb = qb.ReturnApplications()
				showAllSections = true
				showShortTopic = true
				someFlagSet = true
			}
			tutorials, _ := c.Flags().GetBool("tutorials")
			if tutorials {
				qb = qb.ReturnTutorials()
				showAllSections = true
				showShortTopic = true
				someFlagSet = true
			}

			if !topics && !examples && !applications && !tutorials {
				qb = qb.ReturnAllTypes()
			}

			list, _ := c.Flags().GetBool("list")
			if list {
				someFlagSet = true
			}

			options := &RenderOptions{
				Query:                       qb,
				ShowAllSections:             showAllSections,
				ShowShortTopic:              showShortTopic,
				ExplictInformationRequested: explicitInformationRequested,
				SomeFlagSet:                 someFlagSet,
				ListSections:                list,
				HelpCommand:                 root.CommandPath() + " help",
			}

			// first, we check if we can find an explicit help topic
			if len(args) >= 1 {
				topicSection, err := hs.GetSectionWithSlug(args[0])

				// if we found a topic with that slug, show it
				if err == nil {
					// we allow the user to restrict the search to subtopics
					options.Query = options.Query.
						OnlyTopics(args...).
						WithoutSections(topicSection)

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
				if someFlagSet {

				} else if list {
					// TODO(manuel, 2022-12-09): We could show a main help page if specified

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
	// - toc
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
