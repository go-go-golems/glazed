package model

import "github.com/pkg/errors"

type SectionType int

const (
	SectionGeneralTopic SectionType = iota
	SectionExample
	SectionApplication
	SectionTutorial
)

func SectionTypeFromString(s string) (SectionType, error) {
	switch s {
	case "GeneralTopic":
		return SectionGeneralTopic, nil
	case "Example":
		return SectionExample, nil
	case "Application":
		return SectionApplication, nil
	case "Tutorial":
		return SectionTutorial, nil
	}
	return SectionGeneralTopic, errors.Errorf("unknown section type %s", s)
}

func (s SectionType) String() string {
	switch s {
	case SectionGeneralTopic:
		return "GeneralTopic"
	case SectionExample:
		return "Example"
	case SectionApplication:
		return "Application"
	case SectionTutorial:
		return "Tutorial"
	}
	return "Unknown"
}

// SectionTypeFromStringUnsafe returns the SectionType for a string, defaulting to SectionGeneralTopic if unknown.
func SectionTypeFromStringUnsafe(s string) SectionType {
	switch s {
	case "GeneralTopic":
		return SectionGeneralTopic
	case "Example":
		return SectionExample
	case "Application":
		return SectionApplication
	case "Tutorial":
		return SectionTutorial
	}
	return SectionGeneralTopic
}
