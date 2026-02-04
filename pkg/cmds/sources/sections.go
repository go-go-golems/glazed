package sources

import (
	"github.com/go-go-golems/glazed/pkg/cmds/schema"
	"github.com/go-go-golems/glazed/pkg/cmds/values"
	"github.com/pkg/errors"
)

// ReplaceSectionValues is a middleware that replaces parsed section values with a new one.
// It first calls next, then replaces the specified section with a clone of the provided one.
// If the section doesn't exist in the original values, it will be added.
func ReplaceSectionValues(sectionSlug string, newSection *values.SectionValues) Middleware {
	return func(next HandlerFunc) HandlerFunc {
		return func(schema_ *schema.Schema, parsedValues *values.Values) error {
			err := next(schema_, parsedValues)
			if err != nil {
				return err
			}

			if newSection == nil {
				return errors.New("cannot replace with nil section")
			}

			parsedValues.Set(sectionSlug, newSection.Clone())
			return nil
		}
	}
}

// ReplaceValues is a middleware that replaces multiple parsed sections at once.
// It first calls next, then replaces all specified sections with clones of the provided ones.
// If a section doesn't exist in the original values, it will be added.
func ReplaceValues(newValues *values.Values) Middleware {
	return func(next HandlerFunc) HandlerFunc {
		return func(schema_ *schema.Schema, parsedValues *values.Values) error {
			err := next(schema_, parsedValues)
			if err != nil {
				return err
			}

			if newValues == nil {
				return errors.New("cannot replace with nil values")
			}

			newValues.ForEach(func(k string, v *values.SectionValues) {
				parsedValues.Set(k, v.Clone())
			})
			return nil
		}
	}
}

// ReplaceValuesSelective is a middleware that replaces only the specified sections from the provided Values.
// It first calls next, then replaces only the sections specified in slugs with clones from newValues.
// If a section in slugs doesn't exist in newValues, it is skipped.
func ReplaceValuesSelective(newValues *values.Values, slugs []string) Middleware {
	return func(next HandlerFunc) HandlerFunc {
		return func(schema_ *schema.Schema, parsedValues *values.Values) error {
			err := next(schema_, parsedValues)
			if err != nil {
				return err
			}

			if newValues == nil {
				return errors.New("cannot replace with nil values")
			}

			for _, slug := range slugs {
				if sectionValues, ok := newValues.Get(slug); ok {
					parsedValues.Set(slug, sectionValues.Clone())
				}
			}
			return nil
		}
	}
}

// MergeSectionValues is a middleware that merges parsed section values into an existing one.
// It first calls next, then merges the provided values into the specified section.
// If the target section doesn't exist, it will be created.
func MergeSectionValues(sectionSlug string, sectionToMerge *values.SectionValues) Middleware {
	return func(next HandlerFunc) HandlerFunc {
		return func(schema_ *schema.Schema, parsedValues *values.Values) error {
			err := next(schema_, parsedValues)
			if err != nil {
				return err
			}

			if sectionToMerge == nil {
				return errors.New("cannot merge nil section")
			}

			targetSection, ok := parsedValues.Get(sectionSlug)
			if !ok {
				parsedValues.Set(sectionSlug, sectionToMerge.Clone())
				return nil
			}

			err = targetSection.MergeFields(sectionToMerge)
			if err != nil {
				return err
			}
			return nil
		}
	}
}

// MergeValues is a middleware that merges multiple parsed sections at once.
// It first calls next, then merges all provided values into the existing ones.
// If a target section doesn't exist, it will be created.
func MergeValues(valuesToMerge *values.Values) Middleware {
	return func(next HandlerFunc) HandlerFunc {
		return func(schema_ *schema.Schema, parsedValues *values.Values) error {
			err := next(schema_, parsedValues)
			if err != nil {
				return err
			}

			if valuesToMerge == nil {
				return errors.New("cannot merge nil values")
			}

			err = parsedValues.Merge(valuesToMerge)
			if err != nil {
				return err
			}
			return nil
		}
	}
}

// MergeValuesSelective is a middleware that merges only the specified sections from the provided Values.
// It first calls next, then merges only the sections specified in slugs from valuesToMerge into the existing values.
// If a section in slugs doesn't exist in valuesToMerge, it is skipped.
// If a target section doesn't exist in values, it will be created.
func MergeValuesSelective(valuesToMerge *values.Values, slugs []string) Middleware {
	return func(next HandlerFunc) HandlerFunc {
		return func(schema_ *schema.Schema, parsedValues *values.Values) error {
			err := next(schema_, parsedValues)
			if err != nil {
				return err
			}

			if valuesToMerge == nil {
				return errors.New("cannot merge nil values")
			}

			for _, slug := range slugs {
				if sectionValues, ok := valuesToMerge.Get(slug); ok {
					targetSection, exists := parsedValues.Get(slug)
					if !exists {
						parsedValues.Set(slug, sectionValues.Clone())
					} else {
						err = targetSection.MergeFields(sectionValues)
						if err != nil {
							return err
						}
					}
				}
			}
			return nil
		}
	}
}
