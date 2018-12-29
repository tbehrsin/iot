package events

import (
	"sync"
)

type Hub struct {
	channels map[string][]*Channel
	mutex    sync.RWMutex
}

func (e *Hub) create(name string, once bool) *Channel {
	e.mutex.Lock()
	defer e.mutex.Unlock()

	if e.channels == nil {
		e.channels = map[string][]*Channel{}
	}

	if _, ok := e.channels[name]; !ok {
		e.channels[name] = []*Channel{}
	}
	channel := &Channel{make(chan *Event), e, name, once}
	e.channels[name] = append(e.channels[name], channel)
	return channel
}

func (e *Hub) On(name string) *Channel {
	return e.create(name, false)
}

func (e *Hub) Once(name string) *Channel {
	return e.create(name, true)
}

func (e *Hub) Emit(name string, args ...interface{}) {
	e.mutex.Lock()

	if e.channels == nil {
		return
	}

	if _, ok := e.channels[name]; !ok {
		return
	}

	channels := e.channels[name][:]
	channelsToClose := make([]chan *Event, 0)

	for _, l := range channels {
		if l.once {
			l.close(false)
			channelsToClose = append(channelsToClose, l.emitter)
		}
	}

	e.mutex.Unlock()
	event := &Event{name, args}

	for _, l := range channels {
		l.emitter <- event
	}

	for _, e := range channelsToClose {
		close(e)
	}
}

type Event struct {
	Name string
	Args []interface{}
}

type Channel struct {
	emitter chan *Event
	hub     *Hub
	name    string
	once    bool
}

func (c *Channel) Receive() <-chan *Event {
	return c.emitter
}

func (e *Channel) close(lock bool) {
	if lock {
		e.hub.mutex.Lock()
		defer e.hub.mutex.Unlock()
	}

	if e.hub.channels == nil {
		return
	}

	if _, ok := e.hub.channels[e.name]; !ok {
		return
	}

	for i, l := range e.hub.channels[e.name] {
		if l == e {
			e.hub.channels[e.name] = append(e.hub.channels[e.name][:i], e.hub.channels[e.name][i+1:]...)
		}
	}

	if len(e.hub.channels[e.name]) == 0 {
		delete(e.hub.channels, e.name)
	}

	if len(e.hub.channels) == 0 {
		e.hub.channels = nil
	}
}

func (e *Channel) Close() {
	e.close(true)
	close(e.emitter)
}
