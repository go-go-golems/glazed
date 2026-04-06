package fields

// RedactedPlaceholder is used when a sensitive value is omitted from rendered output.
const RedactedPlaceholder = "***"

func redactStringValue(s string) string {
	if len(s) <= 6 {
		return RedactedPlaceholder
	}

	return s[:2] + RedactedPlaceholder + s[len(s)-2:]
}

// RedactValue returns a printable representation of a value for sensitive field types.
// Non-sensitive field types are returned unchanged.
func RedactValue(type_ Type, value interface{}) interface{} {
	if !type_.IsSensitive() {
		return value
	}

	switch v := value.(type) {
	case nil:
		return nil
	case string:
		return redactStringValue(v)
	case []string:
		ret := make([]string, len(v))
		for i, s := range v {
			ret[i] = redactStringValue(s)
		}
		return ret
	case []interface{}:
		ret := make([]interface{}, len(v))
		for i, item := range v {
			ret[i] = RedactValue(type_, item)
		}
		return ret
	case map[string]string:
		ret := make(map[string]string, len(v))
		for key, item := range v {
			ret[key] = redactStringValue(item)
		}
		return ret
	case map[string]interface{}:
		ret := make(map[string]interface{}, len(v))
		for key, item := range v {
			ret[key] = RedactValue(type_, item)
		}
		return ret
	default:
		return value
	}
}

// RedactMetadata returns a redacted copy of the metadata for sensitive field types.
// Non-sensitive field types are returned unchanged.
func RedactMetadata(type_ Type, metadata map[string]interface{}) map[string]interface{} {
	if !type_.IsSensitive() || metadata == nil {
		return metadata
	}

	ret := make(map[string]interface{}, len(metadata))
	for key, value := range metadata {
		ret[key] = RedactValue(type_, value)
	}

	return ret
}

// RedactParseStep returns a copy of the parse step with any sensitive data masked.
func RedactParseStep(type_ Type, step ParseStep) ParseStep {
	if !type_.IsSensitive() {
		return step
	}

	ret := step
	ret.Value = RedactValue(type_, step.Value)
	ret.Metadata = RedactMetadata(type_, step.Metadata)

	return ret
}
