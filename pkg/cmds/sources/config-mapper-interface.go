package sources

// ConfigMapper is an interface that can map raw config data to layer maps.
// This allows both ConfigFileMapper (function) and pattern-based mappers to be used interchangeably.
type ConfigMapper interface {
	Map(rawConfig interface{}) (map[string]map[string]interface{}, error)
}

// Make ConfigFileMapper satisfy ConfigMapper directly.
func (fn ConfigFileMapper) Map(rawConfig interface{}) (map[string]map[string]interface{}, error) {
	return fn(rawConfig)
}
