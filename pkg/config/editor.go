package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
)

type ConfigEditor struct {
	viper   *viper.Viper
	path    string
	appName string
}

// NewConfigEditor creates a new ConfigEditor with a specific path
func NewConfigEditor(path string) (*ConfigEditor, error) {
	log.Debug().Msgf("Creating config editor for path: %s", path)
	v := viper.New()
	v.SetConfigFile(path)

	// Create directory if it doesn't exist
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("could not create config directory: %w", err)
	}

	// Try to read config, but don't fail if it doesn't exist
	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("could not read config: %w", err)
		}
	}

	return &ConfigEditor{
		viper: v,
		path:  path,
	}, nil
}

// NewAppConfigEditor creates a new ConfigEditor for a specific application
func NewAppConfigEditor(appName string, configPath string) (*ConfigEditor, error) {
	log.Debug().Msgf("Creating config editor for app: %s with path: %s", appName, configPath)

	var path string
	if configPath != "" {
		path = configPath
	} else {
		defaultPath, err := GetDefaultConfigPath(appName)
		if err != nil {
			return nil, err
		}
		path = defaultPath
	}

	editor, err := NewConfigEditor(path)
	if err != nil {
		return nil, err
	}

	editor.appName = appName
	return editor, nil
}

func (c *ConfigEditor) Save() error {
	return c.viper.WriteConfig()
}

func (c *ConfigEditor) Set(key string, value interface{}) error {
	c.viper.Set(key, value)
	return nil
}

func (c *ConfigEditor) Get(key string) (interface{}, error) {
	if !c.viper.IsSet(key) {
		return nil, fmt.Errorf("key not found: %s", key)
	}
	return c.viper.Get(key), nil
}

func (c *ConfigEditor) Delete(key string) error {
	if !c.viper.IsSet(key) {
		return fmt.Errorf("key not found: %s", key)
	}
	c.viper.Set(key, nil)
	return nil
}

func (c *ConfigEditor) ListKeys() []string {
	var allKeys []string
	allKeys = append(allKeys, c.viper.AllKeys()...)
	return allKeys
}

func (c *ConfigEditor) GetAll() map[string]interface{} {
	return c.viper.AllSettings()
}

// GetDefaultConfigPath is updated to take an appName parameter
func GetDefaultConfigPath(appName string) (string, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("could not get config dir: %w", err)
	}

	return filepath.Join(configDir, appName, "config.yaml"), nil
}

// Helper function to format config values for display
func FormatValue(value interface{}) string {
	switch v := value.(type) {
	case []interface{}:
		var items []string
		for _, item := range v {
			items = append(items, fmt.Sprintf("%v", item))
		}
		return fmt.Sprintf("[%s]", strings.Join(items, ", "))
	case map[string]interface{}:
		var items []string
		for k, val := range v {
			items = append(items, fmt.Sprintf("%s: %v", k, val))
		}
		return fmt.Sprintf("{%s}", strings.Join(items, ", "))
	default:
		return fmt.Sprintf("%v", value)
	}
}
