package cmds

type CommandDescription struct{}

type CommandDescriptionOption func(*CommandDescription)

func NewCommandDescription(name string, options ...CommandDescriptionOption) *CommandDescription {
	return &CommandDescription{}
}

func WithSections(sections ...interface{}) CommandDescriptionOption {
	return func(*CommandDescription) {}
}

func WithShort(s string) CommandDescriptionOption {
	return func(*CommandDescription) {}
}
