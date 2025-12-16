// Package appconfig provides a small, incremental facade for parsing configuration
// into typed application settings.
//
// The initial implementation is intentionally not “struct-first” in the sense of
// deriving layers from structs. Instead, callers explicitly register Glazed
// ParameterLayers and bind them to fields inside a grouped settings struct T.
//
// See CONFIG-PARSER-001 for the design and implementation diary.
package appconfig
