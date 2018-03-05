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
	"strings"
	"syscall"
	"os/signal"
)

func client(d MQTT.OnConnectHandler, l MQTT.ConnectionLostHandler) (MQTT.Client, error) {
	opts := MQTT.NewClientOptions().AddBroker(viper.GetString("mqtt.broker"))
	opts.Username = viper.GetString("mqtt.username")
	opts.Password = viper.GetString("mqtt.password")
	opts.SetClientID(fmt.Sprintf("mc-%s", time.Now().Format("15:04:05")))
	opts.OnConnect = d
	opts.OnConnectionLost = l
	c := MQTT.NewClient(opts)
	if token := c.Connect(); token.Wait() && token.Error() != nil {
		return nil, token.Error()
	}
	return c, nil
}

func now() string {
	return time.Now().Format(time.Stamp)
}

func connLostHandler(ctx *cli.Context) MQTT.ConnectionLostHandler {
	return func(client MQTT.Client, e error) {
		fmt.Fprintf(ctx.App.Writer, "%s connection to broker lost - reconnecting..\n", color.GreenString(now()))
	}
}

func mqttSubscriber() cli.Command {
	var cliTopics cli.StringSlice
	var cliQos int
	done := make(chan error)
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
			if len(cliTopics) == 0 {
				return fmt.Errorf("no topics were selected")
			}

			sigc := make(chan os.Signal, 1)
			signal.Notify(sigc,
				syscall.SIGHUP,
				syscall.SIGINT,
				syscall.SIGTERM,
				syscall.SIGQUIT)

			topics := map[string]byte{}
			for _, topic := range cliTopics {
				topics[topic] = byte(cliQos)
			}
			fmt.Fprintf(ctx.App.Writer, "%s subscribing to topics %s\n", color.GreenString(now()), color.CyanString(strings.Join(cliTopics, ",")))
			c, err := client(func(c MQTT.Client) {
				if token := c.SubscribeMultiple(topics, func(client MQTT.Client, msg MQTT.Message) {
					if msg.Retained() {
						fmt.Fprintf(ctx.App.Writer, "%s %s → %s (retained)\n", color.GreenString(now()), color.CyanString(msg.Topic()), color.YellowString(string(msg.Payload())))
					} else {
						fmt.Fprintf(ctx.App.Writer, "%s %s → %s\n", color.GreenString(now()), color.CyanString(msg.Topic()), string(msg.Payload()))
					}
				}); token.Wait() && token.Error() != nil {
					done <- token.Error()
				}
			}, connLostHandler(ctx))
			if err != nil {
				return fmt.Errorf("unable to connect to mqtt broker: %v\n", err)
			}
			select {
			case err := <-done:
				return err
			case <-sigc:
				fmt.Fprintf(ctx.App.Writer, "%s disconnecting from broker\n", color.GreenString(now()))
				c.Disconnect(250)
				return nil
			}
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
			if len(cliTopic) == 0 {
				return fmt.Errorf("no topic were selected")
			}
			done := make(chan error)
			c, err := client(func(c MQTT.Client) {
				defer close(done)
				if token := c.Publish(cliTopic, byte(cliQos), retained, payload); token.Wait() && token.Error() != nil {
					done <- fmt.Errorf("unable to publish to requested topic: %v\n", token.Error())
				} else {
					fmt.Fprintf(ctx.App.Writer, "%s %s ← %s\n", color.GreenString(now()), color.CyanString(cliTopic), string(payload))
					done <- nil
				}
			}, connLostHandler(ctx))
			if err != nil {
				return fmt.Errorf("unable to connect to mqtt broker: %v\n", err)
			}
			err = <-done
			c.Disconnect(250)
			return err
		},
	}
}

func main() {

	viper.SetConfigName("config")             // name of config file (without extension)
	viper.AddConfigPath("$HOME/.mqtt-client") // call multiple times to add many search paths
	viper.AddConfigPath(".")                  // optionally look for config in the working directory
	viper.SetDefault("mqtt.broker", "tcp://localhost:1883")
	err := viper.ReadInConfig() // Find and read the config file
	if err != nil { // Handle errors reading the config file
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
