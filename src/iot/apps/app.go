package apps

import (
	"encoding/json"
	"fmt"
	"github.com/tbehrsin/v8"
	"github.com/tbehrsin/v8/v8console"
	"io/ioutil"
	"net/http"
	"os"
	"time"
)

type App struct {
	context *v8.Context
	value   *v8.Value
	routes  []Route
	Name    string
}

func (r *Registry) Load(name string) (*App, error) {
	var p map[string]string
	if fd, err := os.Open(fmt.Sprintf("%s/package.json", name)); err != nil {
		return nil, err
	} else if err := json.NewDecoder(fd).Decode(&p); err != nil {
		return nil, err
	}

	filename := fmt.Sprintf("%s/%s", name, p["main"])

	a := &App{
		routes: make([]Route, 0, 4),
		Name:   name,
	}
	a.context = v8.NewIsolate().NewContext()
	v8console.Config{"", os.Stdout, os.Stderr, true}.Inject(a.context)
	a.createContext()
	if err := a.eval(filename); err != nil {
		return nil, err
	}

	r.apps[name] = a

	http.Handle(fmt.Sprintf("/api/v1/apps/%s/", a.Name), a)

	return a, nil
}

func (a *App) eval(filename string) error {
	if data, err := ioutil.ReadFile(filename); err != nil {
		return err
	} else if _, err := a.context.Eval(string(data), filename); err != nil {
		return err
	}

	return nil
}

func (a *App) createContext() error {
	a.injectApp()
	a.injectRouter()

	if jso, err := a.context.Create(a.SetTimeout); err != nil {
		return err
	} else if err := a.context.Global().Set("setTimeout", jso); err != nil {
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

func (a *App) SetTimeout(in v8.CallbackArgs) (*v8.Value, error) {
	go func() {
		time.Sleep(time.Duration(in.Args[1].Int64()) * time.Millisecond)
		in.Args[0].Call(nil)
	}()
	return nil, nil
}

func (a *App) Terminate() {
	a.context.Terminate()
}
