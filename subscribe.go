package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"text/template"

	MQTT "github.com/eclipse/paho.mqtt.golang"
	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type RawMessage struct {
	Timestamp string
	Topic     string
	Retained  bool
	Payload   string
	Qos       int32
	Parsed    map[string]interface{}
}

const defaultMsgFormat = `[{{ .Timestamp | yellow | faint }}] {{ .Topic | faint }}
{{- if .Retained}} {{ .Payload | yellow | faint }}{{ else }} {{ .Payload }}{{end}}`

func parseTemplate(user string, fallback string) (*template.Template, error) {
	body := fallback
	if user != "" {
		body = user
	}
	tpl := template.New("").Funcs(promptui.FuncMap)
	tpl, err := tpl.Parse(body + "\n")
	if err == nil {
		return tpl, nil
	}
	return nil, err
}

func mqttSubscriber(config *viper.Viper) *cobra.Command {
	done := make(chan error)
	var mqtt MQTT.Client
	c := &cobra.Command{
		Use: "subscribe",
		RunE: func(cmd *cobra.Command, args []string) error {
			topics := getStringArrayFlag(cmd, "topic")
			qos := getIntFlag(cmd, "qos")
			userFormat := getStringFlag(cmd, "format")

			sigc := make(chan os.Signal)
			tpl, err := parseTemplate(userFormat, defaultMsgFormat)
			if err != nil {
				log.Fatal(err)
			}
			topicsMap := map[string]byte{}
			for _, topic := range topics {
				topicsMap[topic] = byte(qos)
			}
			mqtt, err = client(config, func(c MQTT.Client) {
				if token := c.SubscribeMultiple(topicsMap, func(client MQTT.Client, msg MQTT.Message) {
					data := RawMessage{
						Parsed:    nil,
						Payload:   string(msg.Payload()),
						Qos:       int32(msg.Qos()),
						Retained:  msg.Retained(),
						Timestamp: now(),
						Topic:     msg.Topic(),
					}
					json.Unmarshal(msg.Payload(), &data.Parsed)
					tpl.Execute(os.Stdout, data)
				}); token.Wait() && token.Error() != nil {
					done <- token.Error()
				}
			}, connLostHandler(cmd))
			if err != nil {
				return fmt.Errorf("unable to connect to mqtt broker: %v", err)
			}
			signal.Notify(sigc, syscall.SIGINT, syscall.SIGTERM)
			select {
			case err := <-done:
				return err
			case <-sigc:
				fmt.Print("\n")
				mqtt.Disconnect(1000)
				return nil
			}
		},
	}
	c.Flags().StringArrayP("topic", "t", nil, "subscribe to these topics")
	c.Flags().IntP("qos", "q", 0, "set the subscription QoS policy")
	c.Flags().StringP("format", "", "", "use this template to format messages")
	c.Flags().BoolP("raw", "", false, "do not display any spinner")
	return c
}
