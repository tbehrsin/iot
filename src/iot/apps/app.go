package apps

import (
	"encoding/json"
	"fmt"
	"iot/net"
	"net/http"
	"os"
	"time"

	"github.com/behrsin/go-v8"
)

type App struct {
	debug       bool
	network     *net.Network
	registry    *Registry
	context     *v8.Context
	value       *v8.Value
	routes      []Route
	Name        string
	Module      *Module
	moduleCache map[string]*Module
}

func (r *Registry) Load(name string) (*App, error) {
	var p map[string]interface{}
	if fd, err := os.Open(fmt.Sprintf("%s/package.json", name)); err != nil {
		return nil, err
	} else if err := json.NewDecoder(fd).Decode(&p); err != nil {
		return nil, err
	}

	filename := fmt.Sprintf("%s/%s", name, p["main"])

	a := &App{
		debug:       true,
		network:     r.network,
		registry:    r,
		routes:      make([]Route, 0, 4),
		Name:        name,
		moduleCache: map[string]*Module{},
	}
	a.context = r.isolate.NewContext()

	if a.debug {
		r.inspector.AddContext(a.context, name)
	}

	a.createContext()
	if module, err := a.NewModuleFromFile(filename, nil); err != nil {
		return nil, err
	} else {
		a.Module = module
	}

	r.apps[name] = a

	http.Handle(fmt.Sprintf("/api/v1/apps/%s/", a.Name), a)

	return a, nil
}

func (a *App) Context() *v8.Context {
	return a.context
}

func (a *App) createContext() error {

	if err := a.context.Global().Set("global", a.context.Global()); err != nil {
		return err
	}

	a.injectGateways()
	a.injectApp()
	a.injectRouter()

	if jso, err := a.context.Create(a.setTimeout); err != nil {
		return err
	} else if err := a.context.Global().Set("setTimeout", jso); err != nil {
		return err
	}

	if jso, err := a.context.Create(net.NewTestConstructor); err != nil {
		return err
	} else if err := a.context.Global().Set("Test", jso); err != nil {
		return err
	}

	return nil
}

func (a *App) injectApp() error {
	if jso, err := a.context.Create(a); err != nil {
		return err
	} else if err := a.context.Global().Set("app", jso); err != nil {
		return err
	} else {
		a.value = jso
	}

	return nil
}

func (a *App) injectGateways() error {
	for _, gateway := range a.network.Gateways() {
		if jso, err := a.context.Create(gateway); err != nil {
			return err
		} else if err := a.context.Global().Set(gateway.Protocol(), jso); err != nil {
			return err
		}
	}

	return nil
}

func (a *App) setTimeout(in v8.FunctionArgs) (*v8.Value, error) {
	go func() {
		time.Sleep(time.Duration(in.Args[1].Int64()) * time.Millisecond)
		in.Args[0].Call(nil)
	}()
	return nil, nil
}

func (a *App) Terminate() {
	if a.debug {
		a.registry.Inspector().RemoveContext(a.context)
	}
	a.context.GetIsolate().Terminate()
}
