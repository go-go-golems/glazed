package patternmapper

// Exported test helpers and utilities for validation and parsing.

// ValidatePatternSyntax exposes internal validatePatternSyntax for testing and tooling.
func ValidatePatternSyntax(pattern string) error { return validatePatternSyntax(pattern) }

// ExtractCaptureNames exposes internal extractCaptureNames for testing and tooling.
func ExtractCaptureNames(pattern string) []string { return extractCaptureNames(pattern) }

// ExtractCaptureReferences exposes internal extractCaptureReferences for testing and tooling.
func ExtractCaptureReferences(targetParameter string) map[string]bool {
	return extractCaptureReferences(targetParameter)
}

// ResolveTargetParameter exposes internal resolveTargetParameter for testing and tooling.
func ResolveTargetParameter(targetParameter string, captures map[string]string) (string, error) {
	return resolveTargetParameter(targetParameter, captures)
}
