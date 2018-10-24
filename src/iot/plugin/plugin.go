package plugin

import (
	"github.com/augustoroman/v8"
	"github.com/augustoroman/v8/v8console"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"sync"
	"time"
)

type Plugin struct {
	context *v8.Context
}

func NewPlugin(filename string) (*Plugin, error) {
	p := &Plugin{}
	p.context = v8.NewIsolate().NewContext()
	v8console.Config{"", os.Stdout, os.Stderr, true}.Inject(p.context)
	p.createContext()

	if data, err := ioutil.ReadFile(filename); err != nil {
		return nil, err
	} else if _, err := p.context.Eval(string(data), filename); err != nil {
		return nil, err
	}

	return p, nil
}

func (p *Plugin) createContext() error {
	router := Router{plugin: p}
	if jso, err := p.context.Create(map[string]interface{}{
		"get": router.Get,
	}); err != nil {
		return err
	} else if err := p.context.Global().Set("router", jso); err != nil {
		return err
	}

	if jso, err := p.context.Create(p.SetTimeout); err != nil {
		return err
	} else if err := p.context.Global().Set("setTimeout", jso); err != nil {
		return err
	}

	return nil
}

func (p *Plugin) SetTimeout(in v8.CallbackArgs) (*v8.Value, error) {
	go func() {
		time.Sleep(time.Duration(in.Args[1].Int64()) * time.Millisecond)
		in.Args[0].Call(nil)
	}()
	return nil, nil
}

func (p *Plugin) Terminate() {
	p.context.Terminate()
}

type Router struct {
	plugin *Plugin
}

func (r *Router) Get(in v8.CallbackArgs) (*v8.Value, error) {
	pathname := in.Args[0]
	handlers := in.Args[1:]

	log.Printf("path: %s, handlers: %+v", pathname, handlers)

	http.HandleFunc(pathname.String(), func(res http.ResponseWriter, req *http.Request) {
		request, _ := r.plugin.NewRequest(req)
		response, _ := r.plugin.NewResponse(res)

		response.mutex.Lock()

		handlers[0].Call(nil, request.value, response.value)

		response.mutex.Lock()
		response.mutex.Unlock()
	})

	return nil, nil
}

type Request struct {
	req   *http.Request
	value *v8.Value
}

func (p *Plugin) NewRequest(req *http.Request) (*Request, error) {
	request := &Request{}
	if value, err := p.context.Create(map[string]interface{}{}); err != nil {
		return nil, err
	} else {
		request.req = req
		request.value = value
		return request, nil
	}
}

type Response struct {
	res   http.ResponseWriter
	mutex *sync.Mutex
	value *v8.Value
}

func (p *Plugin) NewResponse(res http.ResponseWriter) (*Response, error) {
	response := &Response{}
	if value, err := p.context.Create(map[string]interface{}{
		"send": response.Send,
	}); err != nil {
		return nil, err
	} else {
		response.res = res
		response.mutex = &sync.Mutex{}
		response.value = value
		return response, nil
	}
}

func (r *Response) Send(in v8.CallbackArgs) (*v8.Value, error) {
	data := in.Args[0]

	if _, err := r.res.Write([]byte(data.String())); err != nil {
		return nil, err
	}

	r.mutex.Unlock()

	return nil, nil
}
