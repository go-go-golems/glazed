package cmds

import (
	"fmt"
	"github.com/spf13/cobra"
	flag "github.com/spf13/pflag"
	"sort"
	"strings"
)

// FlagGroup is a group of flags that can be added to a cobra command.
// While we mostly deal with ParameterDefinitions, this uses strings
// because it can be applied to any cobra flag in general.
//
// It limits us in the sense that we can't just get the full ParameterDefinition
// here, but at least we can format our help a little bit more nicely.
type FlagGroup struct {
	ID    string
	Name  string
	Flags []string
	Order int
}

// FlagGroupUsage is used to render the help for a flag group.
// It consists of the group Name for rendering purposes, and a single string per
// flag in the group
type FlagGroupUsage struct {
	Name       string
	FlagUsages []string
}

func (f *FlagGroupUsage) String() string {
	return fmt.Sprintf("FlagGroupUsage{Name: %s, FlagUsages: %v}", f.Name, len(f.FlagUsages))
}

// CommandFlagGroupUsage is used to render the flags for an entire command.
// This gets parsed at rendering time, and passed along the command to the usage or help
// template. Flags that are not assigned to any group are passed as the "" group, with the
// name "Other flags".
type CommandFlagGroupUsage struct {
	LocalGroupUsages     []*FlagGroupUsage
	InheritedGroupUsages []*FlagGroupUsage
}

func (c *CommandFlagGroupUsage) String() string {
	return fmt.Sprintf("CommandFlagGroupUsage{LocalGroupUsages: %v, InheritedGroupUsages: %v}",
		len(c.LocalGroupUsages), len(c.InheritedGroupUsages))
}

func ComputeCommandFlagGroupUsage(c *cobra.Command) *CommandFlagGroupUsage {
	ret := &CommandFlagGroupUsage{}

	// compute the grouped flags, instead of relying on c.LocalFlags() and c.InheritedFlags()
	localFlags := c.LocalFlags()
	inheritedFlags := c.InheritedFlags()

	flagGroups := GetFlagGroups(c)

	localGroupedFlags := map[string]*FlagGroupUsage{}
	inheritedGroupedFlags := map[string]*FlagGroupUsage{}

	flagToGroups := map[string][]string{}

	localGroupedFlags[""] = &FlagGroupUsage{
		Name:       "Other flags",
		FlagUsages: []string{},
	}
	inheritedGroupedFlags[""] = &FlagGroupUsage{
		Name:       "Other flags",
		FlagUsages: []string{},
	}

	// get an overview of which flag to assign to whom
	for _, group := range flagGroups {
		localGroupedFlags[group.ID] = &FlagGroupUsage{
			Name:       group.Name,
			FlagUsages: []string{},
		}
		inheritedGroupedFlags[group.ID] = &FlagGroupUsage{
			Name:       group.Name,
			FlagUsages: []string{},
		}

		for _, flagName := range group.Flags {
			// check if flagToGroups already has the flagName
			if _, ok := flagToGroups[flagName]; !ok {
				flagToGroups[flagName] = []string{}
			}
			flagToGroups[flagName] = append(flagToGroups[flagName], group.ID)
		}
	}

	localFlags.VisitAll(func(f *flag.Flag) {
		usageString := getFlagUsageString(f)

		if groups, ok := flagToGroups[f.Name]; ok {
			for _, group := range groups {
				localGroupedFlags[group].FlagUsages = append(localGroupedFlags[group].FlagUsages, usageString)
			}
		} else {
			localGroupedFlags[""].FlagUsages = append(localGroupedFlags[""].FlagUsages, usageString)
		}
	})

	inheritedFlags.VisitAll(func(f *flag.Flag) {
		usageString := getFlagUsageString(f)

		if groups, ok := flagToGroups[f.Name]; ok {
			for _, group := range groups {
				inheritedGroupedFlags[group].FlagUsages = append(inheritedGroupedFlags[group].FlagUsages, usageString)
			}
		} else {
			inheritedGroupedFlags[""].FlagUsages = append(inheritedGroupedFlags[""].FlagUsages, usageString)
		}
	})

	// now add them in sorted order
	for _, group := range flagGroups {
		if _, ok := localGroupedFlags[group.ID]; ok {
			if len(localGroupedFlags[group.ID].FlagUsages) > 0 {
				ret.LocalGroupUsages = append(ret.LocalGroupUsages, localGroupedFlags[group.ID])
			}
		}
		if _, ok := inheritedGroupedFlags[group.ID]; ok {
			if len(inheritedGroupedFlags[group.ID].FlagUsages) > 0 {
				ret.InheritedGroupUsages = append(ret.InheritedGroupUsages, inheritedGroupedFlags[group.ID])
			}
		}
	}

	ret.LocalGroupUsages = append(ret.LocalGroupUsages, localGroupedFlags[""])
	ret.InheritedGroupUsages = append(ret.InheritedGroupUsages, inheritedGroupedFlags[""])

	return ret
}

