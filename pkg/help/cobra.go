package help

import (
	"context"
	_ "embed"
	"fmt"
	"github.com/Masterminds/sprig"
	glazed_cobra "github.com/go-go-golems/glazed/pkg/cmds/layers"
	"github.com/go-go-golems/glazed/pkg/help/model"
	"github.com/go-go-golems/glazed/pkg/help/query"
	"github.com/go-go-golems/glazed/pkg/helpers/templating"
	"github.com/spf13/cobra"
	"os"
	"strings"
	"text/template"
)

type HelpFunc = func(c *cobra.Command, args []string)
type UsageFunc = func(c *cobra.Command) error

func GetCobraHelpUsageFuncs(hs *HelpSystem) (HelpFunc, UsageFunc) {
	helpFunc := func(c *cobra.Command, args []string) {
		// Build a predicate for all types
		pred := query.Or(
			query.IsType(model.SectionGeneralTopic),
			query.IsType(model.SectionExample),
			query.IsType(model.SectionApplication),
			query.IsType(model.SectionTutorial),
		)

		longHelp, _ := c.Flags().GetBool("long-help")

		options := &RenderOptions{
			Predicate:       pred,
			Store:           hs.Store, // TODO: Ensure HelpSystem has a Store field
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
		// Build a predicate for examples only
		pred := query.IsType(model.SectionExample)

		longHelp, _ := c.Flags().GetBool("long-help")

		options := &RenderOptions{
			Predicate:       pred,
			Store:           hs.Store, // TODO: Ensure HelpSystem has a Store field
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

	// Use the predicate and store from options
	ctx := context.Background() // TODO: Use proper context if available
	data, noResultsFound, err := ComputeRenderData(ctx, options.Store, options.Predicate)
	if err != nil {
		return err
	}

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
		tmpl += c.UsageTemplate()
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

			// Build predicate based on flags
			var preds []query.Predicate
			if c.Flag("topics").Value.String() == "true" {
				preds = append(preds, query.IsType(model.SectionGeneralTopic))
			}
			if c.Flag("examples").Value.String() == "true" {
				preds = append(preds, query.IsType(model.SectionExample))
			}
			if c.Flag("applications").Value.String() == "true" {
				preds = append(preds, query.IsType(model.SectionApplication))
			}
			if c.Flag("tutorials").Value.String() == "true" {
				preds = append(preds, query.IsType(model.SectionTutorial))
			}
			if c.Flag("topic").Value.String() != "" {
				preds = append(preds, query.HasTopic(c.Flag("topic").Value.String()))
			}
			if c.Flag("flag").Value.String() != "" {
				preds = append(preds, query.HasFlag(c.Flag("flag").Value.String()))
			}
			if c.Flag("command").Value.String() != "" {
				preds = append(preds, query.HasCommand(c.Flag("command").Value.String()))
			}
			if len(preds) == 0 {
				preds = append(preds,
					query.Or(
						query.IsType(model.SectionGeneralTopic),
						query.IsType(model.SectionExample),
						query.IsType(model.SectionApplication),
						query.IsType(model.SectionTutorial),
					),
				)
			}
			pred := query.And(preds...)

			showAllSections, _ := c.Flags().GetBool("all")
			showShortTopic, _ := c.Flags().GetBool("short")
			list, _ := c.Flags().GetBool("list")

			options := &RenderOptions{
				Predicate:       pred,
				Store:           hs.Store, // TODO: Ensure HelpSystem has a Store field
				ShowAllSections: showAllSections,
				ShowShortTopic:  showShortTopic,
				ListSections:    list,
				HelpCommand:     root.CommandPath() + " help",
			}

			ctx := context.Background() // TODO: Use proper context if available

			if len(args) >= 1 {
				// TODO: Implement topicSection lookup using the new system if needed
				// For now, just render the help page
				s, err := RenderTopicHelp(ctx, nil, options)
				if err != nil {
					c.Printf("Unknown help topic: %s", args[0])
					cobra.CheckErr(root.Usage())
				}
				_, _ = fmt.Fprint(c.OutOrStderr(), s)
				return
			} else {
				cobra.CheckErr(renderCommandHelpPage(root, options, hs))
				return
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
