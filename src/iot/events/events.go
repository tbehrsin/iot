package events

import "sync"

type EventListener func(args ...interface{})

type listenerInfo struct {
	listener *EventListener
	once     bool
}

type EventEmitter struct {
	eventListeners map[string][]listenerInfo
	mutex          sync.RWMutex
}

func (e *EventEmitter) AddListener(name string, listener EventListener) {
	e.mutex.Lock()
	defer e.mutex.Unlock()

	if e.eventListeners == nil {
		e.eventListeners = map[string][]listenerInfo{}
	}

	if _, ok := e.eventListeners[name]; !ok {
		e.eventListeners[name] = []listenerInfo{}
	}

	e.eventListeners[name] = append(e.eventListeners[name], listenerInfo{&listener, false})
}

func (e *EventEmitter) AddOnceListener(name string, listener EventListener) {
	e.mutex.Lock()
	defer e.mutex.Unlock()

	if e.eventListeners == nil {
		e.eventListeners = map[string][]listenerInfo{}
	}

	if _, ok := e.eventListeners[name]; !ok {
		e.eventListeners[name] = []listenerInfo{}
	}

	e.eventListeners[name] = append(e.eventListeners[name], listenerInfo{&listener, true})
}

func (e *EventEmitter) RemoveListener(name string, listener EventListener) {
	e.mutex.Lock()
	defer e.mutex.Unlock()

	if e.eventListeners == nil {
		return
	}

	if _, ok := e.eventListeners[name]; !ok {
		return
	}

	for i, l := range e.eventListeners[name] {
		if l.listener == &listener {
			e.eventListeners[name] = append(e.eventListeners[name][:i], e.eventListeners[name][i+1:]...)
		}
	}

	if len(e.eventListeners[name]) == 0 {
		delete(e.eventListeners, name)
	}

	if len(e.eventListeners) == 0 {
		e.eventListeners = nil
	}
}

func (e *EventEmitter) Emit(name string, args ...interface{}) {
	e.mutex.RLock()

	if e.eventListeners == nil {
		return
	}

	if _, ok := e.eventListeners[name]; !ok {
		return
	}

	listeners := e.eventListeners[name][:]

	e.mutex.RUnlock()

	for _, l := range listeners {
		(*l.listener)(args...)
		if l.once {
			e.RemoveListener(name, *l.listener)
		}
	}
}
