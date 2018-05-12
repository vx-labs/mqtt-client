package main

import (
	"fmt"
	"strings"

	MQTT "github.com/eclipse/paho.mqtt.golang"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

func mqttSubscriber() *cobra.Command {
	done := make(chan error)
	c := &cobra.Command{
		Use: "subscribe",
		RunE: func(cmd *cobra.Command, args []string) error {
			topics := getStringArrayFlag(cmd, "topic")
			qos := getIntFlag(cmd, "qos")

			topicsMap := map[string]byte{}
			for _, topic := range topics {
				topicsMap[topic] = byte(qos)
			}
			fmt.Fprintf(cmd.OutOrStderr(), "%s subscribing to topics %s\n", color.GreenString(now()), color.CyanString(strings.Join(topics, ",")))
			_, err := client(func(c MQTT.Client) {
				if token := c.SubscribeMultiple(topicsMap, func(client MQTT.Client, msg MQTT.Message) {
					if msg.Retained() {
						fmt.Fprintf(cmd.OutOrStdout(), "%s %s → %s (retained)\n", color.GreenString(now()), color.CyanString(msg.Topic()), color.YellowString(string(msg.Payload())))
					} else {
						fmt.Fprintf(cmd.OutOrStdout(), "%s %s → %s\n", color.GreenString(now()), color.CyanString(msg.Topic()), string(msg.Payload()))
					}
				}); token.Wait() && token.Error() != nil {
					done <- token.Error()
				}
			}, connLostHandler(cmd))
			if err != nil {
				return fmt.Errorf("unable to connect to mqtt broker: %v", err)
			}
			select {
			case err := <-done:
				return err
			}
		},
	}
	c.Flags().StringArrayP("topic", "t", nil, "subscribe to these topics")
	c.Flags().IntP("qos", "q", 0, "set the subscription QoS policy")
	return c
}
