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
//
// NOTE(manuel, 2023-02-20) This doesn't allow for hierarchical flag groups yet.
// Let's see how this feels overall, and if this is something we want to add later on.
// This is useful I think because subsystems such as glaze already pull in so many flags,
// and it could be used in conjunction with renaming the actual flags used on the CLI
// as more colisions are prone to happen.
type FlagGroup struct {
	ID    string
	Name  string
	Flags []string
	Order int
}

// FlagUsage is the structured information we want to show at help time.
// Instead of rendering the full time, we leave how these things are formatted
// all the way to the end, because for example aligning strings can only be done
// at runtime since we don't know which other flags might have been added to the
// one group.
type FlagUsage struct {
	ShortHand  string
	Long       string
	FlagString string
	Help       string
	Default    string
}

// FlagGroupUsage is used to render the help for a flag group.
// It consists of the group Name for rendering purposes, and a single string per
// flag in the group
type FlagGroupUsage struct {
	Name          string
	FlagUsages    []*FlagUsage
	MaxFlagLength int
}

func (f *FlagGroupUsage) String() string {
	return fmt.Sprintf("FlagGroupUsage{Name: %s, FlagUsages: %v}", f.Name, len(f.FlagUsages))
}

func (f *FlagGroupUsage) AddFlagUsage(flag *FlagUsage) {
	f.FlagUsages = append(f.FlagUsages, flag)
	if len(flag.FlagString) > f.MaxFlagLength {
		f.MaxFlagLength = len(flag.FlagString)
	}
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
		Name:       "Flags",
		FlagUsages: []*FlagUsage{},
	}
	inheritedGroupedFlags[""] = &FlagGroupUsage{
		Name:       "flags",
		FlagUsages: []*FlagUsage{},
	}

	// get an overview of which flag to assign to whom
	for _, group := range flagGroups {
		localGroupedFlags[group.ID] = &FlagGroupUsage{
			Name:       group.Name,
			FlagUsages: []*FlagUsage{},
		}
		inheritedGroupedFlags[group.ID] = &FlagGroupUsage{
			Name:       group.Name,
			FlagUsages: []*FlagUsage{},
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
		flagUsage := getFlagUsage(f)
		if flagUsage == nil {
			return
		}

		if groups, ok := flagToGroups[f.Name]; ok {
			for _, group := range groups {
				localGroupedFlags[group].AddFlagUsage(flagUsage)
			}
		} else {
			localGroupedFlags[""].AddFlagUsage(flagUsage)
		}
	})

	inheritedFlags.VisitAll(func(f *flag.Flag) {
		flagUsage := getFlagUsage(f)
		if flagUsage == nil {
			return
		}

		if groups, ok := flagToGroups[f.Name]; ok {
			for _, group := range groups {
				inheritedGroupedFlags[group].AddFlagUsage(flagUsage)
			}
		} else {
			inheritedGroupedFlags[""].AddFlagUsage(flagUsage)
		}
	})

	ret.LocalGroupUsages = []*FlagGroupUsage{
		localGroupedFlags[""],
	}
	ret.InheritedGroupUsages = []*FlagGroupUsage{
		inheritedGroupedFlags[""],
	}

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

	// NOTE(manuel, 2023-02-20) This is where we should compute the necessary alignment indent for each group

	return ret
}

func isZeroValue(v flag.Value, defValue string) bool {
	vType := v.Type()

	switch vType {
	case "string":
		return defValue == ""
	case "bool":
		return defValue == "false"
	case "int":
		return defValue == "0"
	case "stringSlice", "intSlice", "stringArray", "intArray":
	default:
		switch defValue {
		case "0", "false", "", "[]", "map[]", "<nil>":
			return true
		default:
			return false
		}
	}

	return false
}
func getFlagUsage(f *flag.Flag) *FlagUsage {
	if f.Hidden {
		return nil
	}

	ret := &FlagUsage{
		Long: f.Name,
	}

	if f.Shorthand != "" && f.ShorthandDeprecated == "" {
		ret.ShortHand = f.Shorthand
		ret.FlagString = fmt.Sprintf("-%s, --%s", f.Shorthand, f.Name)
	} else {
		ret.FlagString = fmt.Sprintf("--%s", f.Name)
	}

	_, usage := flag.UnquoteUsage(f)
	ret.Help = usage

	if f.NoOptDefVal != "" {
		switch f.Value.Type() {
		case "string":
			ret.Default = fmt.Sprintf("[=\"%s\"]", f.NoOptDefVal)
		case "bool":
			if f.NoOptDefVal != "true" {
				ret.Default = fmt.Sprintf("[=%s]", f.NoOptDefVal)
			}
		case "count":
			if f.NoOptDefVal != "+1" {
				ret.Default = fmt.Sprintf("[=%s]", f.NoOptDefVal)
			}
		default:
			ret.Default = fmt.Sprintf("[=%s]", f.NoOptDefVal)
		}
	}

	if !isZeroValue(f.Value, f.DefValue) {
		if f.Value.Type() == "string" {
			ret.Default += fmt.Sprintf(" (default %q)", f.DefValue)
		} else {
			ret.Default += fmt.Sprintf(" (default %s)", f.DefValue)
		}

	}
	if len(f.Deprecated) != 0 {
		ret.Help += fmt.Sprintf(" (DEPRECATED: %s)", f.Deprecated)
	}

	return ret
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
		id := parts[2]
		name := parts[3]

		flags := strings.Split(v, ",")
		groups[id] = &FlagGroup{
			ID:    id,
			Name:  name,
			Flags: flags,
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

	if cmd.Parent() != nil {
		parentGroups := GetFlagGroups(cmd.Parent())
		returnGroups = append(returnGroups, parentGroups...)
	}

	// sort by order
	sort.Slice(returnGroups, func(i, j int) bool {
		return returnGroups[i].Order < returnGroups[j].Order
	})

	return returnGroups
}