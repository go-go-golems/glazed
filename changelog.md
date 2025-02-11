# Parameter Layer Serialization

Added ability to serialize parameter layers to YAML/JSON format for better interoperability and configuration management.

- Added SerializableParameterLayer struct for YAML/JSON serialization
- Added SerializableLayers struct for serializing collections of layers as a map keyed by slug
- Added conversion functions ToSerializable and LayersToSerializable
- Implemented YAML and JSON marshaling for ParameterLayers
- Updated serialization to maintain layer order while providing slug-based access
- Added custom YAML and JSON marshalers for SerializableLayers

# Parsed Parameters Serialization

Added ability to serialize parsed parameters to YAML/JSON format for better debugging and state persistence.

- Added SerializableParsedParameter struct for YAML/JSON serialization
- Added SerializableParsedParameters struct for serializing collections of parsed parameters
- Added conversion functions for ParsedParameter and ParsedParameters
- Implemented YAML and JSON marshaling for ParsedParameters
- Maintained parameter order while providing name-based access in serialized format

# Parsed Layer Serialization

Added ability to serialize parsed layers to YAML/JSON format, combining layer definitions and parsed parameters.

- Added SerializableParsedLayer struct for YAML/JSON serialization
- Added SerializableParsedLayers struct for serializing collections of parsed layers
- Added conversion functions for ParsedLayer and ParsedLayers
- Implemented YAML and JSON marshaling for ParsedLayer and ParsedLayers
- Included both layer definitions and parsed parameters in serialized output

## Documentation Clarification for Help System Implementation

Clarified the documentation about implementing AddDocToHelpSystem, explaining the recommended approach of creating a doc package with embedded documentation files.

- Updated help entry to show how to properly implement AddDocToHelpSystem in user's own package
- Added example of doc.go implementation with embed functionality

# Optional GlazedCommandLayer in CobraParser

Added ability to skip adding the GlazedCommandLayer when creating a new CobraParser.

- Added skipGlazedCommandLayer flag to CobraParser struct
- Added WithSkipGlazedCommandLayer option function
- Modified NewCobraParserFromLayers to respect the skip flag 