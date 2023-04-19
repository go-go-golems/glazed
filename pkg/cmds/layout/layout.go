package layout

type Layout struct {
	Sections []*Section
}

type Section struct {
	Title       string `yaml:"title"`
	Description string `yaml:"description"`
	Rows        []*Row `yaml:"rows"`
	Style       string `yaml:"style,omitempty"`
	Classes     string `yaml:"classes,omitempty"`
}

type Row struct {
	Inputs  []*Input `yaml:"inputs"`
	Classes string   `yaml:"classes,omitempty"`
	Style   string   `yaml:"style,omitempty"`
}

type Input struct {
	Name string `yaml:"name"`

	// These can be specified to override the values read from the flag / layer parameter definition
	Label        string      `yaml:"label,omitempty"`
	Options      []Option    `yaml:"options,omitempty"`
	DefaultValue interface{} `yaml:"default,omitempty"`
	Help         string      `yaml:"help,omitempty"`

	// TODO(manuel, 2024-04-19) I'm leaving validation out for now, and even the dynamic condition stuff is for now not implemented
	// These are suggestion I got from chatgpt that I find quite interesting, but might not be the way I want to design them.
	//
	// See https://github.com/go-go-golems/parka/issues/29
	Validation map[string]interface{} `yaml:"validation,omitempty"`
	Condition  *Condition             `yaml:"condition,omitempty"`

	// this can be used to customizes the HTML output
	// see https://github.com/go-go-golems/parka/issues/28
	CSS       string `yaml:"css,omitempty"`
	Id        string `yaml:"id,omitempty"`
	Classes   string `yaml:"classes,omitempty"`
	Template  string `yaml:"template,omitempty"`
	InputType string `yaml:"type,omitempty"`

	// TODO(manuel, 2024-04-19) Further interesting ideas from ChatGPT
	// See https://github.com/go-go-golems/parka/issues/29
	//
	// - accessibility
	// - saving / restoring form status
	// - collaboration
	// - autocomplete
	// - internationalization
}

type Condition struct {
	InputName string      `yaml:"name"`
	Operator  string      `yaml:"operator"`
	Value     interface{} `yaml:"value"`
}

type Option struct {
	Label string      `yaml:"label"`
	Value interface{} `yaml:"value"`
}
