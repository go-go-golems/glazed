package middlewares

import (
	"github.com/go-go-golems/glazed/pkg/cmds/layers"
	"github.com/pkg/errors"
)

// ReplaceParsedLayer is a middleware that replaces a parsed layer with a new one.
// It first calls next, then replaces the specified layer with a clone of the provided one.
// If the layer doesn't exist in the original parsedLayers, it will be added.
func ReplaceParsedLayer(layerSlug string, newLayer *layers.ParsedLayer) Middleware {
	return func(next HandlerFunc) HandlerFunc {
		return func(layers_ *layers.ParameterLayers, parsedLayers *layers.ParsedLayers) error {
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

// ReplaceParsedLayers is a middleware that replaces multiple parsed layers at once.
// It first calls next, then replaces all specified layers with clones of the provided ones.
// If a layer doesn't exist in the original parsedLayers, it will be added.
func ReplaceParsedLayers(newLayers *layers.ParsedLayers) Middleware {
	return func(next HandlerFunc) HandlerFunc {
		return func(layers_ *layers.ParameterLayers, parsedLayers *layers.ParsedLayers) error {
			err := next(layers_, parsedLayers)
			if err != nil {
				return err
			}

			if newLayers == nil {
				return errors.New("cannot replace with nil layers")
			}

			newLayers.ForEach(func(k string, v *layers.ParsedLayer) {
				parsedLayers.Set(k, v.Clone())
			})
			return nil
		}
	}
}

// ReplaceParsedLayersSelective is a middleware that replaces only the specified layers from the provided ParsedLayers.
// It first calls next, then replaces only the layers specified in slugs with clones from newLayers.
// If a layer in slugs doesn't exist in newLayers, it is skipped.
func ReplaceParsedLayersSelective(newLayers *layers.ParsedLayers, slugs []string) Middleware {
	return func(next HandlerFunc) HandlerFunc {
		return func(layers_ *layers.ParameterLayers, parsedLayers *layers.ParsedLayers) error {
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

// MergeParsedLayer is a middleware that merges a parsed layer into an existing one.
// It first calls next, then merges the provided layer into the specified one.
// If the target layer doesn't exist, it will be created.
func MergeParsedLayer(layerSlug string, layerToMerge *layers.ParsedLayer) Middleware {
	return func(next HandlerFunc) HandlerFunc {
		return func(layers_ *layers.ParameterLayers, parsedLayers *layers.ParsedLayers) error {
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

// MergeParsedLayers is a middleware that merges multiple parsed layers at once.
// It first calls next, then merges all provided layers into the existing ones.
// If a target layer doesn't exist, it will be created.
func MergeParsedLayers(layersToMerge *layers.ParsedLayers) Middleware {
	return func(next HandlerFunc) HandlerFunc {
		return func(layers_ *layers.ParameterLayers, parsedLayers *layers.ParsedLayers) error {
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

// MergeParsedLayersSelective is a middleware that merges only the specified layers from the provided ParsedLayers.
// It first calls next, then merges only the layers specified in slugs from layersToMerge into the existing layers.
// If a layer in slugs doesn't exist in layersToMerge, it is skipped.
// If a target layer doesn't exist in parsedLayers, it will be created.
func MergeParsedLayersSelective(layersToMerge *layers.ParsedLayers, slugs []string) Middleware {
	return func(next HandlerFunc) HandlerFunc {
		return func(layers_ *layers.ParameterLayers, parsedLayers *layers.ParsedLayers) error {
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
