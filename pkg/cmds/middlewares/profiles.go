package middlewares

import (
	"github.com/go-go-golems/glazed/pkg/cmds/layers"
	"github.com/go-go-golems/glazed/pkg/cmds/parameters"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
	"os"
)

func GatherFlagsFromProfiles(
	defaultProfileFile string,
	profileFile string,
	profile string,
	options ...parameters.ParseStepOption) Middleware {
	return func(next HandlerFunc) HandlerFunc {
		return func(layers_ *layers.ParameterLayers, parsedLayers *layers.ParsedLayers) error {
			err := next(layers_, parsedLayers)
			if err != nil {
				return err
			}

			// if the file does not exist and is not the defaultProfileFile, fail
			_, err = os.Stat(profileFile)
			if os.IsNotExist(err) {
				if profileFile != defaultProfileFile {
					return errors.Errorf("profile file %s does not exist", profileFile)
				}
				return nil
			}

			// parse profileFile as yaml
			f, err := os.Open(profileFile)
			if err != nil {
				return err
			}
			defer func(f *os.File) {
				_ = f.Close()
			}(f)

			// profile1:
			//   layer1:
			//     parameterName: parameterValue
			//   layer2:
			//     parameterName: parameterValue
			// etc...
			v := map[string]map[string]map[string]interface{}{}
			decoder := yaml.NewDecoder(f)
			err = decoder.Decode(&v)
			if err != nil {
				return err
			}

			if profileMap, ok := v[profile]; ok {
				return updateFromMap(layers_, parsedLayers, profileMap, options...)
			} else {
				if profile != "default" {
					return errors.Errorf("profile %s not found in %s", profile, profileFile)
				}

				return nil
			}
		}
	}
}
