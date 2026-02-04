package sources

import (
	"github.com/go-go-golems/glazed/pkg/cmds/schema"
	"github.com/go-go-golems/glazed/pkg/cmds/values"
)

// WhitelistSectionsHandler only leaves the specified sections from the given schema.
// It takes a slice of section slugs, and deletes any sections in the schema
// that don't match those slugs.
func WhitelistSectionsHandler(slugs []string) HandlerFunc {
	slugsToKeep := map[string]interface{}{}
	for _, s := range slugs {
		slugsToKeep[s] = nil
	}
	return func(schema_ *schema.Schema, parsedValues *values.Values) error {
		toDelete := []string{}
		schema_.ForEach(func(key string, l schema.Section) {
			if _, ok := slugsToKeep[key]; !ok {
				toDelete = append(toDelete, key)
			}
		})
		for _, key := range toDelete {
			schema_.Delete(key)
		}
		return nil
	}
}

// WhitelistSectionFieldsHandler restricts each section to the specified field names.
func WhitelistSectionFieldsHandler(fieldsBySection map[string][]string) HandlerFunc {
	return func(schema_ *schema.Schema, parsedValues *values.Values) error {
		sectionsToDelete := []string{}
		sectionsToUpdate := map[string]schema.Section{}
		schema_.ForEach(func(key string, l schema.Section) {
			if _, ok := fieldsBySection[key]; !ok {
				sectionsToDelete = append(sectionsToDelete, key)
				return
			}

			fieldsToKeep := map[string]interface{}{}
			for _, fieldName := range fieldsBySection[key] {
				fieldsToKeep[fieldName] = nil
			}
			sectionsToUpdate[key] = schema.NewWhitelistSection(l, fieldsToKeep)
		})
		for _, key := range sectionsToDelete {
			schema_.Delete(key)
		}
		for key, l := range sectionsToUpdate {
			schema_.Set(key, l)
		}
		return nil
	}
}

// WhitelistSections applies a section whitelist after running next.
func WhitelistSections(slugs []string) Middleware {
	return func(next HandlerFunc) HandlerFunc {
		return func(schema_ *schema.Schema, parsedValues *values.Values) error {
			err := next(schema_, parsedValues)
			if err != nil {
				return err
			}

			return WhitelistSectionsHandler(slugs)(schema_, parsedValues)
		}
	}
}

// WhitelistSectionsFirst applies a section whitelist before running next.
func WhitelistSectionsFirst(slugs []string) Middleware {
	return func(next HandlerFunc) HandlerFunc {
		return func(schema_ *schema.Schema, parsedValues *values.Values) error {
			err := WhitelistSectionsHandler(slugs)(schema_, parsedValues)
			if err != nil {
				return err
			}

			return next(schema_, parsedValues)
		}
	}
}

// WhitelistSectionFields applies a field whitelist per section after running next.
func WhitelistSectionFields(fieldsBySection map[string][]string) Middleware {
	return func(next HandlerFunc) HandlerFunc {
		return func(schema_ *schema.Schema, parsedValues *values.Values) error {
			err := next(schema_, parsedValues)
			if err != nil {
				return err
			}

			return WhitelistSectionFieldsHandler(fieldsBySection)(schema_, parsedValues)
		}
	}
}

// WhitelistSectionFieldsFirst applies a field whitelist per section before running next.
func WhitelistSectionFieldsFirst(fieldsBySection map[string][]string) Middleware {
	return func(next HandlerFunc) HandlerFunc {
		return func(schema_ *schema.Schema, parsedValues *values.Values) error {
			err := WhitelistSectionFieldsHandler(fieldsBySection)(schema_, parsedValues)
			if err != nil {
				return err
			}

			return next(schema_, parsedValues)
		}
	}
}

// BlacklistSectionsHandler removes the specified sections from the given schema.
// It takes a slice of section slugs, and deletes any sections in the schema
// that match those slugs.
func BlacklistSectionsHandler(slugs []string) HandlerFunc {
	slugsToDelete := map[string]interface{}{}
	for _, s := range slugs {
		slugsToDelete[s] = nil
	}
	return func(schema_ *schema.Schema, parsedValues *values.Values) error {
		toDelete := []string{}
		schema_.ForEach(func(key string, l schema.Section) {
			if _, ok := slugsToDelete[key]; ok {
				toDelete = append(toDelete, key)
			}
		})
		for _, key := range toDelete {
			schema_.Delete(key)
		}
		return nil
	}
}

