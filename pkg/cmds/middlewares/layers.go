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

			targetLayer.MergeParameters(layerToMerge)
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

			parsedLayers.Merge(layersToMerge)
			return nil
		}
	}
}
