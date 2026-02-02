package layers

import "github.com/go-go-golems/glazed/pkg/cmds/fields"

type WhitelistParameterLayer struct {
	ParameterLayer
	WhitelistedParameters map[string]interface{}
}

var _ ParameterLayer = (*WhitelistParameterLayer)(nil)

func NewWhitelistParameterLayer(layer ParameterLayer, whitelistedParameters map[string]interface{}) *WhitelistParameterLayer {
	return &WhitelistParameterLayer{
		ParameterLayer:        layer,
		WhitelistedParameters: whitelistedParameters,
	}
}

func (l *WhitelistParameterLayer) GetParameterDefinitions() *fields.Definitions {
	pds := l.ParameterLayer.GetParameterDefinitions()
	ret := fields.NewDefinitions()
	pds.ForEach(func(pd *fields.Definition) {
		if _, ok := l.WhitelistedParameters[pd.Name]; ok {
			ret.Set(pd.Name, pd)
		}
	})
	return ret
}

type BlacklistParameterLayer struct {
	ParameterLayer
	BlacklistedParameters map[string]interface{}
}

var _ ParameterLayer = (*BlacklistParameterLayer)(nil)

func NewBlacklistParameterLayer(layer ParameterLayer, blacklistedParameters map[string]interface{}) *BlacklistParameterLayer {
	return &BlacklistParameterLayer{
		ParameterLayer:        layer,
		BlacklistedParameters: blacklistedParameters,
	}
}

func (l *BlacklistParameterLayer) GetParameterDefinitions() *fields.Definitions {
	pds := l.ParameterLayer.GetParameterDefinitions()
	ret := fields.NewDefinitions()
	pds.ForEach(func(pd *fields.Definition) {
		if _, ok := l.BlacklistedParameters[pd.Name]; !ok {
			ret.Set(pd.Name, pd)
		}
	})
	return ret
}
