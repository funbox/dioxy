package daemon

import (
	"fmt"
	"net/url"
	"path/filepath"
	"pkg.re/essentialkaos/ek.v10/log"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	store "github.com/gongled/dioxy/store"
)

// ////////////////////////////////////////////////////////////////////////////////// //

var datastore *store.Store

// ////////////////////////////////////////////////////////////////////////////////// //

// startObserver starts MQTT listener and updates measurements in memory
func startObserver(ip, port, user, password, topic string, ttl int) error {
	maxStoreTime := time.Duration(ttl) * time.Second
	datastore = store.NewStore(maxStoreTime)

	addr := fmt.Sprintf("tcp://%s:%s@%s:%s/%s", user, password, ip, port, topic)
	uri, err := url.Parse(addr)

	if err != nil {
		log.Crit(err.Error())
		shutdown(1)
	}

	log.Info("Broker listener is connecting to %s:%s", ip, port)
	go listenMQTT(uri, topic)

	return nil
}

// parseMQTTMessage parses MQTT message to Info struct
func parseMQTTMessage(msg mqtt.Message) *store.Info {
	prefix := filepath.Dir(msg.Topic())
	deviceId := filepath.Base(prefix)
	metrics := filepath.Base(msg.Topic())

	return &store.Info{
		Name:     fmt.Sprintf("%s,%s", deviceId, metrics),
		DeviceId: deviceId,
		Metrics:  metrics,
		Value:    string(msg.Payload()),
	}
}

// setMQTTOptions sets up MQTT client to connect as subscriber
func setMQTTOptions(clientId string, uri *url.URL) *mqtt.ClientOptions {
	opts := mqtt.NewClientOptions()
	password, _ := uri.User.Password()

	opts.AddBroker(fmt.Sprintf("tcp://%s", uri.Host))
	opts.SetUsername(uri.User.Username())
	opts.SetPassword(password)
	opts.SetClientID(clientId)

	return opts
}

// connectMQTT connects to MQTT broker to operate with metrics
func connectMQTT(clientId string, uri *url.URL) mqtt.Client {
	opts := setMQTTOptions(clientId, uri)

	client := mqtt.NewClient(opts)
	token := client.Connect()

	for !token.WaitTimeout(3 * time.Second) {
	}

	if err := token.Error(); err != nil {
		log.Crit("Error while connecting to MQTT broker due to: %s", err.Error())
		shutdown(1)
	}

	return client
}

// listenMQTT listens MQTT broker and update store in memory with latest data
func listenMQTT(uri *url.URL, topic string) {
	client := connectMQTT("sub", uri)
	token := client.Subscribe(topic, 0, func(client mqtt.Client, msg mqtt.Message) {
		datastore.Add(parseMQTTMessage(msg))
	})

	for token.Wait() {
	}

	if err := token.Error(); err != nil {
		log.Crit("Error while operating with MQTT broker due to: %s", err.Error())
		shutdown(1)
	}
}
