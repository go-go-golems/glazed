package middlewares

import (
	"regexp"
	"strings"

	"github.com/go-go-golems/glazed/pkg/cmds/layers"
	"github.com/pkg/errors"
)

// MappingRule defines a pattern-based mapping from config file structure to layer parameters.
// A mapping rule specifies:
//   - Source: Pattern to match in the config file (e.g., "app.settings.api_key", "app.{env}.api_key")
//   - TargetLayer: Which layer to place the value in
//   - TargetParameter: Which parameter name to use (supports capture references like "{env}-api-key")
//   - Required: Whether the pattern must match (default: false)
//   - Rules: Optional nested rules for mapping child objects
type MappingRule struct {
	// Source path pattern (e.g., "app.settings.api_key", "app.*.key", "app.{env}.api-key")
	// Supports:
	//   - Exact match: "app.settings.api_key"
	//   - Wildcard: "app.*.api_key" (matches but doesn't capture)
	//   - Named capture: "app.{env}.api_key" (captures "env")
	Source string

	// Target layer slug (e.g., "demo")
	// If not set in child rules, inherits from parent rule
	TargetLayer string

	// Target parameter name (supports captures like "{env}-api-key")
	// Capture references use the format "{name}" where name is a capture group from Source
	TargetParameter string

	// Optional: nested rules for mapping child objects
	// If provided, Source should point to an object, and Rules maps its children
	Rules []MappingRule

	// Optional: whether to skip if source doesn't exist (default: false, means skip silently)
	// If Required is true, pattern must match or an error is returned
	Required bool
}

// ConfigMapper is an interface that can map raw config data to layer maps.
// This allows both ConfigFileMapper (function) and pattern-based mappers to be used interchangeably.
type ConfigMapper interface {
	Map(rawConfig interface{}) (map[string]map[string]interface{}, error)
}

// configFileMapperAdapter adapts ConfigFileMapper function to ConfigMapper interface
type configFileMapperAdapter struct {
	fn ConfigFileMapper
}

func (a *configFileMapperAdapter) Map(rawConfig interface{}) (map[string]map[string]interface{}, error) {
	return a.fn(rawConfig)
}

// adaptConfigFileMapper converts a ConfigFileMapper function to ConfigMapper interface
func adaptConfigFileMapper(fn ConfigFileMapper) ConfigMapper {
	if fn == nil {
		return nil
	}
	return &configFileMapperAdapter{fn: fn}
}

// patternMapper implements ConfigMapper using pattern matching rules
type patternMapper struct {
	rules           []MappingRule
	layers          *layers.ParameterLayers
	compiledPatterns []compiledPattern
}

// compiledPattern represents a compiled pattern with its capture groups
type compiledPattern struct {
	rule        MappingRule
	pattern     *regexp.Regexp
	captureNames []string // ordered list of capture group names
}

// NewConfigMapper creates a new pattern-based config mapper from the given rules.
// The mapper validates that:
//   - All patterns are valid syntax
//   - All target parameters exist in their respective layers
//   - Capture references in target parameters match captures in source patterns
func NewConfigMapper(layers *layers.ParameterLayers, rules ...MappingRule) (ConfigMapper, error) {
	if layers == nil {
		return nil, errors.New("layers cannot be nil")
	}

	mapper := &patternMapper{
		rules:  rules,
		layers: layers,
	}

	// Compile and validate all patterns
	if err := mapper.compilePatterns(); err != nil {
		return nil, err
	}

	return mapper, nil
}

// compilePatterns compiles all patterns and validates them
func (m *patternMapper) compilePatterns() error {
	m.compiledPatterns = make([]compiledPattern, 0, len(m.rules))

	for i, rule := range m.rules {
		compiled, err := m.compileRule(rule, "", nil)
		if err != nil {
			return errors.Wrapf(err, "rule %d (source: %q)", i, rule.Source)
		}
		m.compiledPatterns = append(m.compiledPatterns, compiled...)
	}

	return nil
}

