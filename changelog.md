# Field Section Serialization

Added ability to serialize field sections to YAML/JSON format for better interoperability and configuration management.

- Added SerializableSection struct for YAML/JSON serialization
- Added SerializableSections struct for serializing collections of sections as a map keyed by slug
- Added conversion functions ToSerializable and SectionsToSerializable
- Implemented YAML and JSON marshaling for Schema
- Updated serialization to maintain section order while providing slug-based access
- Added custom YAML and JSON marshalers for SerializableSections

# Parsed Fields Serialization

Added ability to serialize parsed fields to YAML/JSON format for better debugging and state persistence.

- Added SerializableFieldValue struct for YAML/JSON serialization
- Added SerializableFieldValues struct for serializing collections of parsed fields
- Added conversion functions for FieldValue and FieldValues
- Implemented YAML and JSON marshaling for FieldValues
- Maintained field order while providing name-based access in serialized format

# Parsed Section Serialization

Added ability to serialize parsed sections to YAML/JSON format, combining section definitions and parsed fields.

- Added SerializableSectionValues struct for YAML/JSON serialization
- Added SerializableValues struct for serializing collections of parsed sections
- Added conversion functions for SectionValues and Values
- Implemented YAML and JSON marshaling for SectionValues and Values
- Included both section definitions and parsed fields in serialized output

## Documentation Clarification for Help System Implementation

Clarified the documentation about implementing AddDocToHelpSystem, explaining the recommended approach of creating a doc package with embedded documentation files.

- Updated help entry to show how to properly implement AddDocToHelpSystem in user's own package
- Added example of doc.go implementation with embed functionality

# Optional GlazedCommandSection in CobraParser

Added ability to skip adding the GlazedCommandSection when creating a new CobraParser.

- Added skipGlazedCommandSection flag to CobraParser struct
- Added WithSkipGlazedCommandSection option function
- Modified NewCobraParserFromSections to respect the skip flag

# Optional Profile and Create Command Settings Sections in CobraParser

Added ability to enable ProfileSettingsSection and CreateCommandSettingsSection when creating a new CobraParser. These sections are disabled by default and must be explicitly enabled.

- Added enableProfileSettingsSection flag to CobraParser struct
- Added enableCreateCommandSettingsSection flag to CobraParser struct
- Added WithProfileSettingsSection option function to enable profile settings
- Added WithCreateCommandSettingsSection option function to enable create command settings
- Modified NewCobraParserFromSections to only add these sections when explicitly enabled 