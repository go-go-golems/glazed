package schema

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/go-go-golems/glazed/pkg/cmds/fields"
	"github.com/spf13/cobra"
	flag "github.com/spf13/pflag"
)

// FlagGroup is a group of flags that can be added to a cobra command.
// While we mostly deal with Definitions, this uses strings
// because it can be applied to any cobra flag in general.
//
// It limits us in the sense that we can't just get the full FieldDefinition
// here, but at least we can format our help a little bit more nicely.
//
// NOTE(manuel, 2023-02-20) This doesn't allow for hierarchical flag groups yet.
// Let's see how this feels overall, and if this is something we want to add later on.
// This is useful I think because subsystems such as glaze already pull in so many flags,
// and it could be used in conjunction with renaming the actual flags used on the CLI
// as more colisions are prone to happen.
type FlagGroup struct {
	ID     string
	Name   string
	Prefix string
	Flags  []string
	Order  int
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
	Slug          string
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
// template. Fields that are not assigned to any group are passed as the "" group, with the
// name "Other flags".
type CommandFlagGroupUsage struct {
	LocalGroupUsages     []*FlagGroupUsage
	InheritedGroupUsages []*FlagGroupUsage
}

func (c *CommandFlagGroupUsage) String() string {
	return fmt.Sprintf("CommandFlagGroupUsage{LocalGroupUsages: %v, InheritedGroupUsages: %v}",
		len(c.LocalGroupUsages), len(c.InheritedGroupUsages))
}

const GlobalDefaultSlug = "global-default"

// ComputeCommandFlagGroupUsage is used to compute the flag groups to be shown in the
// Usage help function.
//
// It is a fairly complex function that gathers all LocalFlags() and InheritedFlags()
// from the cobra backend. It then iterated over the FlagGroups that have been added
// through sections usually.
func ComputeCommandFlagGroupUsage(c *cobra.Command) *CommandFlagGroupUsage {
	ret := &CommandFlagGroupUsage{}

	// compute the grouped flags, instead of relying on c.LocalFlags() and c.InheritedFlags()
	localFlags := c.LocalFlags()
	inheritedFlags := c.InheritedFlags()

	flagGroups := GetFlagGroups(c)

	localGroupedFlags := map[string]*FlagGroupUsage{}
	inheritedGroupedFlags := map[string]*FlagGroupUsage{}

	flagToGroups := map[string][]string{}

	localGroupedFlags[DefaultSlug] = &FlagGroupUsage{
		Name:       "Flags",
		FlagUsages: []*FlagUsage{},
		Slug:       DefaultSlug,
	}
	inheritedGroupedFlags[GlobalDefaultSlug] = &FlagGroupUsage{
		Name:       "flags", // This will get displayed as "Global flags"
		FlagUsages: []*FlagUsage{},
		Slug:       GlobalDefaultSlug,
	}

	// Get an overview of which flag to assign to whom.
	//
	// Iterate over all the flag groups, and store which flag belongs to which group
	// in flagToGroups.
	for _, group := range flagGroups {
		localGroupedFlags[group.ID] = &FlagGroupUsage{
			Name:       group.Name,
			Slug:       group.ID,
			FlagUsages: []*FlagUsage{},
		}
		inheritedGroupedFlags[group.ID] = &FlagGroupUsage{
			Name:       group.Name,
			Slug:       group.ID,
			FlagUsages: []*FlagUsage{},
		}

		for _, flagName := range group.Flags {
			flagName = group.Prefix + flagName

			// check if flagToGroups already has the flagName
			if _, ok := flagToGroups[flagName]; !ok {
				flagToGroups[flagName] = []string{}
			}
			flagToGroups[flagName] = append(flagToGroups[flagName], group.ID)
		}
	}

	// Now visit all cobra flags, get their usage as defined when added to cobra
	// (usually through the FieldDefinition), and add them to the correct
	// group.
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
			localGroupedFlags[DefaultSlug].AddFlagUsage(flagUsage)
		}
	})

	// Do the same for inherited flags.
	inheritedFlags.VisitAll(func(f *flag.Flag) {
		flagUsage := getFlagUsage(f)
		if flagUsage == nil {
			return
		}

		// We move the help commands to the local group so that they always get displayed
		if f.Name == "long-help" {
			localGroupedFlags[DefaultSlug].AddFlagUsage(flagUsage)
			return
		}

		if groups, ok := flagToGroups[f.Name]; ok {
			for _, group := range groups {
				inheritedGroupedFlags[group].AddFlagUsage(flagUsage)
			}
		} else {
			inheritedGroupedFlags[GlobalDefaultSlug].AddFlagUsage(flagUsage)
		}
	})

	ret.LocalGroupUsages = []*FlagGroupUsage{
		localGroupedFlags[DefaultSlug],
	}
	ret.InheritedGroupUsages = []*FlagGroupUsage{
		inheritedGroupedFlags[GlobalDefaultSlug],
	}

	// now add them in sorted order
	for _, group := range flagGroups {
		// Skip the default slug since it is always added, since it also contains the general purpose flags
		if group.ID == DefaultSlug || group.ID == GlobalDefaultSlug {
			continue
		}
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

	// NOTE(manuel, 2023-02-20) This is where we should compute the necessary alignment indent
	// for each group

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
	default:
		switch defValue {
		case "0", "false", "", "[]", "map[]", "<nil>":
			return true
		default:
			return false
		}
	}
}

