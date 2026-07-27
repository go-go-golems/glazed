package settings

type StructuredOutputSection struct{}

func NewStructuredOutputSection() (*StructuredOutputSection, error) {
	return &StructuredOutputSection{}, nil
}