// compileRule compiles a single rule and its nested rules (if any)
func (m *patternMapper) compileRule(rule MappingRule, parentPath string, parentCaptures []string) ([]compiledPattern, error) {
	var compiled []compiledPattern

	// Validate pattern syntax
	if err := validatePatternSyntax(rule.Source); err != nil {
		return nil, errors.Wrapf(err, "invalid pattern syntax")
	}

	// Build full path if nested
	fullPath := rule.Source
	if parentPath != "" {
		fullPath = parentPath + "." + rule.Source
	}

	// Extract captures from this rule
	ruleCaptures := extractCaptureNames(rule.Source)
	allCaptures := append(parentCaptures, ruleCaptures...)

	// Validate target parameter if this is a leaf rule (no nested rules)
	if len(rule.Rules) == 0 {
		// Validate target layer exists
		if rule.TargetLayer == "" {
			return nil, errors.New("target layer is required for leaf rules")
		}

		_, ok := m.layers.Get(rule.TargetLayer)
		if !ok {
			return nil, errors.Errorf("target layer %q does not exist", rule.TargetLayer)
		}

		// Validate capture references in target parameter
		// Check against all captures (parent + current)
		if err := validateCaptureReferences(allCaptures, rule.TargetParameter); err != nil {
			return nil, errors.Wrapf(err, "invalid capture reference in target parameter")
		}

		// Compile regex pattern (for future optimization)
		pattern, captureNames, err := compilePatternToRegex(fullPath)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to compile pattern")
		}

		// Create a new rule with the full path for matching
		flattenedRule := rule
		flattenedRule.Source = fullPath

		compiled = append(compiled, compiledPattern{
			rule:         flattenedRule,
			pattern:      pattern,
			captureNames: captureNames,
		})
	} else {
		// Nested rules: compile each child rule
		for i, childRule := range rule.Rules {
			// Inherit target layer if not set
			if childRule.TargetLayer == "" {
				childRule.TargetLayer = rule.TargetLayer
			}

			childCompiled, err := m.compileRule(childRule, fullPath, allCaptures)
			if err != nil {
				return nil, errors.Wrapf(err, "nested rule %d", i)
			}
			compiled = append(compiled, childCompiled...)
		}
	}

	return compiled, nil
}

// Map implements ConfigMapper interface
func (m *patternMapper) Map(rawConfig interface{}) (map[string]map[string]interface{}, error) {
	result := make(map[string]map[string]interface{})

	// Convert config to map[string]interface{} if needed
	configMap, ok := rawConfig.(map[string]interface{})
	if !ok {
		return nil, errors.Errorf("expected map[string]interface{}, got %T", rawConfig)
	}

	// Match each pattern against the config
	for _, compiled := range m.compiledPatterns {
		matches, err := m.matchPattern(compiled, configMap, "")
		if err != nil {
			return nil, err
		}

		// Process each match
		for _, match := range matches {
			// Resolve target parameter name (replace captures)
			targetParam, err := resolveTargetParameter(compiled.rule.TargetParameter, match.captures)
			if err != nil {
				return nil, errors.Wrapf(err, "failed to resolve target parameter")
			}

			// Validate parameter exists
			layer, ok := m.layers.Get(match.layer)
			if !ok {
				return nil, errors.Errorf("target layer %q does not exist", match.layer)
			}

			// Check if parameter exists (accounting for prefix)
			paramName := targetParam
			if layer.GetPrefix() != "" {
				// If layer has prefix, check if targetParam already includes it
				if !strings.HasPrefix(targetParam, layer.GetPrefix()) {
					paramName = layer.GetPrefix() + targetParam
				}
			}

			paramDef, ok := layer.GetParameterDefinitions().Get(paramName)
			if !ok || paramDef == nil {
				// Parameter doesn't exist - this is a validation error
				return nil, errors.Errorf(
					"target parameter %q does not exist in layer %q (pattern: %q)",
					paramName,
					match.layer,
					compiled.rule.Source,
				)
			}

			// Initialize layer map if needed
			if result[match.layer] == nil {
				result[match.layer] = make(map[string]interface{})
			}

			// Set the value
			result[match.layer][paramName] = match.value
		}
	}

	return result, nil
}

