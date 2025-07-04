package help

import (
	_ "embed"
	"fmt"
	"github.com/Masterminds/sprig"
	glazed_cobra "github.com/go-go-golems/glazed/pkg/cmds/layers"
	"github.com/go-go-golems/glazed/pkg/helpers/templating"
	"github.com/spf13/cobra"
	"os"
	"strings"
	"text/template"
)

type HelpFunc = func(c *cobra.Command, args []string)
type UsageFunc = func(c *cobra.Command) error
type UIFunc = func(hs *HelpSystem) error

func GetCobraHelpUsageFuncs(hs *HelpSystem) (HelpFunc, UsageFunc) {
	helpFunc := func(c *cobra.Command, args []string) {
		qb := NewSectionQuery().
			ReturnAllTypes()

		longHelp, _ := c.Flags().GetBool("long-help")

		options := &RenderOptions{
			Query:           qb,
			ShowAllSections: false,
			ShowShortTopic:  false,
			LongHelp:        longHelp,
			HelpCommand:     c.Root().CommandPath() + " help",
		}

		c.NamePadding()
		if c.Parent() == nil {
			options.OnlyTopLevel = true
		}
		cobra.CheckErr(renderCommandHelpPage(c, options, hs))
	}

	usageFunc := func(c *cobra.Command) error {
		qb := NewSectionQuery().
			ReturnExamples()

		longHelp, _ := c.Flags().GetBool("long-help")

		options := &RenderOptions{
			Query:           qb,
			ShowAllSections: false,
			ShowShortTopic:  true,
			LongHelp:        longHelp,
			HelpCommand:     c.Root().CommandPath() + " help",
		}
		if c.Parent() == nil {
			options.OnlyTopLevel = true
		}
		return renderCommandHelpPage(c, options, hs)
	}

	return helpFunc, usageFunc
}

func renderCommandHelpPage(c *cobra.Command, options *RenderOptions, hs *HelpSystem) error {
	t := template.New("commandUsage")

	isTopLevel := c.Parent() == nil

	userQuery := options.Query
	userQuery.OnlyTopLevel = options.OnlyTopLevel

	if !isTopLevel {
		userQuery = userQuery.SearchForCommand(c.Name())
	}

	data, noResultsFound := hs.ComputeRenderData(userQuery)

	t.Funcs(templating.TemplateFuncs).Funcs(sprig.TxtFuncMap())

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
		if options.LongHelp {
			if options.ShowAllSections {
				tmpl += HELP_LONG_SECTION_TEMPLATE
			} else {
				tmpl += HELP_SHORT_SECTION_TEMPLATE
			}
		}
	}
	template.Must(t.Parse(tmpl))

	flagGroupUsage := glazed_cobra.ComputeCommandFlagGroupUsage(c)

	// if we are showing the short help and shortHelpLayers annotation was set,
	// skip all the groups that are not in the list
	if !options.LongHelp {
		shortHelpLayers_, ok := c.Annotations["shortHelpLayers"]
		if ok {
			shortHelpLayers := map[string]interface{}{}
			for _, v := range strings.Split(shortHelpLayers_, ",") {
				shortHelpLayers[v] = true
			}

			localGroupUsages := []*glazed_cobra.FlagGroupUsage{}
			inheritedGroupUsages := []*glazed_cobra.FlagGroupUsage{}
			for _, f := range flagGroupUsage.LocalGroupUsages {
				if _, ok = shortHelpLayers[f.Slug]; ok {
					localGroupUsages = append(localGroupUsages, f)
				}
			}
			for _, f := range flagGroupUsage.InheritedGroupUsages {
				if _, ok = shortHelpLayers[f.Slug]; ok {
					inheritedGroupUsages = append(inheritedGroupUsages, f)
				}
			}

			flagGroupUsage.LocalGroupUsages = localGroupUsages
			flagGroupUsage.InheritedGroupUsages = inheritedGroupUsages
		}

	}

	// really this is where we need to compute the max length, not on a group basis
	maxLength := 0
	for _, group := range flagGroupUsage.LocalGroupUsages {
		for _, usage := range group.FlagUsages {
			if len(usage.FlagString) > maxLength {
				maxLength = len(usage.FlagString)
			}
		}
	}

	data["Command"] = c
	data["FlagGroupUsage"] = flagGroupUsage
	data["FlagUsageMaxLength"] = maxLength
	data["HelpCommand"] = options.HelpCommand
	data["Slug"] = c.Name()
	data["LongHelp"] = options.LongHelp

	maxCommandNameLen := 0
	for _, c := range c.Commands() {
		if len(c.Name()) > maxCommandNameLen {
			maxCommandNameLen = len(c.Name())
		}
	}

	data["MaxCommandNameLen"] = maxCommandNameLen

	s, err := RenderToMarkdown(t, data, os.Stderr)
	if err != nil {
		return err
	}
	_, _ = fmt.Fprint(c.OutOrStderr(), s)

	return err
}

func GetCobraHelpUsageTemplates(hs *HelpSystem) (string, string) {
	_ = hs
	return COBRA_COMMAND_HELP_TEMPLATE, COBRA_COMMAND_USAGE_TEMPLATE
}

// NewCobraHelpCommand uses the InitDefaultHelpCommand code from cobra.
// This code is lifted from cobra and modified to accommodate help sections
//
// # Copyright 2013-2022 The Cobra Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
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
	return NewCobraHelpCommandWithUI(hs, nil)
}

