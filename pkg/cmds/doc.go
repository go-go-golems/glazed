package cmds

// The commands part of glazed contains multiple things:
//
// # Describe command line applications declaratively
//
// - structs to describe flags and arguments for commands
// - helper functions to register these structs as commands with cobra
//
// Because these applications are domain specific, they usually need to be overloaded
// and this is where the current messy situation is starting.
//
// Currently, sqleton, escuse-me and pinocchio benefit from loading applications
// declaratively. In addition to flags and arguments, they need to loader:
//
// - a SQL query template, in sqleton's case
// - a ElasticSearch query template, in escuse-me's case
// - a complex array of steps and factory settings, in pinocchio's case
//
// Currently, all 3 applications use a single YAML file to store commands.
//
// # Create aliases for command line applications
//
// - a generic CommandAlias struct that should be possible to use by
//   any kind of command line application.
//
// # Load commands and aliases from disk and register with cobra
//
// If an application implements YAMLCommandLoader, glazed provides helper
// functions to loader all commands and applications from a directory
// containing YAML files, by giving the application control over how
// these YAML files are parsed.
//
// This part probably will be refactored soon
// See https://github.com/go-go-golems/glazed/issues/117
