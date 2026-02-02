package schema

import "github.com/go-go-golems/glazed/pkg/cmds/fields"

type WhitelistParameterLayer struct {
	Section
	WhitelistedParameters map[string]interface{}
}

var _ Section = (*WhitelistParameterLayer)(nil)

func NewWhitelistParameterLayer(layer Section, whitelistedParameters map[string]interface{}) *WhitelistParameterLayer {
	return &WhitelistParameterLayer{
		Section:               layer,
		WhitelistedParameters: whitelistedParameters,
	}
}

func (l *WhitelistParameterLayer) GetDefinitions() *fields.Definitions {
	pds := l.Section.GetDefinitions()
	ret := fields.NewDefinitions()
	pds.ForEach(func(pd *fields.Definition) {
		if _, ok := l.WhitelistedParameters[pd.Name]; ok {
			ret.Set(pd.Name, pd)
		}
	})
	return ret
}

type BlacklistParameterLayer struct {
	Section
	BlacklistedParameters map[string]interface{}
}

var _ Section = (*BlacklistParameterLayer)(nil)

func NewBlacklistParameterLayer(layer Section, blacklistedParameters map[string]interface{}) *BlacklistParameterLayer {
	return &BlacklistParameterLayer{
		Section:               layer,
		BlacklistedParameters: blacklistedParameters,
	}
}

func (l *BlacklistParameterLayer) GetDefinitions() *fields.Definitions {
	pds := l.Section.GetDefinitions()
	ret := fields.NewDefinitions()
	pds.ForEach(func(pd *fields.Definition) {
		if _, ok := l.BlacklistedParameters[pd.Name]; !ok {
			ret.Set(pd.Name, pd)
		}
	})
	return ret
}
