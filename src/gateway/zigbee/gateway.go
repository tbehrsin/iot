package zigbee

import (
	"encoding/json"
	"fmt"
	"gateway/net"
	"log"
	"reflect"
	"strings"
	"sync"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

const (
	topicBase         string = "gw"
	Heartbeat                = "heartbeat"
	DeviceList               = "devices"
	RelayList                = "relays"
	Settings                 = "settings"
	DeviceJoined             = "devicejoined"
	DeviceLeft               = "deviceleft"
	DeviceStateChange        = "devicestatechange"
	OTAEvent                 = "otaevent"
	Executed                 = "executed"
	ZCLResponse              = "zclresponse"
	ZDOResponse              = "zdoresponse"
	APSResponse              = "apsresponse"
	CommandList              = "commands"
	PublishState             = "publishstate"
	UpdateSettings           = "updatesettings"
)

var emptyMap = map[string]interface{}{}

type GatewayInterface interface {
	OnHeartbeat(eui64 EUI64, message HeartbeatMessage)
	OnDeviceList(eui64 EUI64, message DeviceListMessage)
	OnRelayList(eui64 EUI64, message RelayListMessage)
	OnSettings(eui64 EUI64, message SettingsMessage)
	OnDeviceJoined(eui64 EUI64, message DeviceMessage)
	OnDeviceLeft(eui64 EUI64, message DeviceLeftMessage)
	OnDeviceStateChange(eui64 EUI64, message DeviceStateChangeMessage)
	OnOTAEvent(eui64 EUI64, message OTAEventMessage)
	OnExecuted(eui64 EUI64, message ExecutedMessage)
	OnZCLResponse(eui64 EUI64, message ZCLResponseMessage)
	OnZDOResponse(eui64 EUI64, message ZDOResponseMessage)
	OnAPSResponse(eui64 EUI64, message APSResponseMessage)
	OnCommandList(eui64 EUI64, message CommandListMessage)
	OnPublishState(eui64 EUI64, message PublishStateMessage)
	OnUpdateSettings(eui64 EUI64, messafe UpdateSettingsMessage)
}

type Gateway struct {
	client   mqtt.Client
	eui64    EUI64
	network  *net.Network
	devices  *sync.Map
	iface    GatewayInterface
	commands chan CommandListMessage
	executed chan string
	mutex    *sync.Mutex
}

func MQTTConnect(client mqtt.Client) {
	retry := time.NewTicker(5 * time.Second)
RetryLoop:
	for {
		select {
		case <-retry.C:
			if token := client.Connect(); token.Wait() && token.Error() != nil {

			} else {
				retry.Stop()
				break RetryLoop
			}
		}
	}
}

func NewGateway(client mqtt.Client) *Gateway {
	if client == nil {
		opts := mqtt.NewClientOptions()
		opts.AddBroker("tcp://localhost:1883")
		opts.SetAutoReconnect(true)
		opts.SetClientID("zigbee-gateway")

		client = mqtt.NewClient(opts)
		MQTTConnect(client)
	}

	gw := &Gateway{
		devices:  &sync.Map{},
		commands: make(chan CommandListMessage),
		executed: make(chan string),
		mutex:    &sync.Mutex{},
		client:   client,
	}

	return gw
}

func (g *Gateway) SetInterface(iface GatewayInterface) {
	g.iface = iface
}

func (g *Gateway) SetEUI64(eui64 EUI64) {
	g.eui64 = eui64
}

func (g *Gateway) OnHeartbeat(eui64 EUI64, message HeartbeatMessage) {
	if net.EUI64(g.eui64) == net.NullEUI64 {
		log.Printf("found gateway: %s\n", eui64)
		g.eui64 = eui64

		if err := g.Publish(PublishState, PublishStateMessage{}); err != nil {
			log.Println(err)
		}

		if !message.NetworkUp {
			if err := g.Publish(CommandList, CommandListMessage{Commands: []Command{
				Command{"plugin network-creator start 1", 5000},
			}}); err != nil {
				log.Println(err)
			}
		}

		if err := g.Publish(CommandList, CommandListMessage{Commands: []Command{
			Command{"plugin network-creator-security open-network", 0},
		}}); err != nil {
			log.Println(err)
		}
	}
}

func (g *Gateway) OnDeviceList(eui64 EUI64, message DeviceListMessage) {
	for _, deviceMessage := range message.Devices {
		g.OnDeviceJoined(eui64, deviceMessage)
	}
}

func (g *Gateway) OnRelayList(eui64 EUI64, message RelayListMessage) {
	log.Println("onRelayList: ", message)
}

func (g *Gateway) OnSettings(eui64 EUI64, message SettingsMessage) {
	log.Println("onSettings: ", message)
}

func (g *Gateway) OnDeviceJoined(eui64 EUI64, message DeviceMessage) {
	if d, ok := g.devices.Load(message.Endpoint.EUI64); !ok {
		device := NewDevice(g, &message)
		g.devices.Store(message.Endpoint.EUI64, device)

		go func() {
			channel := device.Once("merge")

			<-channel.Receive()

			device.onNet = true
			g.network.AddDevice(device)
		}()
	} else {
		device := d.(*Device)

		device.mutex.Lock()
		defer device.mutex.Unlock()

		device.merge(&message)
		if device.onNet {
			g.network.UpdateDevice(device)
		}
	}
}

func (g *Gateway) OnDeviceLeft(eui64 EUI64, message DeviceLeftMessage) {
	if d, ok := g.devices.Load(message.EUI64); ok {
		device := d.(*Device)

		device.mutex.Lock()
		defer device.mutex.Unlock()

		device.onNet = false
		g.network.RemoveDevice(device)
		g.devices.Delete(message.EUI64)
	}
}

func (g *Gateway) OnDeviceStateChange(eui64 EUI64, message DeviceStateChangeMessage) {
	if d, ok := g.devices.Load(message.EUI64); ok {
		device := d.(*Device)

		device.mutex.Lock()
		defer device.mutex.Unlock()

		device.State = message.State
		g.network.UpdateDevice(device)
	}
}

func (g *Gateway) OnOTAEvent(eui64 EUI64, message OTAEventMessage) {
	log.Println("onOTAEvent: ", message)
}

func (g *Gateway) OnExecuted(eui64 EUI64, message ExecutedMessage) {
	g.executed <- message.Command
}

func (g *Gateway) OnZCLResponse(eui64 EUI64, message ZCLResponseMessage) {
	if message.CommandID.Value == 0x00 {
		return
	}

	switch message.CommandID.Value {
	case ReadAttributesResponse:
		g.onReadAttributesResponse(eui64, message)
	case DefaultResponse:
	case WriteAttributesResponse:
		fallthrough
	case ConfigureReportingResponse:
		fallthrough
	case ReadReportingConfigurationResponse:
		fallthrough
	case ReportAttributes:
		fallthrough
	case DiscoverAttributesResponse:
		fallthrough
	case WriteAttributesStructuredResponse:
		fallthrough
	case DiscoverCommandsReceivedResponse:
		fallthrough
	case DiscoverCommandsGeneratedResponse:
		fallthrough
	case DiscoverAttributesExtendedResponse:
		fallthrough
	default:
		log.Println("onZCLResponse: ", message)
	}
}

func (g *Gateway) OnZDOResponse(eui64 EUI64, message ZDOResponseMessage) {
	log.Println("onZDOResponse: ", message)
}

func (g *Gateway) OnAPSResponse(eui64 EUI64, message APSResponseMessage) {
	//log.Println("onAPSResponse: ", message)
}

func (g *Gateway) OnCommandList(eui64 EUI64, message CommandListMessage) {

}

func (g *Gateway) OnPublishState(eui64 EUI64, message PublishStateMessage) {

}

func (g *Gateway) OnUpdateSettings(eui64 EUI64, messafe UpdateSettingsMessage) {

}

func trimTopic(topic string, suffix string) string {
	return strings.Trim(strings.TrimSuffix(strings.TrimPrefix(topic, topicBase), suffix), "/")
}

func (g *Gateway) Publish(topic string, message interface{}) error {
	g.mutex.Lock()
	defer g.mutex.Unlock()

	if data, err := json.Marshal(message); err != nil {
		return err
	} else if token := g.client.Publish(fmt.Sprintf("%s/%s/%s", topicBase, net.EUI64(g.eui64), topic), 2, false, data); token.Wait() && token.Error() != nil {
		return token.Error()
	}

	return nil
}

func (g *Gateway) processCommands() {
	for {
		command := <-g.commands

		go func() {
			if err := g.Publish(CommandList, command); err != nil {
				log.Println(err)
			}
		}()

		for {
			executed := <-g.executed

			for i, c := range command.Commands {
				if c.Command == executed {
					command.Commands = append(command.Commands[:i], command.Commands[i+1:]...)
				}
			}

			if len(command.Commands) == 0 {
				break
			}
		}
	}
}

func (g *Gateway) Start(n *net.Network) error {
	g.network = n

	iface := g.iface
	if iface == nil {
		iface = g
	}

	go g.processCommands()

	var messageHandlers = map[string]interface{}{
		Heartbeat:         iface.OnHeartbeat,
		DeviceList:        iface.OnDeviceList,
		RelayList:         iface.OnRelayList,
		Settings:          iface.OnSettings,
		DeviceJoined:      iface.OnDeviceJoined,
		DeviceLeft:        iface.OnDeviceLeft,
		DeviceStateChange: iface.OnDeviceStateChange,
		OTAEvent:          iface.OnOTAEvent,
		Executed:          iface.OnExecuted,
		ZCLResponse:       iface.OnZCLResponse,
		ZDOResponse:       iface.OnZDOResponse,
		APSResponse:       iface.OnAPSResponse,
		CommandList:       iface.OnCommandList,
		PublishState:      iface.OnPublishState,
		UpdateSettings:    iface.OnUpdateSettings,
	}

	wg := &sync.WaitGroup{}
	wg.Add(len(messageHandlers))
	for t, h := range messageHandlers {
		go func(t string, h interface{}) {
			g.mutex.Lock()
			defer g.mutex.Unlock()

			method := reflect.ValueOf(h)
			mt := method.Type().In(1)
			if token := g.client.Subscribe(fmt.Sprintf("%s/+/%s", topicBase, t), 2, func(client mqtt.Client, message mqtt.Message) {
				g.mutex.Lock()
				defer g.mutex.Unlock()

				m := reflect.New(mt)
				if e, err := net.NewEUI64(trimTopic(message.Topic(), t)); err != nil {
					log.Println(err)
					return
				} else if err := json.Unmarshal(message.Payload(), m.Interface()); err != nil {
					log.Println(err)
					return
				} else {
					go method.Call([]reflect.Value{reflect.ValueOf(EUI64(e)), m.Elem()})
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

	close(g.commands)
	close(g.executed)

	return nil
}
