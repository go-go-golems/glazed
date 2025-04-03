package config

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type ConfigCommand struct {
	*cobra.Command
	appName string
}

func NewConfigCommand(appName string) (*ConfigCommand, error) {
	cmd := &ConfigCommand{
		appName: appName,
	}

	cobraCmd := &cobra.Command{
		Use:   "config",
		Short: "Commands for manipulating configuration",
	}

	cobraCmd.AddCommand(cmd.newListCommand())
	cobraCmd.AddCommand(cmd.newGetCommand())
	cobraCmd.AddCommand(cmd.newSetCommand())
	cobraCmd.AddCommand(cmd.newDeleteCommand())
	cobraCmd.AddCommand(cmd.newEditCommand())
	cobraCmd.AddCommand(cmd.newShowCommand())
	cobraCmd.AddCommand(cmd.newPathCommand())

	cmd.Command = cobraCmd
	return cmd, nil
}

// getEditor returns a new ConfigEditor instance for the current config file
func (c *ConfigCommand) getEditor() (*ConfigEditor, error) {
	configPath := viper.ConfigFileUsed()
	log.Debug().Str("config_path", configPath).Msg("using config file")

	// If no config file is used, get the default path
	if configPath == "" {
		var err error
		configPath, err = GetDefaultConfigPath(c.appName)
		if err != nil {
			return nil, err
		}
	}

	editor, err := NewConfigEditor(configPath)
	if err != nil {
		return nil, fmt.Errorf("could not create config editor: %w", err)
	}

	return editor, nil
}

func (c *ConfigCommand) newListCommand() *cobra.Command {
	var concise bool
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List all configuration keys and values",
		RunE: func(cmd *cobra.Command, args []string) error {
			editor, err := c.getEditor()
			if err != nil {
				return err
			}

			if concise {
				keys := editor.ListKeys()
				for _, key := range keys {
					fmt.Println(key)
				}
				return nil
			}

			settings := editor.GetAll()
			for key, value := range settings {
				fmt.Printf("%s: %s\n", key, FormatValue(value))
			}
			return nil
		},
	}

	cmd.Flags().BoolVarP(&concise, "concise", "c", true, "Only show keys")
	return cmd
}

func (c *ConfigCommand) newGetCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "get <key>",
		Short: "Get a configuration value",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			editor, err := c.getEditor()
			if err != nil {
				return err
			}

			key := args[0]
			value, err := editor.Get(key)
			if err != nil {
				return err
			}

			fmt.Printf("%s\n", FormatValue(value))
			return nil
		},
	}
}

func (c *ConfigCommand) newSetCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "set <key> <value>",
		Short: "Set a configuration value",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			editor, err := c.getEditor()
			if err != nil {
				return err
			}

			key := args[0]
			value := args[1]

			if err := editor.Set(key, value); err != nil {
				return err
			}

			return editor.Save()
		},
	}
}

func (c *ConfigCommand) newDeleteCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "delete <key>",
		Short: "Delete a configuration value",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			editor, err := c.getEditor()
			if err != nil {
				return err
			}

			key := args[0]

			if err := editor.Delete(key); err != nil {
				return err
			}

			return editor.Save()
		},
	}
}

func (c *ConfigCommand) newEditCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "edit",
		Short: "Edit the configuration file in your default editor",
		RunE: func(cmd *cobra.Command, args []string) error {
			editor := os.Getenv("EDITOR")
			if editor == "" {
				editor = "vim"
			}

			configPath := viper.ConfigFileUsed()
			if configPath == "" {
				var err error
				configPath, err = GetDefaultConfigPath(c.appName)
				if err != nil {
					return err
				}
			}

			// Ensure the directory exists
			if err := os.MkdirAll(filepath.Dir(configPath), 0755); err != nil {
				return fmt.Errorf("could not create config directory: %w", err)
			}

			// Validate editor to prevent command injection
			// Common editors are allowed, otherwise fallback to a safe default
			validEditors := map[string]bool{
				"vim":        true,
				"vi":         true,
				"nano":       true,
				"emacs":      true,
				"code":       true,
				"subl":       true,
				"gedit":      true,
				"notepad":    true,
				"notepad++":  true,
				"atom":       true,
				"sublime":    true,
				"vscode":     true,
				"textmate":   true,
				"neovim":     true,
				"nvim":       true,
				"micro":      true,
				"kwrite":     true,
				"kate":       true,
				"mousepad":   true,
				"leafpad":    true,
				"gvim":       true,
				"pluma":      true,
				"xed":        true,
				"jedit":      true,
				"codeblocks": true,
			}

			// Extract the base editor command without arguments
			editorParts := strings.Fields(editor)
			baseEditor := filepath.Base(editorParts[0])

			if !validEditors[baseEditor] {
				return fmt.Errorf("editor '%s' not in allowed list, please use a standard editor", baseEditor)
			}

			editCmd := exec.Command(editor, configPath)
			editCmd.Stdin = os.Stdin
			editCmd.Stdout = os.Stdout
			editCmd.Stderr = os.Stderr

			return editCmd.Run()
		},
	}
}

func (c *ConfigCommand) newShowCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "show",
		Short: "Show the entire configuration file contents",
		RunE: func(cmd *cobra.Command, args []string) error {
			configPath := viper.ConfigFileUsed()
			if configPath == "" {
				var err error
				configPath, err = GetDefaultConfigPath(c.appName)
				if err != nil {
					return err
				}
			}

			// Check if file exists
			_, err := os.Stat(configPath)
			if os.IsNotExist(err) {
				fmt.Printf("No configuration file found at %s\n", configPath)
				return nil
			} else if err != nil {
				return fmt.Errorf("error checking config file: %w", err)
			}

			// Read and print the file contents
			content, err := os.ReadFile(configPath)
			if err != nil {
				return fmt.Errorf("error reading config file: %w", err)
			}

			fmt.Printf("Configuration file at %s:\n\n%s", configPath, string(content))
			return nil
		},
	}
}

func (c *ConfigCommand) newPathCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "path",
		Short: "Show or set the configuration file path",
		RunE: func(cmd *cobra.Command, args []string) error {
			// Just show the current path
			configPath := viper.ConfigFileUsed()
			if configPath == "" {
				defaultPath, err := GetDefaultConfigPath(c.appName)
				if err != nil {
					return err
				}
				configPath = defaultPath
			}

			fmt.Println(configPath)
			return nil
		},
	}

	return cmd
}
