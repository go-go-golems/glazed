package middlewares

import (
	"github.com/go-go-golems/glazed/pkg/cmds/layers"
)

// WhitelistLayers only leaves the specified layers from the given ParameterLayers.
// It takes a slice of layer slugs, and deletes any layers in the ParameterLayers
// that don't match those slugs.
func WhitelistLayers(slugs []string) HandlerFunc {
	slugsToKeep := map[string]interface{}{}
	for _, s := range slugs {
		slugsToKeep[s] = nil
	}
	return func(layers_ *layers.ParameterLayers, parsedLayers *layers.ParsedLayers) error {
		toDelete := []string{}
		layers_.ForEach(func(key string, l layers.ParameterLayer) {
			if _, ok := slugsToKeep[key]; !ok {
				toDelete = append(toDelete, key)
			}
		})
		for _, key := range toDelete {
			layers_.Delete(key)
		}
		return nil
	}
}

func WhitelistLayerParameters(parameters_ map[string][]string) HandlerFunc {
	return func(layers_ *layers.ParameterLayers, parsedLayers *layers.ParsedLayers) error {
		layersToDelete := []string{}
		layersToUpdate := map[string]layers.ParameterLayer{}
		layers_.ForEach(func(key string, l layers.ParameterLayer) {
			if _, ok := parameters_[key]; !ok {
				layersToDelete = append(layersToDelete, key)
				return
			}

			parametersToKeep := map[string]interface{}{}
			for _, p := range parameters_[key] {
				parametersToKeep[p] = nil
			}
			layersToUpdate[key] = layers.NewWhitelistParameterLayer(l, parametersToKeep)
		})
		for _, key := range layersToDelete {
			layers_.Delete(key)
		}
		for key, l := range layersToUpdate {
			layers_.Set(key, l)
		}
		return nil
	}
}

// BlacklistLayers removes the specified layers from the given ParameterLayers.
// It takes a slice of layer slugs, and deletes any layers in the ParameterLayers
// that match those slugs.
func BlacklistLayers(slugs []string) HandlerFunc {
	slugsToDelete := map[string]interface{}{}
	for _, s := range slugs {
		slugsToDelete[s] = nil
	}
	return func(layers_ *layers.ParameterLayers, parsedLayers *layers.ParsedLayers) error {
		toDelete := []string{}
		layers_.ForEach(func(key string, l layers.ParameterLayer) {
			if _, ok := slugsToDelete[key]; ok {
				toDelete = append(toDelete, key)
			}
		})
		for _, key := range toDelete {
			layers_.Delete(key)
		}
		return nil
	}
}

func BlacklistLayerParameters(parameters_ map[string][]string) HandlerFunc {
	return func(layers_ *layers.ParameterLayers, parsedLayers *layers.ParsedLayers) error {
		layersToDelete := []string{}
		layersToUpdate := map[string]layers.ParameterLayer{}
		layers_.ForEach(func(key string, l layers.ParameterLayer) {
			if _, ok := parameters_[key]; !ok {
				return
			}

			parametersToKeep := map[string]interface{}{}
			for _, p := range parameters_[key] {
				parametersToKeep[p] = nil
			}
			layersToUpdate[key] = layers.NewBlacklistParameterLayer(l, parametersToKeep)
		})
		for _, key := range layersToDelete {
			layers_.Delete(key)
		}
		for key, l := range layersToUpdate {
			layers_.Set(key, l)
		}
		return nil
	}
}

// WrapWithLayerModifyingHandler wraps a middleware that modifies the layers
// with additional middlewares. It clones the original layers, calls the
// layer modifying middleware, chains any additional middlewares, calls
// next with the original layers, and returns any errors.
//
// This makes it possible to restrict a set of middlewares to only apply to a
// restricted subset of layers. However, the normal set of middlewares is allowed
// to continue as normal.
func WrapWithLayerModifyingHandler(m HandlerFunc, nextMiddlewares ...Middleware) Middleware {
	return func(next HandlerFunc) HandlerFunc {
		return func(layers_ *layers.ParameterLayers, parsedLayers *layers.ParsedLayers) error {
			err := next(layers_, parsedLayers)
			if err != nil {
				return err
			}

			chain := Chain(nextMiddlewares...)

			clonedLayers := layers_.Clone()
			err = m(clonedLayers, parsedLayers)
			if err != nil {
				return err
			}

			err = chain(Identity)(clonedLayers, parsedLayers)
			if err != nil {
				return err
			}

			return nil
		}
	}
}

// WrapWithWhitelistedLayers wraps a middleware that restricts layers
// to a specified set of slugs, with any additional middlewares.
// It makes it possible to apply a subset of middlewares to only
// certain restricted layers.
func WrapWithWhitelistedLayers(slugs []string, nextMiddlewares ...Middleware) Middleware {
	return WrapWithLayerModifyingHandler(WhitelistLayers(slugs), nextMiddlewares...)
}

func WrapWithWhitelistedParameterLayers(parameters_ map[string][]string, nextMiddlewares ...Middleware) Middleware {
	return WrapWithLayerModifyingHandler(WhitelistLayerParameters(parameters_), nextMiddlewares...)
}

// WrapWithBlacklistedLayers wraps a middleware that restricts layers
// to a specified set of slugs, with any additional middlewares.
// It makes it possible to apply a subset of middlewares to only
// certain restricted layers.
func WrapWithBlacklistedLayers(slugs []string, nextMiddlewares ...Middleware) Middleware {
	return WrapWithLayerModifyingHandler(BlacklistLayers(slugs), nextMiddlewares...)
}

func WrapWithBlacklistedParameterLayers(parameters_ map[string][]string, nextMiddlewares ...Middleware) Middleware {
	return WrapWithLayerModifyingHandler(BlacklistLayerParameters(parameters_), nextMiddlewares...)
}
