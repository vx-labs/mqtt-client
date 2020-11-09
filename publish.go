package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"github.com/dustin/go-humanize"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	MQTT "github.com/eclipse/paho.mqtt.golang"
)

type msgPublished struct {
	Timestamp   string
	Topic       string
	Payload     string
	PayloadSize string
}

const defaultPublishMsgFormat = `[{{ .Timestamp | yellow | faint }}] {{ .Topic | faint }} ← {{ .PayloadSize }}`

func mqttPublisher(config *viper.Viper) *cobra.Command {
	c := &cobra.Command{
		Use: "publish",
		Run: func(cmd *cobra.Command, args []string) {
			topic := getStringFlag(cmd, "topic")
			retain := getBoolFlag(cmd, "retain")
			payload := []byte(getStringFlag(cmd, "message"))
			qos := getIntFlag(cmd, "qos")
			userFormat := getStringFlag(cmd, "format")

			tpl, err := parseTemplate(userFormat, defaultPublishMsgFormat)
			if err != nil {
				log.Fatal(err)
			}

			if len(payload) == 0 {
				filePath := getStringFlag(cmd, "message-from")
				if len(filePath) == 0 {
					payload = nil
				} else {
					var err error
					payload, err = ioutil.ReadFile(filePath)
					if err != nil {
						fmt.Printf("failed to read payload: %v\n", err)
						return
					}
				}
			}
			if qos > 2 || qos < 0 {
				fmt.Printf("invalid qos provided\n")
				return
			}
			done := make(chan error)
			c, err := client(config, func(c MQTT.Client) {
				defer close(done)
				if token := c.Publish(topic, byte(qos), retain, payload); token.Wait() && token.Error() != nil {
					done <- fmt.Errorf("failed to publish to requested topic: %v", token.Error())
				} else {
					done <- nil
				}
			}, connLostHandler(cmd))
			if err != nil {
				fmt.Printf("failed to connect to mqtt broker: %v\n", err)
				return
			}
			err = <-done
			if err != nil {
				fmt.Fprintln(os.Stderr, err.Error())
				return
			}
			msg := msgPublished{
				Timestamp:   now(),
				Payload:     string(payload),
				Topic:       topic,
				PayloadSize: humanize.Bytes(uint64(len(payload))),
			}
			tpl.Execute(os.Stdout, msg)
			c.Disconnect(1000)
		},
	}
	c.Flags().StringP("topic", "t", "", "publish the message to the given topic")
	c.Flags().StringP("message", "m", "", "set the message payload")
	c.Flags().StringP("message-from", "f", "", "set the message payload from a file")
	c.Flags().BoolP("retain", "r", false, "ask the broker to retain the message")
	c.Flags().IntP("qos", "q", 0, "set the message QoS policy")
	c.Flags().StringP("format", "", "", "use this template to format output")

	return c
}
