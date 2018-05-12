package main

import (
	"fmt"
	"log"
	"os"

	"github.com/spf13/viper"
	"github.com/urfave/cli"
)

func main() {
	viper.AutomaticEnv()
	viper.SetConfigName("config")
	viper.AddConfigPath("$HOME/.mqtt-client")
	viper.AddConfigPath(".")
	viper.SetDefault("mqtt.broker", "tcp://localhost:1883")
	err := viper.ReadInConfig()
	if err != nil {
		fmt.Printf("error while loading configuration: %s \n", err)
		fmt.Print("default configuration values will be used\n")
	}
	app := cli.NewApp()
	app.Name = "mqtt-client"
	app.Usage = "command-line mqtt client"
	app.Commands = []cli.Command{
		mqttSubscriber(),
		mqttPublisher(),
	}
	err = app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
