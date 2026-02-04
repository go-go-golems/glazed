package schema

import "github.com/go-go-golems/glazed/pkg/cmds/fields"

type WhitelistSection struct {
	Section
	WhitelistedFields map[string]interface{}
}

var _ Section = (*WhitelistSection)(nil)

func NewWhitelistSection(section Section, whitelistedFields map[string]interface{}) *WhitelistSection {
	return &WhitelistSection{
		Section:           section,
		WhitelistedFields: whitelistedFields,
	}
}

func (l *WhitelistSection) GetDefinitions() *fields.Definitions {
	pds := l.Section.GetDefinitions()
	ret := fields.NewDefinitions()
	pds.ForEach(func(pd *fields.Definition) {
		if _, ok := l.WhitelistedFields[pd.Name]; ok {
			ret.Set(pd.Name, pd)
		}
	})
	return ret
}

type BlacklistSection struct {
	Section
	BlacklistedFields map[string]interface{}
}

var _ Section = (*BlacklistSection)(nil)

func NewBlacklistSection(section Section, blacklistedFields map[string]interface{}) *BlacklistSection {
	return &BlacklistSection{
		Section:           section,
		BlacklistedFields: blacklistedFields,
	}
}

func (l *BlacklistSection) GetDefinitions() *fields.Definitions {
	pds := l.Section.GetDefinitions()
	ret := fields.NewDefinitions()
	pds.ForEach(func(pd *fields.Definition) {
		if _, ok := l.BlacklistedFields[pd.Name]; !ok {
			ret.Set(pd.Name, pd)
		}
	})
	return ret
}
