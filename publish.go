package main

import (
	"fmt"

	"github.com/spf13/cobra"

	MQTT "github.com/eclipse/paho.mqtt.golang"
	"github.com/fatih/color"
)

func mqttPublisher() *cobra.Command {
	c := &cobra.Command{
		Use: "publish",
		RunE: func(cmd *cobra.Command, args []string) error {
			topic := getStringFlag(cmd, "topic")
			retain := getBoolFlag(cmd, "retain")
			payload := getStringFlag(cmd, "message")
			qos := getIntFlag(cmd, "qos")
			if qos > 2 || qos < 0 {
				return fmt.Errorf("invalid qos provided")
			}
			if len(topic) == 0 {
				return fmt.Errorf("no topic were selected")
			}
			done := make(chan error)
			c, err := client(func(c MQTT.Client) {
				defer close(done)
				if token := c.Publish(topic, byte(qos), retain, payload); token.Wait() && token.Error() != nil {
					done <- fmt.Errorf("unable to publish to requested topic: %v", token.Error())
				} else {
					fmt.Fprintf(cmd.OutOrStdout(), "%s %s ← %s\n", color.GreenString(now()), color.CyanString(topic), payload)
					done <- nil
				}
			}, connLostHandler(cmd))
			if err != nil {
				return fmt.Errorf("unable to connect to mqtt broker: %v", err)
			}
			err = <-done
			c.Disconnect(250)
			return err
		},
	}
	c.Flags().StringP("topic", "t", "", "publish the message to the given topic")
	c.Flags().StringP("message", "m", "", "set the message payload")
	c.Flags().BoolP("retain", "r", false, "ask the broker to retain the message")
	c.Flags().IntP("qos", "q", 0, "set the message QoS policy")

	return c
}
