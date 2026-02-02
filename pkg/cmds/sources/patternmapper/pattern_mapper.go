package patternmapper

import (
	"fmt"
	"github.com/go-go-golems/glazed/pkg/cmds/schema"
	sources "github.com/go-go-golems/glazed/pkg/cmds/sources"
	"github.com/iancoleman/orderedmap"
	"github.com/pkg/errors"
	"os"
	"regexp"
	"sort"
	"strings"
)

// MappingRule defines a pattern-based mapping from config file structure to layer fields.
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

// patternMapper implements ConfigMapper using pattern matching rules
type patternMapper struct {
	rules            []MappingRule
	layers           *schema.Schema
	compiledPatterns []compiledPattern
}

// compiledPattern represents a compiled pattern with its capture groups
type compiledPattern struct {
	rule         MappingRule
	pattern      *regexp.Regexp
	captureNames []string // ordered list of capture group names
}

// (Compatibility options removed) Simplified: ambiguous cases error by default.

// NewConfigMapper creates a new pattern-based config mapper from the given rules.
// The mapper validates that:
//   - All patterns are valid syntax
//   - All target parameters exist in their respective layers
//   - Capture references in target parameters match captures in source patterns
func NewConfigMapper(layers *schema.Schema, rules ...MappingRule) (sources.ConfigMapper, error) {
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

	// Proposal 6: Warn on capture shadowing in nested rules
	if len(parentCaptures) > 0 && len(ruleCaptures) > 0 {
		duplicates := make(map[string]bool)
		for _, pc := range parentCaptures {
			for _, rc := range ruleCaptures {
				if pc == rc {
					duplicates[pc] = true
				}
			}
		}
		if len(duplicates) > 0 {
			names := make([]string, 0, len(duplicates))
			for n := range duplicates {
				names = append(names, n)
			}
			fmt.Fprintf(os.Stderr, "WARNING: capture shadowing detected for %v in nested rule %q under %q\n", names, rule.Source, parentPath)
		}
	}

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

		// Proposal 5: Early validation for static target parameters (no capture refs)
		if len(extractCaptureReferences(rule.TargetParameter)) == 0 {
			layer, _ := m.layers.Get(rule.TargetLayer)
			if layer != nil {
				canonical := resolveCanonicalParameterName(layer, rule.TargetParameter)
				if pd, ok := layer.GetDefinitions().Get(canonical); !ok || pd == nil {
					if canonical != rule.TargetParameter {
						return nil, errors.Errorf("target parameter %q (checked as %q) does not exist in layer %q", rule.TargetParameter, canonical, rule.TargetLayer)
					}
					return nil, errors.Errorf("target parameter %q does not exist in layer %q", rule.TargetParameter, rule.TargetLayer)
				}
			}
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

	// Convert to ordered map tree for deterministic traversal
	orderedRoot := toOrderedMap(configMap)

	// Track collisions across rules (proposal 3)
	// Key: layer+"."+paramName, Value: pattern source that last wrote to it
	collisionTracker := make(map[string]string)

	// Match each pattern against the config
	for _, compiled := range m.compiledPatterns {
		matches, err := m.matchPattern(compiled, orderedRoot, "")
		if err != nil {
			return nil, err
		}

		// Proposal 2: Track multi-matches per rule
		// Key: resolved target parameter name, Value: list of distinct values
		multiMatchTracker := make(map[string][]interface{})

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

			// Resolve canonical parameter name (using helper from proposal 9)
			paramName := resolveCanonicalParameterName(layer, targetParam)

			paramDef, ok := layer.GetDefinitions().Get(paramName)
			if !ok || paramDef == nil {
				// Proposal 4: Prefix-aware error messages
				// Include both the user-provided targetParam and the resolved paramName
				errorMsg := fmt.Sprintf("target parameter %q", targetParam)
				if paramName != targetParam {
					errorMsg += fmt.Sprintf(" (checked as %q)", paramName)
				}
				errorMsg += fmt.Sprintf(" does not exist in layer %q (pattern: %q)", match.layer, compiled.rule.Source)
				return nil, errors.New(errorMsg)
			}

			// Track multi-matches for this rule
			multiMatchTracker[paramName] = append(multiMatchTracker[paramName], match.value)
		}

		// Check for multi-matches (proposal 2)
		for paramName, values := range multiMatchTracker {
			if len(values) > 1 {
				// Check if values are distinct
				distinctValues := make(map[interface{}]bool)
				for _, v := range values {
					distinctValues[v] = true
				}

				if len(distinctValues) > 1 {
					// Multiple distinct values found: error
					return nil, errors.Errorf(
						"pattern %q matched multiple distinct values for parameter %q: found %d distinct values",
						compiled.rule.Source,
						paramName,
						len(distinctValues),
					)
				}
			}
		}

		// Process matches and set values
		// Track which parameters were written by this rule to avoid false collision detection
		writtenByThisRule := make(map[string]bool)
		for _, match := range matches {
			// Resolve target parameter name (replace captures)
			targetParam, err := resolveTargetParameter(compiled.rule.TargetParameter, match.captures)
			if err != nil {
				return nil, errors.Wrapf(err, "failed to resolve target parameter")
			}

			layer, _ := m.layers.Get(match.layer)
			paramName := resolveCanonicalParameterName(layer, targetParam)

			// Proposal 3: Collision detection across rules
			// Only check for collisions if this parameter wasn't already written by this rule
			collisionKey := match.layer + "." + paramName
			if !writtenByThisRule[collisionKey] {
				if previousPattern, exists := collisionTracker[collisionKey]; exists {
					// Collision detected (different rule writing to same parameter): error
					return nil, errors.Errorf(
						"collision: parameter %q in layer %q is written by multiple patterns: %q and %q",
						paramName,
						match.layer,
						previousPattern,
						compiled.rule.Source,
					)
				}
				collisionTracker[collisionKey] = compiled.rule.Source
				writtenByThisRule[collisionKey] = true
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
	config interface{},
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

	// If no matches and required, return error with context (Proposal 8)
	if len(matches) == 0 && compiled.rule.Required {
		nearest, missing, keys := nearestExistingPathInfo(compiled.rule.Source, config)
		extra := ""
		if nearest != "" || missing != "" {
			extra = fmt.Sprintf("; nearest existing path: %q; missing segment: %q", nearest, missing)
			if len(keys) > 0 {
				extra += fmt.Sprintf("; available keys: %v", keys)
			}
		}
		return nil, errors.Errorf(
			"required pattern %q did not match any paths in config%s",
			compiled.rule.Source,
			extra,
		)
	}

	return matches, nil
}

// traverseAndMatch recursively traverses the config and matches patterns
func (m *patternMapper) traverseAndMatch(
	compiled compiledPattern,
	config interface{},
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
	config interface{},
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
		// Match all keys at this level in deterministic order
		keys, getter, ok := iterMap(config)
		if !ok {
			return nil
		}
		for _, key := range keys {
			value, _ := getter(key)
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
		// Match all keys at this level in deterministic order
		keys, getter, ok := iterMap(config)
		if !ok {
			return nil
		}
		for _, key := range keys {
			value, _ := getter(key)
			if err := m.matchSegmentsRecursive(remaining, value, currentPath+"."+key, captures, compiled, matches); err != nil {
				return err
			}
		}
		return nil
	}

	// Exact match
	if _, getter, ok := iterMap(config); ok {
		// Fast path direct get
		value, exists := getter(segment)
		if !exists {
			return nil // No match, but not an error (might be optional)
		}

		newPath := segment
		if currentPath != "" {
			newPath = currentPath + "." + segment
		}

		return m.matchSegmentsRecursive(remaining, value, newPath, captures, compiled, matches)
	}
	return nil
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
	return m.matchSegments(remaining, value, currentPath, captures, compiled, matches)
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

// resolveCanonicalParameterName resolves the canonical parameter name including prefix
// This is proposal 9: explicit helper for canonical parameter name resolution
func resolveCanonicalParameterName(layer schema.Section, targetParam string) string {
	if layer.GetPrefix() != "" {
		// If layer has prefix, check if targetParam already includes it
		if !strings.HasPrefix(targetParam, layer.GetPrefix()) {
			return layer.GetPrefix() + targetParam
		}
	}
	return targetParam
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

// nearestExistingPathInfo returns a hint about where a required pattern stopped matching.
// It returns the nearest existing path prefix, the missing segment, and available keys at that point.
func nearestExistingPathInfo(pattern string, config interface{}) (string, string, []string) {
	segments := strings.Split(pattern, ".")
	var parts []string
	current := config

	for _, seg := range segments {
		// If current is not a map-like, we can't go deeper
		keys, getter, ok := iterMap(current)
		if !ok {
			return strings.Join(parts, "."), seg, nil
		}

		// Capture/wildcard: if there are no keys, we fail here; otherwise, we can't disambiguate further
		if seg == "*" || (strings.HasPrefix(seg, "{") && strings.HasSuffix(seg, "}")) {
			if len(keys) == 0 {
				return strings.Join(parts, "."), seg, keys
			}
			// choose not to go deeper to avoid misleading path; report at this level
			return strings.Join(parts, "."), seg, keys
		}

		// Literal segment
		if v, exists := getter(seg); exists {
			parts = append(parts, seg)
			current = v
			continue
		}
		// Missing literal key
		return strings.Join(parts, "."), seg, keys
	}

	// All segments consumed but no match recorded; fall back
	return strings.Join(parts, "."), "", nil
}

// toOrderedMap converts a map[string]interface{} into an ordered map with
// lexicographically sorted keys. Nested maps are converted recursively.
func toOrderedMap(m map[string]interface{}) *orderedmap.OrderedMap {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	om := orderedmap.New()
	for _, k := range keys {
		om.Set(k, toOrderedValue(m[k]))
	}
	return om
}

func toOrderedValue(v interface{}) interface{} {
	switch t := v.(type) {
	case map[string]interface{}:
		return toOrderedMap(t)
	case []interface{}:
		// Convert nested maps inside arrays as well for consistency
		out := make([]interface{}, len(t))
		for i, e := range t {
			out[i] = toOrderedValue(e)
		}
		return out
	default:
		return v
	}
}

// iterMap returns a deterministic key order and a getter for the provided map-like value.
// Supports *orderedmap.OrderedMap and map[string]interface{}.
func iterMap(value interface{}) ([]string, func(string) (interface{}, bool), bool) {
	if om, ok := value.(*orderedmap.OrderedMap); ok {
		keys := om.Keys()
		getter := func(k string) (interface{}, bool) {
			v, ok := om.Get(k)
			return v, ok
		}
		return keys, getter, true
	}
	if m, ok := value.(map[string]interface{}); ok {
		keys := make([]string, 0, len(m))
		for k := range m {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		getter := func(k string) (interface{}, bool) {
			v, ok := m[k]
			return v, ok
		}
		return keys, getter, true
	}
	return nil, nil, false
}
