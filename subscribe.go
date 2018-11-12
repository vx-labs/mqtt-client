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

func mqttSubscriber() *cobra.Command {
	done := make(chan error)
	var mqtt MQTT.Client
	c := &cobra.Command{
		Use: "subscribe",
		RunE: func(cmd *cobra.Command, args []string) error {
			topics := getStringArrayFlag(cmd, "topic")
			qos := getIntFlag(cmd, "qos")

			sigc := make(chan os.Signal)

			topicsMap := map[string]byte{}
			for _, topic := range topics {
				topicsMap[topic] = byte(qos)
			}
			spinner := newSpinner(cmd.OutOrStderr(), fmt.Sprintf("subscribing to topics %s", strings.Join(topics, ",")))
			var err error
			mqtt, err = client(func(c MQTT.Client) {
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
			spinner.Stop()
			if err != nil {
				return fmt.Errorf("unable to connect to mqtt broker: %v", err)
			}
			signal.Notify(sigc, syscall.SIGINT, syscall.SIGTERM)
			select {
			case err := <-done:
				return err
			case <-sigc:
				spinner = newSpinner(cmd.OutOrStderr(), "disconnecting from broker")
				spinner.FinalMSG = "disconnected from broker\n"
				mqtt.Disconnect(1000)
				spinner.Stop()
				return nil
			}
		},
	}
	c.Flags().StringArrayP("topic", "t", nil, "subscribe to these topics")
	c.Flags().IntP("qos", "q", 0, "set the subscription QoS policy")
	return c
}
