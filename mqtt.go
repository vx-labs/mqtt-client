package main

import (
	"fmt"
	"time"

	MQTT "github.com/eclipse/paho.mqtt.golang"
	"github.com/fatih/color"
	"github.com/spf13/viper"
	"github.com/urfave/cli"
)

func client(d MQTT.OnConnectHandler, l MQTT.ConnectionLostHandler) (MQTT.Client, error) {
	opts := MQTT.NewClientOptions().AddBroker(viper.GetString("mqtt.broker"))
	opts.Username = viper.GetString("mqtt.username")
	opts.Password = viper.GetString("mqtt.password")
	opts.SetClientID(fmt.Sprintf("mqtt-cli-%s", time.Now().Format("15:04:05")))
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
