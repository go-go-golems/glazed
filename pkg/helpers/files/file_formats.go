package files

import (
	"encoding/json"
	"gopkg.in/yaml.v3"
)

func ConvertYAMLMapToJSON(yamlContent string) (string, error) {
	// parse YAML into a map
	source := make(map[string]interface{})
	err := yaml.Unmarshal([]byte(yamlContent), &source)
	if err != nil {
		return "", err
	}

	// convert the map to JSON
	jsonContent, err := json.MarshalIndent(source, "", "  ")
	if err != nil {
		return "", err
	}

	return string(jsonContent), nil
}

func ConvertYAMLArrayToJSON(yamlContent string) (string, error) {
	// parse YAML into a map
	source := make([]interface{}, 0)
	err := yaml.Unmarshal([]byte(yamlContent), &source)
	if err != nil {
		return "", err
	}

	// convert the map to JSON
	jsonContent, err := json.MarshalIndent(source, "", "  ")
	if err != nil {
		return "", err
	}

	return string(jsonContent), nil
}

func ConvertJSONMapToYAML(jsonContent string) (string, error) {
	// parse JSON into a map
	source := make(map[string]interface{})
	err := json.Unmarshal([]byte(jsonContent), &source)
	if err != nil {
		return "", err
	}

	// convert the map to YAML
	yamlContent, err := yaml.Marshal(source)
	if err != nil {
		return "", err
	}

	return string(yamlContent), nil
}

func ConvertJSONArrayToYAML(jsonContent string) (string, error) {
	// parse JSON into a map
	source := make([]interface{}, 0)
	err := json.Unmarshal([]byte(jsonContent), &source)
	if err != nil {
		return "", err
	}

	// convert the map to YAML
	yamlContent, err := yaml.Marshal(source)
	if err != nil {
		return "", err
	}

	return string(yamlContent), nil
}
