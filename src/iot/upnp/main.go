package upnp

import (
	igd "github.com/emersion/go-upnp-igd"
	"log"
	"math/rand"
	"time"
)

const TimeToLive = time.Duration(60) * time.Second

type MappingListener func(m *Mapping)

type Mapping struct {
	LocalPort   uint16
	PublicPort  uint16
	ExternalIP  string
	Description string
	gateway     *igd.IGD
	listener    MappingListener
}

func NewMapping(localPort uint16, externalPort uint16, description string, listener MappingListener) *Mapping {
	m := &Mapping{
		LocalPort:   localPort,
		PublicPort:  externalPort,
		ExternalIP:  "",
		Description: description,
		listener:    listener,
	}

	go m.Update()
	return m
}

func (m *Mapping) Update() {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))

	ch := make(chan igd.Device, 5)
	done := make(chan error)

	if m.gateway != nil {
		ch <- m.gateway
	}

	go func() {
		if err := igd.Discover(ch, time.Duration(3)*time.Second); err != nil {
			close(ch)
			done <- err
			return
		}
	}()

	go func() {
		for {
			device, more := <-ch
			gateway := device.(*igd.IGD)

			if !more {
				m.gateway = nil
				m.PublicPort = 0
				m.ExternalIP = ""
				done <- nil
				close(done)
				return
			}

			// try a random port above 10000 and continue until successful
			i := 0
			for {
				old := *m
				port := m.PublicPort
				if port == 0 {
					port = uint16(10000 + r.Intn(55535))
				}

				if assignedPort, err := gateway.AddPortMapping(igd.TCP, int(m.LocalPort), int(port), m.Description, TimeToLive); err != nil {
					i++
					done <- err
					if i == 5 {
						break
					} else {
						continue
					}
				} else {
					m.gateway = gateway
					m.PublicPort = uint16(assignedPort)
				}
				if ip, err := gateway.GetExternalIPAddress(); err != nil {
					i++
					done <- err
					if i == 5 {
						break
					} else {
						continue
					}
				} else {
					m.ExternalIP = ip.String()
				}

				if m.listener != nil && old.gateway != gateway || old.PublicPort != m.PublicPort || old.ExternalIP != m.ExternalIP {
					m.listener(m)
				}

				done <- nil
				close(done)
				return
			}
		}
	}()

	for {
		err, more := <-done
		if err != nil {
			log.Printf("error during internet gateway update: %+v\n", err)
		}

		if !more {
			// schedule another update before the port mapping is about to expire
			go func() {
				time.Sleep(TimeToLive - time.Duration(5)*time.Second)
				m.Update()
			}()

			return
		}
	}
}

func (m *Mapping) Close() {
	if m.gateway != nil && m.PublicPort != 0 {
		m.gateway.DeletePortMapping(igd.TCP, int(m.PublicPort))
		m.gateway = nil
		m.PublicPort = 0
		m.ExternalIP = ""
		if m.listener != nil {
			m.listener(m)
		}
	}
}
