package zigbee

import (
	"encoding/json"
	"fmt"
	"iot/net"
	"log"
	"reflect"
	"strings"
	"sync"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

const (
	topicBase           string = "gw"
	topicCommands              = "commands"
	topicPublishState          = "publishstate"
	topicUpdateSettings        = "updatesettings"
)

var emptyMap = map[string]interface{}{}

type Gateway struct {
	client  mqtt.Client
	eui64   EUI64
	network *net.Network
	devices *sync.Map
}

func NewGateway() *Gateway {
	gw := &Gateway{
		devices: &sync.Map{},
	}

	return gw
}

func (g *Gateway) onHeartbeat(eui64 EUI64, message heartbeatMessage) {
	if net.EUI64(g.eui64) == net.NullEUI64 {
		fmt.Printf("found gateway: %s\n", eui64)
		g.eui64 = eui64

		if err := g.publish(topicPublishState, emptyMap); err != nil {
			log.Println(err)
		}

		if !message.NetworkUp {
			if err := g.publish("commands", commandMessage{Commands: []command{
				command{"plugin network-creator start 1", 5000},
			}}); err != nil {
				log.Println(err)
			}
		}

		if err := g.publish("commands", commandMessage{Commands: []command{
			command{"plugin network-creator-security open-network", 0},
		}}); err != nil {
			log.Println(err)
		}
	}
}

func (g *Gateway) onDeviceList(eui64 EUI64, message deviceListMessage) {
	for _, deviceMessage := range message.Devices {
		if d, ok := g.devices.Load(deviceMessage.Endpoint.EUI64); !ok {
			device := NewDevice(g, &deviceMessage)
			g.devices.Store(deviceMessage.Endpoint.EUI64, device)
			g.network.AddDevice(device)
		} else {
			device := d.(*device)
			device.merge(&deviceMessage)
			g.network.UpdateDevice(device)
		}
	}
}

func (g *Gateway) onRelayList(eui64 EUI64, message interface{}) {
	log.Println("onRelayList: ", message)
}

func (g *Gateway) onSettings(eui64 EUI64, message settingsMessage) {
	log.Println("onSettings: ", message)
}

func (g *Gateway) onDeviceJoined(eui64 EUI64, message deviceMessage) {
	if d, ok := g.devices.Load(message.Endpoint.EUI64); !ok {
		device := NewDevice(g, &message)
		g.devices.Store(message.Endpoint.EUI64, device)
		g.network.AddDevice(device)
	} else {
		device := d.(*device)
		device.merge(&message)
		g.network.UpdateDevice(device)
	}
}

func (g *Gateway) onDeviceLeft(eui64 EUI64, message deviceLeftMessage) {
	if d, ok := g.devices.Load(message.EUI64); ok {
		device := d.(*device)
		g.network.RemoveDevice(device)
		g.devices.Delete(message.EUI64)
	}
}

func (g *Gateway) onDeviceStateChange(eui64 EUI64, message deviceStateChangeMessage) {
	if d, ok := g.devices.Load(message.EUI64); ok {
		device := d.(*device)
		device.State = message.State
		g.network.UpdateDevice(device)
	}
}

func (g *Gateway) onOTAEvent(eui64 EUI64, message interface{}) {
	log.Println("onOTAEvent: ", message)
}

func (g *Gateway) onExecuted(eui64 EUI64, message interface{}) {
	log.Println("onExecuted: ", message)
}

func (g *Gateway) onZCLResponse(eui64 EUI64, message interface{}) {
	log.Println("onZCLResponse: ", message)
}

func (g *Gateway) onZDOResponse(eui64 EUI64, message interface{}) {
	log.Println("onZDOResponse: ", message)
}

func (g *Gateway) onAPSResponse(eui64 EUI64, message interface{}) {
	//log.Println("onAPSResponse: ", message)
}

func trimTopic(topic string, suffix string) string {
	return strings.Trim(strings.TrimSuffix(strings.TrimPrefix(topic, topicBase), suffix), "/")
}

func (g *Gateway) publish(topic string, message interface{}) error {
	if data, err := json.Marshal(message); err != nil {
		return err
	} else if token := g.client.Publish(fmt.Sprintf("%s/%s/%s", topicBase, net.EUI64(g.eui64), topic), 2, false, data); token.Wait() && token.Error() != nil {
		return token.Error()
	}

	return nil
}

func (g *Gateway) Start(n *net.Network) error {
	if g.client != nil {
		return fmt.Errorf("gateway already started")
	}

	g.network = n

	opts := mqtt.NewClientOptions()
	opts.AddBroker("tcp://localhost:1883")
	opts.SetClientID("iot-gateway")

	g.client = mqtt.NewClient(opts)
	if token := g.client.Connect(); token.Wait() && token.Error() != nil {
		g.client = nil
		return token.Error()
	}

	var messageHandlers = map[string]interface{}{
		"heartbeat":         g.onHeartbeat,
		"devices":           g.onDeviceList,
		"relays":            g.onRelayList,
		"settings":          g.onSettings,
		"devicejoined":      g.onDeviceJoined,
		"deviceleft":        g.onDeviceLeft,
		"devicestatechange": g.onDeviceStateChange,
		"otaevent":          g.onOTAEvent,
		"executed":          g.onExecuted,
		"zclresponse":       g.onZCLResponse,
		"zdoresponse":       g.onZDOResponse,
		"apsresponse":       g.onAPSResponse,
	}

	wg := &sync.WaitGroup{}
	wg.Add(len(messageHandlers))
	for t, h := range messageHandlers {
		go func(t string, h interface{}) {
			method := reflect.ValueOf(h)
			mt := method.Type().In(1)

			if token := g.client.Subscribe(fmt.Sprintf("%s/+/%s", topicBase, t), 2, func(client mqtt.Client, message mqtt.Message) {
				m := reflect.New(mt)
				if e, err := net.NewEUI64(trimTopic(message.Topic(), t)); err != nil {
					log.Println(err)
					return
				} else if err := json.Unmarshal(message.Payload(), m.Interface()); err != nil {
					log.Println(err)
					return
				} else {
					method.Call([]reflect.Value{reflect.ValueOf(EUI64(e)), m.Elem()})
				}
			}); token.Wait() && token.Error() != nil {
				log.Println(token.Error())
			}
			wg.Done()
		}(t, h)
	}
	wg.Wait()

	return nil
}

func (g *Gateway) EUI64() net.EUI64 {
	return net.EUI64(g.eui64)
}

func (g *Gateway) Protocol() string {
	return "zigbee"
}

func (g *Gateway) Stop() error {
	if g.client == nil {
		return fmt.Errorf("gateway not started")
	}

	return nil
}
