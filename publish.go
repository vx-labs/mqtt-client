package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"

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
			payload := []byte(getStringFlag(cmd, "message"))
			qos := getIntFlag(cmd, "qos")
			if len(payload) == 0 {
				filePath := getStringFlag(cmd, "message-from")
				if len(filePath) == 0 {
					payload = nil
				} else {
					var err error
					payload, err = ioutil.ReadFile(filePath)
					if err != nil {
						log.Printf("failed to read payload: %v", err)
						return err
					}
				}
			}
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
					done <- nil
				}
			}, connLostHandler(cmd))
			if err != nil {
				return fmt.Errorf("unable to connect to mqtt broker: %v", err)
			}
			err = <-done
			fmt.Fprintf(os.Stdout, "%s %s ← %d bytes\n", color.GreenString(now()), color.CyanString(topic), len(payload))
			c.Disconnect(1000)
			return err
		},
	}
	c.Flags().StringP("topic", "t", "", "publish the message to the given topic")
	c.Flags().StringP("message", "m", "", "set the message payload")
	c.Flags().StringP("message-from", "f", "", "set the message payload from a file")
	c.Flags().BoolP("retain", "r", false, "ask the broker to retain the message")
	c.Flags().IntP("qos", "q", 0, "set the message QoS policy")

	return c
}