func NewCobraHelpCommandWithUI(hs *HelpSystem, uiFunc UIFunc) *cobra.Command {
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

			// Check for UI flag first
			useUI, _ := c.Flags().GetBool("ui")
			if useUI {
				if uiFunc != nil {
					err := uiFunc(hs)
					if err != nil {
						c.Printf("Error running UI: %v\n", err)
					}
				} else {
					c.Printf("UI mode not available\n")
				}
				return
			}

			qb := NewSectionQuery()

			// Check for DSL query first
			queryDSL := c.Flag("query").Value.String()
			printQuery, _ := c.Flags().GetBool("print-query")
			printSQL, _ := c.Flags().GetBool("print-sql")

			// Check if debug flags are used without a query
			if (printQuery || printSQL) && queryDSL == "" {
				c.Printf("Error: --print-query and --print-sql can only be used with --query\n")
				return
			}

			if queryDSL != "" {
				// Handle debug printing
				if printQuery || printSQL {
					err := hs.PrintQueryDebug(queryDSL, printQuery, printSQL)
					if err != nil {
						c.Printf("Debug error: %s\n", err)
						return
					}
					if printQuery || printSQL {
						c.Printf("\n")
					}
				}

				// Handle DSL query
				sections, err := hs.QuerySections(queryDSL)
				if err != nil {
					c.Printf("Query error: %s\n", err)
					// Show syntax help for DSL queries
					c.Printf("\nQuery syntax:\n")
					c.Printf("  examples AND topic:templates  - Boolean AND operation\n")
					c.Printf("  type:example OR type:tutorial  - Boolean OR operation\n")
					c.Printf("  NOT type:application          - Boolean NOT operation\n")
					c.Printf("  (examples OR tutorials) AND topic:database - Parentheses\n")
					c.Printf("  \"search text\"                - Text search\n")
					c.Printf("  type:example                  - Field queries\n")
					c.Printf("\nValid fields: type, topic, flag, command, slug\n")
					c.Printf("Valid types: example, tutorial, topic, application\n")
					c.Printf("Valid shortcuts: examples, tutorials, topics, applications, toplevel, defaults\n")
					return
				}

				// Display results
				if len(sections) == 0 {
					c.Printf("No results found for query: %s\n", queryDSL)
					c.Printf("Try:\n")
					c.Printf("  examples           - Show all examples\n")
					c.Printf("  tutorials          - Show all tutorials\n")
					c.Printf("  topics             - Show all topics\n")
					c.Printf("  applications       - Show all applications\n")
					c.Printf("  type:example       - Show sections of type 'example'\n")
					c.Printf("  \"search text\"      - Search for text in content\n")
					return
				}

				c.Printf("Found %d section(s) for query: %s\n\n", len(sections), queryDSL)
				for _, section := range sections {
					c.Printf("• %s - %s\n", section.Slug, section.Title)
				}
				return
			}

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
				qb.OnlyTopLevel = false
				showAllSections = true
				showShortTopic = true
			}
			examples, _ := c.Flags().GetBool("examples")
			if examples {
				qb = qb.ReturnExamples()
				qb.OnlyTopLevel = false
				showAllSections = true
				showShortTopic = true
			}
			applications, _ := c.Flags().GetBool("applications")
			if applications {
				qb = qb.ReturnApplications()
				qb.OnlyTopLevel = false
				showAllSections = true
				showShortTopic = true
			}
			tutorials, _ := c.Flags().GetBool("tutorials")
			if tutorials {
				qb = qb.ReturnTutorials()
				qb.OnlyTopLevel = false
				showAllSections = true
				showShortTopic = true
			}

			if !topics && !examples && !applications && !tutorials {
				qb = qb.ReturnAllTypes()
			}

			list, _ := c.Flags().GetBool("list")

			if showAllSections || topics || examples || applications || tutorials {
				list = true
			}

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
					_, _ = fmt.Fprint(c.OutOrStderr(), s)
					return
				}
			} else {
				// TODO(manuel, 2022-12-09): We could show a main help page if specified
				cobra.CheckErr(renderCommandHelpPage(root, options, hs))
				return
			}

			// if we couldn't find an explicit help page, show command help
			// NOTE(manuel, 2024-04-04) This code is never reached, is it?
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
	ret.Flags().String("query", "", "Use query DSL to search help sections")

	ret.Flags().Bool("list", false, "List all sections")
	ret.Flags().Bool("topics", false, "Show all topics")
	ret.Flags().Bool("examples", false, "Show all examples")
	ret.Flags().Bool("applications", false, "Show all applications")
	ret.Flags().Bool("tutorials", false, "Show all tutorials")

	ret.Flags().Bool("all", false, "Show all sections, not just default")
	ret.Flags().Bool("short", false, "Show short version")

	// Interactive UI
	ret.Flags().Bool("ui", false, "Open interactive help UI")

	// Debug flags
	ret.Flags().Bool("print-query", false, "Print parsed query AST for debugging")
	ret.Flags().Bool("print-sql", false, "Print generated SQL query for debugging")

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
// # Copyright 2013-2022 The Cobra Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
// 2022-12-03 - Manuel Odendahl - Augmented template with sections
// 2022-12-04 - Manuel Odendahl - Significantly reworked to support markdown sections
//
//go:embed templates/cobra-usage.tmpl
var COBRA_COMMAND_USAGE_TEMPLATE string

//go:embed templates/cobra-help.tmpl
var COBRA_COMMAND_HELP_TEMPLATE string

//go:embed templates/cobra-short-help.tmpl
var COBRA_COMMAND_SHORT_HELP_TEMPLATE string
