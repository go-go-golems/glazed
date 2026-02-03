package patternmapper

// Exported test helpers and utilities for validation and parsing.

// ValidatePatternSyntax exposes internal validatePatternSyntax for testing and tooling.
func ValidatePatternSyntax(pattern string) error { return validatePatternSyntax(pattern) }

// ExtractCaptureNames exposes internal extractCaptureNames for testing and tooling.
func ExtractCaptureNames(pattern string) []string { return extractCaptureNames(pattern) }

// ExtractCaptureReferences exposes internal extractCaptureReferences for testing and tooling.
func ExtractCaptureReferences(targetField string) map[string]bool {
	return extractCaptureReferences(targetField)
}

// ResolveTargetField exposes internal resolveTargetField for testing and tooling.
func ResolveTargetField(targetField string, captures map[string]string) (string, error) {
	return resolveTargetField(targetField, captures)
}