func getFlagUsageString(f *flag.Flag) string {
	if f.Hidden {
		return ""
	}

	line := ""
	if f.Shorthand != "" && f.ShorthandDeprecated == "" {
		line = fmt.Sprintf(".  **-%s**, **--%s**", f.Shorthand, f.Name)
	} else {
		line = fmt.Sprintf(".      **--%s**", f.Name)
	}

	varname, usage := flag.UnquoteUsage(f)
	if varname != "" {
		line += " " + varname
	}
	if f.NoOptDefVal != "" {
		switch f.Value.Type() {
		case "string":
			line += fmt.Sprintf("[=\"%s\"]", f.NoOptDefVal)
		case "bool":
			if f.NoOptDefVal != "true" {
				line += fmt.Sprintf("[=%s]", f.NoOptDefVal)
			}
		case "count":
			if f.NoOptDefVal != "+1" {
				line += fmt.Sprintf("[=%s]", f.NoOptDefVal)
			}
		default:
			line += fmt.Sprintf("[=%s]", f.NoOptDefVal)
		}
	}

	line += " "

	line += usage
	if f.Value.Type() == "string" {
		line += fmt.Sprintf(" (default %q)", f.DefValue)
	} else {
		line += fmt.Sprintf(" (default %s)", f.DefValue)
	}
	if len(f.Deprecated) != 0 {
		line += fmt.Sprintf(" (DEPRECATED: %s)", f.Deprecated)
	}

	return line
}

func AddFlagGroupToCobraCommand(cmd *cobra.Command, id string, name string, flags []*ParameterDefinition) {
	flagNames := []string{}
	for _, flag := range flags {
		flagNames = append(flagNames, flag.Name)
	}

	if cmd.Annotations == nil {
		cmd.Annotations = map[string]string{}
	}
	cmd.Annotations[fmt.Sprintf("glazed:flag-group:%s:%s", id, name)] = strings.Join(flagNames, ",")
}

func SetFlagGroupOrder(cmd *cobra.Command, order []string) {
	cmd.Annotations["glazed:flag-group-order"] = strings.Join(order, ",")
}

func GetFlagGroups(cmd *cobra.Command) []*FlagGroup {
	groups := map[string]*FlagGroup{}

	for k, v := range cmd.Annotations {
		if !strings.HasPrefix(k, "glazed:flag-group:") {
			continue
		}

		parts := strings.Split(k, ":")
		groups[parts[1]] = &FlagGroup{
			ID:    parts[2],
			Name:  parts[3],
			Flags: strings.Split(v, ","),
		}
	}

	// check for the presence of glazed:flag-group-order
	if cmd.Annotations["glazed:flag-group-order"] != "" {
		order := strings.Split(cmd.Annotations["glazed:flag-group-order"], ",")
		for i, id := range order {
			if groups[id] != nil {
				groups[id].Order = i
			}
		}
	}

	// now convert to a slice
	returnGroups := []*FlagGroup{}
	for _, group := range groups {
		returnGroups = append(returnGroups, group)
	}

	// sort by order
	sort.Slice(returnGroups, func(i, j int) bool {
		return returnGroups[i].Order < returnGroups[j].Order
	})

	return returnGroups
}