// BlacklistSectionFieldsHandler removes the specified fields per section.
func BlacklistSectionFieldsHandler(fieldsBySection map[string][]string) HandlerFunc {
	return func(schema_ *schema.Schema, parsedValues *values.Values) error {
		sectionsToDelete := []string{}
		sectionsToUpdate := map[string]schema.Section{}
		schema_.ForEach(func(key string, l schema.Section) {
			if _, ok := fieldsBySection[key]; !ok {
				return
			}

			fieldsToKeep := map[string]interface{}{}
			for _, fieldName := range fieldsBySection[key] {
				fieldsToKeep[fieldName] = nil
			}
			sectionsToUpdate[key] = schema.NewBlacklistSection(l, fieldsToKeep)
		})
		for _, key := range sectionsToDelete {
			schema_.Delete(key)
		}
		for key, l := range sectionsToUpdate {
			schema_.Set(key, l)
		}
		return nil
	}
}

// BlacklistSections removes the given sections after running next.
func BlacklistSections(slugs []string) Middleware {
	return func(next HandlerFunc) HandlerFunc {
		return func(schema_ *schema.Schema, parsedValues *values.Values) error {
			err := next(schema_, parsedValues)
			if err != nil {
				return err
			}

			return BlacklistSectionsHandler(slugs)(schema_, parsedValues)
		}
	}
}

// BlacklistSectionsFirst removes the given sections before running next.
func BlacklistSectionsFirst(slugs []string) Middleware {
	return func(next HandlerFunc) HandlerFunc {
		return func(schema_ *schema.Schema, parsedValues *values.Values) error {
			err := next(schema_, parsedValues)
			if err != nil {
				return err
			}

			return BlacklistSectionsHandler(slugs)(schema_, parsedValues)
		}
	}
}

// BlacklistSectionFields removes the given fields after running next.
func BlacklistSectionFields(fieldsBySection map[string][]string) Middleware {
	return func(next HandlerFunc) HandlerFunc {
		return func(schema_ *schema.Schema, parsedValues *values.Values) error {
			err := next(schema_, parsedValues)
			if err != nil {
				return err
			}

			return BlacklistSectionFieldsHandler(fieldsBySection)(schema_, parsedValues)
		}
	}
}

// BlacklistSectionFieldsFirst removes the given fields before running next.
func BlacklistSectionFieldsFirst(fieldsBySection map[string][]string) Middleware {
	return func(next HandlerFunc) HandlerFunc {
		return func(schema_ *schema.Schema, parsedValues *values.Values) error {
			err := BlacklistSectionFieldsHandler(fieldsBySection)(schema_, parsedValues)
			if err != nil {
				return err
			}

			return next(schema_, parsedValues)
		}
	}
}

// WrapWithSectionModifyingHandler wraps a middleware that modifies the schema
// with additional middlewares. It clones the original schema, calls the
// section-modifying middleware, chains any additional middlewares, calls
// next with the original schema, and returns any errors.
//
// This makes it possible to restrict a set of middlewares to only apply to a
// restricted subset of sections. However, the normal set of middlewares is allowed
// to continue as normal.
func WrapWithSectionModifyingHandler(m HandlerFunc, nextMiddlewares ...Middleware) Middleware {
	return func(next HandlerFunc) HandlerFunc {
		return func(schema_ *schema.Schema, parsedValues *values.Values) error {
			err := next(schema_, parsedValues)
			if err != nil {
				return err
			}

			chain := Chain(nextMiddlewares...)

			clonedSchema := schema_.Clone()
			err = m(clonedSchema, parsedValues)
			if err != nil {
				return err
			}

			err = chain(Identity)(clonedSchema, parsedValues)
			if err != nil {
				return err
			}

			return nil
		}
	}
}

// WrapWithWhitelistedSections wraps a middleware that restricts sections
// to a specified set of slugs, with any additional middlewares.
// It makes it possible to apply a subset of middlewares to only
// certain restricted sections.
func WrapWithWhitelistedSections(slugs []string, nextMiddlewares ...Middleware) Middleware {
	return WrapWithSectionModifyingHandler(WhitelistSectionsHandler(slugs), nextMiddlewares...)
}

// WrapWithWhitelistedSectionFields restricts fields within a subset of sections.
func WrapWithWhitelistedSectionFields(fieldsBySection map[string][]string, nextMiddlewares ...Middleware) Middleware {
	return WrapWithSectionModifyingHandler(WhitelistSectionFieldsHandler(fieldsBySection), nextMiddlewares...)
}

// WrapWithBlacklistedSections wraps a middleware that restricts sections
// to a specified set of slugs, with any additional middlewares.
// It makes it possible to apply a subset of middlewares to only
// certain restricted sections.
func WrapWithBlacklistedSections(slugs []string, nextMiddlewares ...Middleware) Middleware {
	return WrapWithSectionModifyingHandler(BlacklistSectionsHandler(slugs), nextMiddlewares...)
}

// WrapWithBlacklistedSectionFields removes fields within a subset of sections.
func WrapWithBlacklistedSectionFields(fieldsBySection map[string][]string, nextMiddlewares ...Middleware) Middleware {
	return WrapWithSectionModifyingHandler(BlacklistSectionFieldsHandler(fieldsBySection), nextMiddlewares...)
}
