package sources

import (
	"github.com/go-go-golems/glazed/pkg/cmds/schema"
	"github.com/go-go-golems/glazed/pkg/cmds/values"
)

// WhitelistLayersHandler only leaves the specified layers from the given ParameterLayers.
// It takes a slice of layer slugs, and deletes any layers in the ParameterLayers
// that don't match those slugs.
func WhitelistLayersHandler(slugs []string) HandlerFunc {
	slugsToKeep := map[string]interface{}{}
	for _, s := range slugs {
		slugsToKeep[s] = nil
	}
	return func(layers_ *schema.Schema, parsedLayers *values.Values) error {
		toDelete := []string{}
		layers_.ForEach(func(key string, l schema.Section) {
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

func WhitelistLayerParametersHandler(parameters_ map[string][]string) HandlerFunc {
	return func(layers_ *schema.Schema, parsedLayers *values.Values) error {
		layersToDelete := []string{}
		layersToUpdate := map[string]schema.Section{}
		layers_.ForEach(func(key string, l schema.Section) {
			if _, ok := parameters_[key]; !ok {
				layersToDelete = append(layersToDelete, key)
				return
			}

			parametersToKeep := map[string]interface{}{}
			for _, p := range parameters_[key] {
				parametersToKeep[p] = nil
			}
			layersToUpdate[key] = schema.NewWhitelistParameterLayer(l, parametersToKeep)
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

func WhitelistLayers(slugs []string) Middleware {
	return func(next HandlerFunc) HandlerFunc {
		return func(layers_ *schema.Schema, parsedLayers *values.Values) error {
			err := next(layers_, parsedLayers)
			if err != nil {
				return err
			}

			return WhitelistLayersHandler(slugs)(layers_, parsedLayers)
		}
	}
}

func WhitelistLayersFirst(slugs []string) Middleware {
	return func(next HandlerFunc) HandlerFunc {
		return func(layers_ *schema.Schema, parsedLayers *values.Values) error {
			err := WhitelistLayersHandler(slugs)(layers_, parsedLayers)
			if err != nil {
				return err
			}

			return next(layers_, parsedLayers)
		}
	}
}

func WhitelistLayerParameters(parameters_ map[string][]string) Middleware {
	return func(next HandlerFunc) HandlerFunc {
		return func(layers_ *schema.Schema, parsedLayers *values.Values) error {
			err := next(layers_, parsedLayers)
			if err != nil {
				return err
			}

			return WhitelistLayerParametersHandler(parameters_)(layers_, parsedLayers)
		}
	}
}

func WhitelistLayerParametersFirst(parameters_ map[string][]string) Middleware {
	return func(next HandlerFunc) HandlerFunc {
		return func(layers_ *schema.Schema, parsedLayers *values.Values) error {
			err := WhitelistLayerParametersHandler(parameters_)(layers_, parsedLayers)
			if err != nil {
				return err
			}

			return next(layers_, parsedLayers)
		}
	}
}

// BlacklistLayersHandler removes the specified layers from the given ParameterLayers.
// It takes a slice of layer slugs, and deletes any layers in the ParameterLayers
// that match those slugs.
func BlacklistLayersHandler(slugs []string) HandlerFunc {
	slugsToDelete := map[string]interface{}{}
	for _, s := range slugs {
		slugsToDelete[s] = nil
	}
	return func(layers_ *schema.Schema, parsedLayers *values.Values) error {
		toDelete := []string{}
		layers_.ForEach(func(key string, l schema.Section) {
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

func BlacklistLayerParametersHandler(parameters_ map[string][]string) HandlerFunc {
	return func(layers_ *schema.Schema, parsedLayers *values.Values) error {
		layersToDelete := []string{}
		layersToUpdate := map[string]schema.Section{}
		layers_.ForEach(func(key string, l schema.Section) {
			if _, ok := parameters_[key]; !ok {
				return
			}

			parametersToKeep := map[string]interface{}{}
			for _, p := range parameters_[key] {
				parametersToKeep[p] = nil
			}
			layersToUpdate[key] = schema.NewBlacklistParameterLayer(l, parametersToKeep)
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

// BlacklistLayers is a middleware that removes the given layers from ParameterLayers after running `next`.
func BlacklistLayers(slugs []string) Middleware {
	return func(next HandlerFunc) HandlerFunc {
		return func(layers_ *schema.Schema, parsedLayers *values.Values) error {
			err := next(layers_, parsedLayers)
			if err != nil {
				return err
			}

			return BlacklistLayersHandler(slugs)(layers_, parsedLayers)
		}
	}
}

// BlacklistLayersFirst is a middleware that removes the given layers from ParameterLayers before running `next`.
func BlacklistLayersFirst(slugs []string) Middleware {
	return func(next HandlerFunc) HandlerFunc {
		return func(layers_ *schema.Schema, parsedLayers *values.Values) error {
			err := next(layers_, parsedLayers)
			if err != nil {
				return err
			}

			return BlacklistLayersHandler(slugs)(layers_, parsedLayers)
		}
	}
}

// BlacklistLayerParameters is a middleware that removes the given parameters from ParameterLayers after running `next`.
func BlacklistLayerParameters(parameters_ map[string][]string) Middleware {
	return func(next HandlerFunc) HandlerFunc {
		return func(layers_ *schema.Schema, parsedLayers *values.Values) error {
			err := next(layers_, parsedLayers)
			if err != nil {
				return err
			}

			return BlacklistLayerParametersHandler(parameters_)(layers_, parsedLayers)
		}
	}
}

// BlacklistLayerParametersFirst is a middleware that removes the given parameters from ParameterLayers before running `next`.
func BlacklistLayerParametersFirst(parameters_ map[string][]string) Middleware {
	return func(next HandlerFunc) HandlerFunc {
		return func(layers_ *schema.Schema, parsedLayers *values.Values) error {
			err := BlacklistLayerParametersHandler(parameters_)(layers_, parsedLayers)
			if err != nil {
				return err
			}

			return next(layers_, parsedLayers)
		}
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
		return func(layers_ *schema.Schema, parsedLayers *values.Values) error {
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
	return WrapWithLayerModifyingHandler(WhitelistLayersHandler(slugs), nextMiddlewares...)
}

func WrapWithWhitelistedParameterLayers(parameters_ map[string][]string, nextMiddlewares ...Middleware) Middleware {
	return WrapWithLayerModifyingHandler(WhitelistLayerParametersHandler(parameters_), nextMiddlewares...)
}

// WrapWithBlacklistedLayers wraps a middleware that restricts layers
// to a specified set of slugs, with any additional middlewares.
// It makes it possible to apply a subset of middlewares to only
// certain restricted layers.
func WrapWithBlacklistedLayers(slugs []string, nextMiddlewares ...Middleware) Middleware {
	return WrapWithLayerModifyingHandler(BlacklistLayersHandler(slugs), nextMiddlewares...)
}

func WrapWithBlacklistedParameterLayers(parameters_ map[string][]string, nextMiddlewares ...Middleware) Middleware {
	return WrapWithLayerModifyingHandler(BlacklistLayerParametersHandler(parameters_), nextMiddlewares...)
}
