---
Title: Flag Groups
Slug: flag-groups
Short: |
  Flag groups are a way to group flags together in the help system. This is useful
  if you have a lot of flags that are related to each other.
Topics:
- help
IsTemplate: false
IsTopLevel: true
ShowPerDefault: true
SectionType: GeneralTopic
---

By default, the cobra help system does not provide the functionality to group
flags in the help output.

Because glazed provides many different flags for controlling all kinds of functionality,
the usage help for most commands using glazed become overwhelmed with glazed flags,
which are secondary to the commands actual flags.

In order to simplify the usage page of glazed commands, glazed provides a way to
group flags together. This is done by using using the functions `cmds.AddFlagGroupToCobraCommand`
and `cmds.SetFlagGroupOrder`.

These two functions store the necessary data inside custom `cobra.Command` annotations:
- `glazed:flag-group:$ID:$NAME` which contains the comma separated list of flags in that group
- `glazed:flag-group-order` which contains the comma separated list of flag group
  IDs in the order they should be displayed

These annotations are parsed by the glazed HelpSystem UsageFunc and are used to group both
local and global flags into groups at display time.