package files

import (
	"encoding/json"
	"gopkg.in/yaml.v3"
	"os"
)

func LoadJSONFile(path string, target interface{}) error {
	b, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	return json.Unmarshal(b, target)
}

func LoadYAMLFile(path string, target interface{}) error {
	b, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	return yaml.Unmarshal(b, target)
}