// patternMatch represents a single pattern match result
type patternMatch struct {
	layer    string
	value    interface{}
	captures map[string]string
}

// matchPattern matches a compiled pattern against the config and returns all matches
func (m *patternMapper) matchPattern(
	compiled compiledPattern,
	config map[string]interface{},
	pathPrefix string,
) ([]patternMatch, error) {
	var matches []patternMatch

	// For nested rules, we need to first match the parent pattern
	// For now, we only handle one level of nesting, so we can simplify

	// Convert pattern to a path traversal
	err := m.traverseAndMatch(compiled, config, pathPrefix, make(map[string]string), &matches)
	if err != nil {
		return nil, err
	}

	// If no matches and required, return error
	if len(matches) == 0 && compiled.rule.Required {
		return nil, errors.Errorf(
			"required pattern %q did not match any paths in config",
			compiled.rule.Source,
		)
	}

	return matches, nil
}

// traverseAndMatch recursively traverses the config and matches patterns
func (m *patternMapper) traverseAndMatch(
	compiled compiledPattern,
	config map[string]interface{},
	currentPath string,
	parentCaptures map[string]string,
	matches *[]patternMatch,
) error {
	// Split pattern into segments
	segments := strings.Split(compiled.rule.Source, ".")
	return m.matchSegments(segments, config, currentPath, parentCaptures, compiled, matches)
}

// matchSegments matches pattern segments against the config
func (m *patternMapper) matchSegments(
	segments []string,
	config map[string]interface{},
	currentPath string,
	captures map[string]string,
	compiled compiledPattern,
	matches *[]patternMatch,
) error {
	if len(segments) == 0 {
		// All segments matched, this is a value
		if currentPath == "" {
			// This shouldn't happen for leaf values
			return nil
		}
		return nil
	}

	segment := segments[0]
	remaining := segments[1:]

	// Check if segment is a capture group {name}
	if strings.HasPrefix(segment, "{") && strings.HasSuffix(segment, "}") {
		name := segment[1 : len(segment)-1]
		// Match all keys at this level
		for key, value := range config {
			newCaptures := make(map[string]string)
			for k, v := range captures {
				newCaptures[k] = v
			}
			newCaptures[name] = key

			// Continue matching
			if err := m.matchSegmentsRecursive(remaining, value, currentPath+"."+key, newCaptures, compiled, matches); err != nil {
				return err
			}
		}
		return nil
	}

	// Check if segment is a wildcard *
	if segment == "*" {
		// Match all keys at this level
		for key, value := range config {
			if err := m.matchSegmentsRecursive(remaining, value, currentPath+"."+key, captures, compiled, matches); err != nil {
				return err
			}
		}
		return nil
	}

	// Exact match
	value, ok := config[segment]
	if !ok {
		return nil // No match, but not an error (might be optional)
	}

	newPath := segment
	if currentPath != "" {
		newPath = currentPath + "." + segment
	}

	return m.matchSegmentsRecursive(remaining, value, newPath, captures, compiled, matches)
}

// matchSegmentsRecursive handles the recursive matching logic
func (m *patternMapper) matchSegmentsRecursive(
	remaining []string,
	value interface{},
	currentPath string,
	captures map[string]string,
	compiled compiledPattern,
	matches *[]patternMatch,
) error {
	if len(remaining) == 0 {
		// No more segments - this is the value we're looking for
		*matches = append(*matches, patternMatch{
			layer:    compiled.rule.TargetLayer,
			value:    value,
			captures: captures,
		})
		return nil
	}

	// Continue matching remaining segments
	valueMap, ok := value.(map[string]interface{})
	if !ok {
		// Value is not a map, can't continue
		return nil
	}

	return m.matchSegments(remaining, valueMap, currentPath, captures, compiled, matches)
}

