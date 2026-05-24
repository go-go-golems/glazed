package cobra

import "github.com/spf13/pflag"

type Command struct{}

func (c *Command) Flags() *pflag.FlagSet           { return &pflag.FlagSet{} }
func (c *Command) PersistentFlags() *pflag.FlagSet { return &pflag.FlagSet{} }
