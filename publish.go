package main

import (
	"fmt"

	MQTT "github.com/eclipse/paho.mqtt.golang"
	"github.com/fatih/color"
	"github.com/urfave/cli"
)

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
