package sources

import (
	"github.com/go-go-golems/glazed/pkg/cmds/schema"
	"github.com/go-go-golems/glazed/pkg/cmds/values"
	"github.com/pkg/errors"
)

// ReplaceSectionValues is a middleware that replaces a parsed layer with a new one.
// It first calls next, then replaces the specified layer with a clone of the provided one.
// If the layer doesn't exist in the original parsedLayers, it will be added.
func ReplaceSectionValues(layerSlug string, newLayer *values.SectionValues) Middleware {
	return func(next HandlerFunc) HandlerFunc {
		return func(layers_ *schema.Schema, parsedLayers *values.Values) error {
			err := next(layers_, parsedLayers)
			if err != nil {
				return err
			}

			if newLayer == nil {
				return errors.New("cannot replace with nil layer")
			}

			parsedLayers.Set(layerSlug, newLayer.Clone())
			return nil
		}
	}
}

// ReplaceValues is a middleware that replaces multiple parsed layers at once.
// It first calls next, then replaces all specified layers with clones of the provided ones.
// If a layer doesn't exist in the original parsedLayers, it will be added.
func ReplaceValues(newLayers *values.Values) Middleware {
	return func(next HandlerFunc) HandlerFunc {
		return func(layers_ *schema.Schema, parsedLayers *values.Values) error {
			err := next(layers_, parsedLayers)
			if err != nil {
				return err
			}

			if newLayers == nil {
				return errors.New("cannot replace with nil layers")
			}

			newLayers.ForEach(func(k string, v *values.SectionValues) {
				parsedLayers.Set(k, v.Clone())
			})
			return nil
		}
	}
}

// ReplaceValuesSelective is a middleware that replaces only the specified layers from the provided Values.
// It first calls next, then replaces only the layers specified in slugs with clones from newLayers.
// If a layer in slugs doesn't exist in newLayers, it is skipped.
func ReplaceValuesSelective(newLayers *values.Values, slugs []string) Middleware {
	return func(next HandlerFunc) HandlerFunc {
		return func(layers_ *schema.Schema, parsedLayers *values.Values) error {
			err := next(layers_, parsedLayers)
			if err != nil {
				return err
			}

			if newLayers == nil {
				return errors.New("cannot replace with nil layers")
			}

			for _, slug := range slugs {
				if layer, ok := newLayers.Get(slug); ok {
					parsedLayers.Set(slug, layer.Clone())
				}
			}
			return nil
		}
	}
}

// MergeSectionValues is a middleware that merges a parsed layer into an existing one.
// It first calls next, then merges the provided layer into the specified one.
// If the target layer doesn't exist, it will be created.
func MergeSectionValues(layerSlug string, layerToMerge *values.SectionValues) Middleware {
	return func(next HandlerFunc) HandlerFunc {
		return func(layers_ *schema.Schema, parsedLayers *values.Values) error {
			err := next(layers_, parsedLayers)
			if err != nil {
				return err
			}

			if layerToMerge == nil {
				return errors.New("cannot merge nil layer")
			}

			targetLayer, ok := parsedLayers.Get(layerSlug)
			if !ok {
				parsedLayers.Set(layerSlug, layerToMerge.Clone())
				return nil
			}

			err = targetLayer.MergeParameters(layerToMerge)
			if err != nil {
				return err
			}
			return nil
		}
	}
}

// MergeValues is a middleware that merges multiple parsed layers at once.
// It first calls next, then merges all provided layers into the existing ones.
// If a target layer doesn't exist, it will be created.
func MergeValues(layersToMerge *values.Values) Middleware {
	return func(next HandlerFunc) HandlerFunc {
		return func(layers_ *schema.Schema, parsedLayers *values.Values) error {
			err := next(layers_, parsedLayers)
			if err != nil {
				return err
			}

			if layersToMerge == nil {
				return errors.New("cannot merge nil layers")
			}

			err = parsedLayers.Merge(layersToMerge)
			if err != nil {
				return err
			}
			return nil
		}
	}
}

// MergeValuesSelective is a middleware that merges only the specified layers from the provided Values.
// It first calls next, then merges only the layers specified in slugs from layersToMerge into the existing layers.
// If a layer in slugs doesn't exist in layersToMerge, it is skipped.
// If a target layer doesn't exist in parsedLayers, it will be created.
func MergeValuesSelective(layersToMerge *values.Values, slugs []string) Middleware {
	return func(next HandlerFunc) HandlerFunc {
		return func(layers_ *schema.Schema, parsedLayers *values.Values) error {
			err := next(layers_, parsedLayers)
			if err != nil {
				return err
			}

			if layersToMerge == nil {
				return errors.New("cannot merge nil layers")
			}

			for _, slug := range slugs {
				if layer, ok := layersToMerge.Get(slug); ok {
					targetLayer, exists := parsedLayers.Get(slug)
					if !exists {
						parsedLayers.Set(slug, layer.Clone())
					} else {
						err = targetLayer.MergeParameters(layer)
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
