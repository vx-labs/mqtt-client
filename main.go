package main

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func main() {

	app := cobra.Command{
		Use:   "mqtt",
		Short: "command-line mqtt client",
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			viper.AutomaticEnv()
			viper.SetConfigName("config")
			viper.AddConfigPath("$HOME/.mqtt-client")
			viper.AddConfigPath("$HOME/.config/mqtt-client")
			viper.AddConfigPath(".")
			viper.SetDefault("mqtt.broker", "tcp://localhost:1883")
			err := viper.ReadInConfig()
			if err != nil {
				fmt.Fprintf(cmd.OutOrStderr(), "error while loading configuration: %s", err)
				fmt.Fprintf(cmd.OutOrStderr(), "default configuration values will be used")
			}
			return nil
		},
	}
	app.AddCommand(mqttPublisher())
	app.AddCommand(mqttSubscriber())
	app.Execute()
}
