package apps

import (
	"api"
	"encoding/json"
	"gateway/net"
	"path/filepath"
	"reflect"
	"strings"
	"time"

	"github.com/behrsin/go-v8"
)

type App struct {
	backend     api.Backend
	pkg         api.Package
	debug       bool
	network     *net.Network
	registry    *Registry
	isolate     *v8.Isolate
	context     *v8.Context
	value       *v8.Value
	routes      []Route
	Module      *Module
	moduleCache map[string]*Module
}

type PackageIoT struct {
	Package_Public *string `json:"public,omitempty"`
}

type Package struct {
	Package_Name *string     `json:"name,omitempty"`
	Package_Main *string     `json:"main,omitempty"`
	Package_IoT  *PackageIoT `json:"iot,omitempty"`
}

func (p *Package) Main() string {
	if p.Package_Main == nil {
		return "/index.js"
	} else {
		return filepath.Join("/", strings.TrimPrefix(*p.Package_Main, "/"))
	}
}

func (p *Package) Name() string {
	if p.Package_Name == nil {
		return "(unnamed)"
	} else {
		return *p.Package_Name
	}
}

func (p *Package) Public() string {
	def := "/dist/"

	if p.Package_IoT == nil {
		return def
	} else if p.Package_IoT.Package_Public == nil {
		return def
	} else {
		return filepath.Join("/", strings.TrimPrefix(*p.Package_IoT.Package_Public, "/"))
	}
}

func (r *Registry) LoadFromName(name string) (api.Application, error) {
	return r.Load(NewLocalBackend(name))
}

func (r *Registry) Load(backend api.Backend) (api.Application, error) {
	var p Package
	if b, err := backend.ReadFile("/package.json"); err != nil {
		return nil, err
	} else if err := json.Unmarshal(b, &p); err != nil {
		return nil, err
	}

	a := &App{
		backend:     backend,
		pkg:         &p,
		debug:       true,
		network:     r.network,
		registry:    r,
		routes:      make([]Route, 0, 4),
		moduleCache: map[string]*Module{},
	}

	a.isolate = v8.NewIsolate()
	a.isolate.SetData("app", a)
	a.context, _ = a.isolate.NewContext()
	global, _ := a.context.Global()

	if err := a.createContext(global); err != nil {
		a.Terminate()
		return nil, err
	} else if module, err := a.NewModuleFromFile(p.Main(), nil); err != nil {
		a.Terminate()
		return nil, err
	} else {
		a.Module = module
	}

	r.inspector.AddApp(a)
	return a, nil
}

func (a *App) Context() *v8.Context {
	return a.context
}

func (a *App) createContext(global *v8.Value) error {
	if err := global.Set("global", global); err != nil {
		return err
	} else if err := a.injectApp(global); err != nil {
		return err
	} else if err := a.injectRouter(global); err != nil {
		return err
	} else if err := a.injectGateways(global); err != nil {
		return err
	}

	if jso, err := a.context.Create(a.setTimeout); err != nil {
		return err
	} else if err := global.Set("setTimeout", jso); err != nil {
		return err
	}

	return nil
}

func (a *App) injectApp(global *v8.Value) error {
	if jso, err := a.context.Create(a); err != nil {
		return err
	} else if err := global.Set("app", jso); err != nil {
		return err
	} else {
		a.value = jso
		value := reflect.ValueOf(a)
		a.value.SetReceiver(&value)
	}

	return nil
}

func (a *App) injectGateways(global *v8.Value) error {
	for _, gateway := range a.network.Gateways() {
		if jso, err := a.context.Create(gateway); err != nil {
			return err
		} else if err := global.Set(gateway.Protocol(), jso); err != nil {
			return err
		}
	}

	return nil
}

func (a *App) setTimeout(in v8.FunctionArgs) (*v8.Value, error) {
	duration, _ := in.Args[1].Int64()
	fn := in.Args[0]

	go func() {
		time.Sleep(time.Duration(duration) * time.Millisecond)
		go fn.Call(nil)
	}()
	return nil, nil
}

func (a *App) Terminate() {
	a.registry.inspector.RemoveApp(a)
	if a.isolate != nil {
		a.isolate.Terminate()
	}
	a.Module = nil
	a.moduleCache = nil
	a.backend = nil
	a.registry = nil
	a.isolate = nil
	a.context = nil
	a.value = nil
	a.routes = nil
}

func (a *App) Backend() api.Backend {
	return a.backend
}

func (a *App) Package() api.Package {
	return a.pkg
}
