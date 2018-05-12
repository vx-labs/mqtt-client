package main

import (
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"

	MQTT "github.com/eclipse/paho.mqtt.golang"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

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