// getFlagUsage returns the FlagUsage for a given flag.
// It tries to do its best to transform the data that cobra provides about the
// flag into a pleasantly human-readable string.
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

// AddFlagGroupToCobraCommand adds a flag group to a cobra command.
// This is done by adding a set of annotations to the command:
//   - glazed:flag-group-order: a comma-separated list of flag group IDs in the
//     order they should be displayed
//   - glazed:flag-group-count: the number of flag groups
//   - glazed:flag-group:<id>:<name> - a list of the flag names in the group
//   - glazed:flag-group-prefix:<id> - the prefix to use for the group
func AddFlagGroupToCobraCommand(
	cmd *cobra.Command,
	id string,
	name string,
	flags *fields.Definitions,
	prefix string,
) {
	flagNames := []string{}
	for v := flags.Oldest(); v != nil; v = v.Next() {
		f := v.Value
		// replace _ with -
		name_ := strings.ReplaceAll(f.Name, "_", "-")
		flagNames = append(flagNames, name_)
	}

	if cmd.Annotations == nil {
		cmd.Annotations = map[string]string{
			"glazed:flag-group-order": "",
			"glazed:flag-group-count": "0",
		}
	}

	count_, ok := cmd.Annotations["glazed:flag-group-count"]
	if !ok {
		count_ = "0"
	}
	count, err := strconv.Atoi(count_)
	if err != nil {
		count = 0
	}

	order, ok := cmd.Annotations["glazed:flag-group-order"]
	if !ok {
		order = ""
	}

	cmd.Annotations[fmt.Sprintf("glazed:flag-group:%s:%s", id, name)] = strings.Join(flagNames, ",")
	if prefix != "" {
		cmd.Annotations[fmt.Sprintf("glazed:flag-group-prefix:%s", id)] = prefix
	}

	order = fmt.Sprintf("%s,%s", order, id)
	cmd.Annotations["glazed:flag-group-order"] = order

	count += 1
	cmd.Annotations["glazed:flag-group-count"] = strconv.Itoa(count)
}

func SetFlagGroupOrder(cmd *cobra.Command, order []string) {
	cmd.Annotations["glazed:flag-group-order"] = strings.Join(order, ",")
}

// GetFlagGroups returns a list of flag groups for the given command.
// It does so by gathering all parents flag groups and then checking
// for cobra Annotations of the form `glazed:flag-group:<id>:<name>`.
//
// The order of the groups is determined by the order of the ids
// in the `glazed:flag-group-order` annotation.
//
// Finally, the `glazed:flag-group-prefix:<id>:<prefix>` annotation
// is used to determine the prefix for the group.
func GetFlagGroups(cmd *cobra.Command) []*FlagGroup {
	groups := map[string]*FlagGroup{}

	if cmd.Parent() != nil {
		parentGroups := GetFlagGroups(cmd.Parent())
		for _, g := range parentGroups {
			groups[g.ID] = g
		}
	}

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

	// load the prefixes if present
	for id, group := range groups {
		prefixKey := fmt.Sprintf("glazed:flag-group-prefix:%s", id)
		if cmd.Annotations[prefixKey] != "" {
			group.Prefix = cmd.Annotations[prefixKey]
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
