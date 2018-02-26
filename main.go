package main

import (
	"fmt"
	MQTT "github.com/eclipse/paho.mqtt.golang"
	"os"
	"github.com/urfave/cli"
	"log"
	"github.com/fatih/color"
	"time"
	"github.com/spf13/viper"
)

func client() (MQTT.Client, error) {
	opts := MQTT.NewClientOptions().AddBroker(viper.GetString("mqtt.broker"))
	opts.Username = viper.GetString("mqtt.username")
	opts.Password = viper.GetString("mqtt.password")
	opts.SetClientID(fmt.Sprintf("mqtt-client-%s",time.Now().Format(time.Stamp)))

	c := MQTT.NewClient(opts)
	if token := c.Connect(); token.Wait() && token.Error() != nil {
		return nil, token.Error()
	}
	return c, nil
}

func now() string {
	return time.Now().Format(time.Stamp)
}

func mqttSubscriber() cli.Command {
	var cliTopics cli.StringSlice
	var cliQos int
	return cli.Command{
		Name: "subscribe",
		Flags: []cli.Flag{
			cli.StringSliceFlag{
				Name:  "topic, t",
				Value: &cliTopics,
			},
			cli.IntFlag{
				Name:        "qos, q",
				Destination: &cliQos,
			},
		},
		Action: func(ctx *cli.Context) error {
			if cliQos > 2 || cliQos < 0 {
				return fmt.Errorf("invalid qos provided")
			}
			c, err := client()
			if err != nil {
				return fmt.Errorf("unable to connect to mqtt broker: %v\n", err)
			}
			defer c.Disconnect(250)
			topics := map[string]byte{}
			for _, topic := range cliTopics {
				topics[topic] = byte(cliQos)
			}
			if token := c.SubscribeMultiple(topics, func(client MQTT.Client, msg MQTT.Message) {
				fmt.Fprintf(ctx.App.Writer, "%s %s → %s\n", color.GreenString(now()), color.CyanString(msg.Topic()), string(msg.Payload()))
			}); token.Wait() && token.Error() != nil {
				return fmt.Errorf("unable to subscribe to requested topics: %v\n", token.Error())
			}
			fmt.Fprintf(ctx.App.Writer, "%s subscribed\n", color.GreenString(now()))
			<-make(chan struct{})
			return nil
		},
	}
}

func mqttPublisher() cli.Command {
	var cliTopic string
	var payload string
	var cliQos int
	var retained bool
	return cli.Command{
		Name: "publish",
		Flags: []cli.Flag{
			cli.StringFlag{
				Name:        "topic, t",
				Destination: &cliTopic,
			},
			cli.IntFlag{
				Name:        "qos, q",
				Destination: &cliQos,
			},
			cli.StringFlag{
				Name:        "message, m",
				Destination: &payload,
			},
			cli.BoolFlag{
				Name:        "retained, r",
				Destination: &retained,
			},
		},
		Action: func(ctx *cli.Context) error {
			if cliQos > 2 || cliQos < 0 {
				return fmt.Errorf("invalid qos provided")
			}
			c, err := client()
			if err != nil {
				return fmt.Errorf("unable to connect to mqtt broker: %v\n", err)
			}
			defer c.Disconnect(250)

			if token := c.Publish(cliTopic, byte(cliQos), retained, payload); token.Wait() && token.Error() != nil {
				return fmt.Errorf("unable to publish to requested topic: %v\n", token.Error())
			}
			fmt.Fprintf(ctx.App.Writer, "%s %s ← %s\n", color.GreenString(now()), color.CyanString(cliTopic), string(payload))
			return nil
		},
	}
}

func main() {

	viper.SetConfigName("config") // name of config file (without extension)
	viper.AddConfigPath("$HOME/.mqtt-client")  // call multiple times to add many search paths
	viper.AddConfigPath(".")               // optionally look for config in the working directory
	err := viper.ReadInConfig() // Find and read the config file
	if err != nil { // Handle errors reading the config file
		panic(fmt.Errorf("Fatal error config file: %s \n", err))
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
