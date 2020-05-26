package main

import (
	"crypto/tls"
	"fmt"
	"net"
	"net/url"
	"time"

	"github.com/spf13/cobra"

	MQTT "github.com/eclipse/paho.mqtt.golang"
	"github.com/fatih/color"
	"github.com/spf13/viper"
)

func client(d MQTT.OnConnectHandler, l MQTT.ConnectionLostHandler) (MQTT.Client, error) {
	broker := viper.GetString("mqtt.broker")
	opts := MQTT.NewClientOptions().AddBroker(broker)
	opts.Username = viper.GetString("mqtt.username")
	opts.Password = viper.GetString("mqtt.password")
	opts.SetClientID(fmt.Sprintf("mqtt-cli"))
	opts.OnConnect = d
	opts.OnConnectionLost = l
	brokerURL, err := url.Parse(broker)
	if err != nil {
		panic(err)
	}
	if brokerURL.Scheme == "tls" {
		host, _, _ := net.SplitHostPort(brokerURL.Host)
		opts.TLSConfig = &tls.Config{
			MinVersion:               tls.VersionTLS12,
			CurvePreferences:         []tls.CurveID{tls.CurveP521, tls.CurveP384, tls.CurveP256},
			PreferServerCipherSuites: true,
			CipherSuites: []uint16{
				tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
				tls.TLS_RSA_WITH_AES_256_GCM_SHA384,
			},
			ServerName: host,
		}
	}
	c := MQTT.NewClient(opts)
	if token := c.Connect(); token.Wait() && token.Error() != nil {
		return nil, token.Error()
	}
	return c, nil
}

func now() string {
	return time.Now().Format(time.Stamp)
}

func connLostHandler(app *cobra.Command) MQTT.ConnectionLostHandler {
	return func(client MQTT.Client, e error) {
		fmt.Fprintf(app.OutOrStderr(), "%s connection to broker lost (%v) - reconnecting..\n", e, color.GreenString(now()))
	}
}