// validatePatternSyntax validates that a pattern string is syntactically valid
func validatePatternSyntax(pattern string) error {
	if pattern == "" {
		return errors.New("pattern cannot be empty")
	}

	segments := strings.Split(pattern, ".")
	for _, segment := range segments {
		if segment == "" {
			return errors.New("pattern cannot contain empty segments")
		}

		// Validate capture groups
		if strings.HasPrefix(segment, "{") {
			if !strings.HasSuffix(segment, "}") {
				return errors.Errorf("unclosed capture group in segment %q", segment)
			}
			name := segment[1 : len(segment)-1]
			if name == "" {
				return errors.Errorf("capture group name cannot be empty in segment %q", segment)
			}
			// Validate name is alphanumeric + underscore
			if !regexp.MustCompile(`^[a-zA-Z_][a-zA-Z0-9_]*$`).MatchString(name) {
				return errors.Errorf("invalid capture group name %q (must be alphanumeric or underscore)", name)
			}
		}
	}

	return nil
}

// validateCaptureReferences validates that all capture references in target parameter
// correspond to captures in the available captures list
func validateCaptureReferences(availableCaptures []string, targetParameter string) error {
	// Extract capture references from target parameter
	targetRefs := extractCaptureReferences(targetParameter)

	// Check all target references exist in available captures
	for ref := range targetRefs {
		found := false
		for _, cap := range availableCaptures {
			if cap == ref {
				found = true
				break
			}
		}
		if !found {
			return errors.Errorf("capture reference {%s} in target parameter not found in source pattern", ref)
		}
	}

	return nil
}

// extractCaptureNames extracts all capture group names from a pattern
func extractCaptureNames(pattern string) []string {
	var captures []string
	segments := strings.Split(pattern, ".")
	for _, segment := range segments {
		if strings.HasPrefix(segment, "{") && strings.HasSuffix(segment, "}") {
			name := segment[1 : len(segment)-1]
			captures = append(captures, name)
		}
	}
	return captures
}

// extractCaptureReferences extracts all capture references from a target parameter string
func extractCaptureReferences(targetParameter string) map[string]bool {
	refs := make(map[string]bool)
	// Find all {name} patterns
	re := regexp.MustCompile(`\{([a-zA-Z_][a-zA-Z0-9_]*)\}`)
	matches := re.FindAllStringSubmatch(targetParameter, -1)
	for _, match := range matches {
		if len(match) > 1 {
			refs[match[1]] = true
		}
	}
	return refs
}

// compilePatternToRegex compiles a pattern string to a regex (for future use in optimization)
// Returns the regex and list of capture names
func compilePatternToRegex(pattern string) (*regexp.Regexp, []string, error) {
	// For now, we don't use regex - we use manual traversal
	// This function is here for future optimization
	var captureNames []string
	segments := strings.Split(pattern, ".")
	regexParts := make([]string, 0, len(segments))

	for _, segment := range segments {
		if strings.HasPrefix(segment, "{") && strings.HasSuffix(segment, "}") {
			name := segment[1 : len(segment)-1]
			captureNames = append(captureNames, name)
			regexParts = append(regexParts, `([^\.]+)`)
		} else if segment == "*" {
			regexParts = append(regexParts, `[^\.]+`)
		} else {
			// Escape special regex characters
			escaped := regexp.QuoteMeta(segment)
			regexParts = append(regexParts, escaped)
		}
	}

	regexStr := "^" + strings.Join(regexParts, `\.`) + "$"
	re, err := regexp.Compile(regexStr)
	if err != nil {
		return nil, nil, errors.Wrapf(err, "failed to compile regex")
	}

	return re, captureNames, nil
}

// resolveTargetParameter resolves capture references in target parameter name
func resolveTargetParameter(targetParameter string, captures map[string]string) (string, error) {
	result := targetParameter
	re := regexp.MustCompile(`\{([a-zA-Z_][a-zA-Z0-9_]*)\}`)
	matches := re.FindAllStringSubmatch(targetParameter, -1)

	for _, match := range matches {
		if len(match) < 2 {
			continue
		}
		captureName := match[1]
		value, ok := captures[captureName]
		if !ok {
			return "", errors.Errorf("capture %q not found", captureName)
		}
		result = strings.ReplaceAll(result, match[0], value)
	}

	return result, nil
}

