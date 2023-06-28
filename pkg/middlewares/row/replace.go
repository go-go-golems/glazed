package row

import (
	"context"
	"github.com/go-go-golems/glazed/pkg/types"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
	"regexp"
	"strings"
)

type Replacement struct {
	Pattern     string
	Replacement string
}

type RegexReplacement struct {
	Regexp      *regexp.Regexp
	Replacement string
}

type Skip struct {
	Pattern string
}

type RegexpSkip struct {
	Regexp *regexp.Regexp
}

type ReplaceMiddleware struct {
	Replacements      map[types.FieldName][]*Replacement
	RegexReplacements map[types.FieldName][]*RegexpReplacement
	RegexSkips        map[types.FieldName][]*RegexpSkip
	Skips             map[types.FieldName][]*Skip
}

func NewReplaceMiddleware(
	replacements map[types.FieldName][]*Replacement,
	regexReplacements map[types.FieldName][]*RegexpReplacement,
	regexSkips map[types.FieldName][]*RegexpSkip,
	skips map[types.FieldName][]*Skip,
) *ReplaceMiddleware {
	return &ReplaceMiddleware{
		Replacements:      replacements,
		RegexReplacements: regexReplacements,
		RegexSkips:        regexSkips,
		Skips:             skips,
	}
}

func NewReplaceMiddlewareFromYAML(b []byte) (*ReplaceMiddleware, error) {
	var file interface{}
	err := yaml.Unmarshal(b, &file)
	if err != nil {
		return nil, err
	}

	replacements := make(map[types.FieldName][]*Replacement)
	regexReplacements := make(map[types.FieldName][]*RegexpReplacement)
	regexSkips := make(map[types.FieldName][]*RegexpSkip)
	skips := make(map[types.FieldName][]*Skip)

	fieldReplacements, ok := file.(map[string]interface{})
	if !ok {
		return nil, errors.New("invalid format")
	}

	for fieldName, fieldReplacements := range fieldReplacements {
		fieldReplacements, ok := fieldReplacements.(map[string]interface{})
		if !ok {
			return nil, errors.New("invalid format")
		}

		for replacementType, replacementValue := range fieldReplacements {
			switch replacementType {
			case "replace":
				rs_, ok := replacementValue.([]interface{})
				if !ok {
					return nil, errors.Errorf(
						"invalid value %v for replacements in field %s",
						replacementValue, fieldName)
				}
				for _, r_ := range rs_ {
					r, ok := r_.(map[string]interface{})
					if !ok {
						return nil, errors.Errorf(
							"invalid value %v for replacements in field %s",
							replacementValue, fieldName)
					}
					if len(r) != 1 {
						return nil, errors.Errorf(
							"invalid value %v for replacements in field %s",
							replacementValue, fieldName)
					}
					for pattern, replacement_ := range r {
						replacement, ok := replacement_.(string)
						if !ok {
							return nil, errors.Errorf(
								"invalid value %v for replacement in field %s",
								replacement_, fieldName)
						}
						replacements[types.FieldName(fieldName)] = append(
							replacements[types.FieldName(fieldName)],
							&Replacement{Pattern: pattern, Replacement: replacement})
					}
				}

			case "regex_replace":
				rs_, ok := replacementValue.([]interface{})
				if !ok {
					return nil, errors.Errorf(
						"invalid value %v for regex_replace in field %s",
						replacementValue, fieldName)
				}

				for _, r_ := range rs_ {
					r, ok := r_.(map[string]interface{})
					if !ok {
						return nil, errors.Errorf(
							"invalid value %v for regex_replace in field %s",
							replacementValue, fieldName)
					}
					if len(r) != 1 {
						return nil, errors.Errorf(
							"invalid value %v for regex_replace in field %s",
							replacementValue, fieldName)
					}
					for pattern, replacement_ := range r {
						replacement, ok := replacement_.(string)
						if !ok {
							return nil, errors.Errorf(
								"invalid value %v for replacement in field %s",
								replacement_, fieldName)
						}
						re, err := regexp.Compile(pattern)
						if err != nil {
							return nil, errors.Wrapf(err, "invalid regex %s in field %s", pattern, fieldName)
						}

						regexReplacements[types.FieldName(fieldName)] = append(
							regexReplacements[types.FieldName(fieldName)],
							&RegexpReplacement{Regexp: re, Replacement: replacement})
					}

				}
			case "skip":
				skipPatterns, ok := replacementValue.([]interface{})
				if !ok {
					return nil, errors.Errorf(
						"invalid value %v for skip in field %s",
						replacementValue, fieldName)
				}
				for _, pattern_ := range skipPatterns {
					pattern, ok := pattern_.(string)
					if !ok {
						return nil, errors.Errorf(
							"invalid value %v for skip in field %s",
							pattern_, fieldName)
					}
					skips[types.FieldName(fieldName)] = append(
						skips[types.FieldName(fieldName)],
						&Skip{Pattern: pattern})
				}
			case "regex_skip":
				skipPatterns, ok := replacementValue.([]interface{})
				if !ok {
					return nil, errors.Errorf(
						"invalid value %v for regex_skip in field %s",
						replacementValue, fieldName)
				}
				for _, pattern_ := range skipPatterns {
					pattern, ok := pattern_.(string)
					if !ok {
						return nil, errors.Errorf(
							"invalid value %v for regex_skip in field %s",
							pattern_, fieldName)
					}
					re, err := regexp.Compile(pattern)
					if err != nil {
						return nil, errors.Wrapf(err, "invalid regex %s in field %s", pattern, fieldName)
					}

					regexSkips[types.FieldName(fieldName)] = append(
						regexSkips[types.FieldName(fieldName)],
						&RegexpSkip{Regexp: re})
				}
			}
		}

	}

	return NewReplaceMiddleware(replacements, regexReplacements, regexSkips, skips), nil
}

func (r *ReplaceMiddleware) Process(ctx context.Context, row types.Row) ([]types.Row, error) {
	newRow := types.NewMapRow()

	for pair := row.Oldest(); pair != nil; pair = pair.Next() {
		rowField, value := pair.Key, pair.Value

		s, ok := value.(string)
		if !ok {
			newRow.Set(rowField, value)
			continue
		}

		for _, skip := range r.Skips[rowField] {
			if strings.Contains(s, skip.Pattern) {
				return nil, nil
			}
		}

		for _, regexSkip := range r.RegexSkips[rowField] {
			if regexSkip.Regexp.MatchString(s) {
				return nil, nil
			}
		}

		for _, replacement := range r.Replacements[rowField] {
			s = strings.ReplaceAll(s, replacement.Pattern, replacement.Replacement)
		}

		for _, regexReplacement := range r.RegexReplacements[rowField] {
			s = regexReplacement.Regexp.ReplaceAllString(s, regexReplacement.Replacement)
		}

		newRow.Set(rowField, s)
	}

	return []types.Row{newRow}, nil
}
