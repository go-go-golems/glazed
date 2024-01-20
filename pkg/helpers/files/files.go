package files

import (
	"encoding/json"
	"gopkg.in/yaml.v3"
	"io"
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

// ConcatFiles creates an io.Reader that reads from the provided files in order.
func ConcatFiles(filePaths ...string) (io.Reader, error) {
	var readers []io.Reader

	// Open each file and create a reader for it.
	for _, path := range filePaths {
		file, err := os.Open(path)
		if err != nil {
			// Close all opened files upon an error.
			for _, r := range readers {
				if f, ok := r.(*os.File); ok {
					_ = f.Close()
				}
			}
			return nil, err
		}
		readers = append(readers, file)
	}

	// Combine all readers into a single MultiReader.
	return io.MultiReader(readers...), nil
}
