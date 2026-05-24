package pflag

type FlagSet struct{}

func String(name string, value string, usage string) *string       { return &value }
func StringVar(p *string, name string, value string, usage string) {}
func Bool(name string, value bool, usage string) *bool             { return &value }
func BoolVar(p *bool, name string, value bool, usage string)       {}
func NewFlagSet(name string, errorHandling int) *FlagSet           { return &FlagSet{} }

func (f *FlagSet) String(name string, value string, usage string) *string       { return &value }
func (f *FlagSet) StringVar(p *string, name string, value string, usage string) {}
func (f *FlagSet) Bool(name string, value bool, usage string) *bool             { return &value }
func (f *FlagSet) BoolVar(p *bool, name string, value bool, usage string)       {}
