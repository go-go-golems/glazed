package cmds

import (
	"context"
	"github.com/DataDog/datadog-api-client-go/v2/api/datadog"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var CmdContext context.Context

var RootCmd = &cobra.Command{
	Use:   "dd-cli",
	Short: "dd-cli is a small client for the DataDog API",
	PersistentPreRun: func(cmd *cobra.Command, args []string) {

		CmdContext = context.WithValue(
			context.Background(),
			datadog.ContextAPIKeys,
			map[string]datadog.APIKey{
				"apiKeyAuth": {
					Key: viper.GetString("API_KEY"),
				},
				"appKeyAuth": {
					Key: viper.GetString("APP_KEY"),
				},
			},
		)
	},
}

func init() {
	viper.SetEnvPrefix("DD")
	viper.SetConfigName("datadog")
	viper.AddConfigPath("$HOME/.config")

	RootCmd.PersistentFlags().String("api-key", "", "API key for the DataDog API")
	RootCmd.PersistentFlags().String("app-key", "", "Application key for the DataDog API")

	_ = viper.BindEnv("API_KEY")
	_ = viper.BindEnv("APP_KEY")
	_ = viper.BindPFlag("API_KEY", RootCmd.PersistentFlags().Lookup("api-key"))
	_ = viper.BindPFlag("APP_KEY", RootCmd.PersistentFlags().Lookup("app-key"))

	RootCmd.AddCommand(&RumCmd)

}
