package types

// SectionType represents the type of a help section.
type SectionType string

const (
	SectionGeneralTopic SectionType = "GeneralTopic"
	SectionExample      SectionType = "Example"
	SectionApplication  SectionType = "Application"
	SectionTutorial     SectionType = "Tutorial"
)

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