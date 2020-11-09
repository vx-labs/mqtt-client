package main

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func main() {
	config := viper.New()
	config.SetEnvPrefix("MQTT_CLIENT")
	config.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))
	config.AutomaticEnv()

	app := cobra.Command{
		Use:   "mqtt",
		Short: "command-line mqtt client",
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			config.AutomaticEnv()
			config.SetConfigName("config")
			config.AddConfigPath("$HOME/.mqtt-client")
			config.AddConfigPath("$HOME/.config/mqtt-client")
			config.AddConfigPath(".")
			config.SetDefault("mqtt.broker", "tcp://localhost:1883")
			config.BindPFlags(cmd.Flags())

			err := config.ReadInConfig()
			if err != nil {
				fmt.Fprintf(cmd.OutOrStderr(), "failed to load configuration: %s", err)
				fmt.Fprintf(cmd.OutOrStderr(), "default configuration values will be used")
			}
			return nil
		},
	}
	app.AddCommand(mqttPublisher(config))
	app.AddCommand(mqttSubscriber(config))
	app.Execute()
}
