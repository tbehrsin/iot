package apps

import (
	"fmt"
	"github.com/tbehrsin/v8"
	"log"
	"net/http"
	"strings"
	"sync"
)

type Route struct {
	Method  string
	Path    string
	Handler *v8.Value
}

type Router interface {
	AppendRoute(route *Route)
	Routes() []Route
	Handle(in v8.CallbackArgs) (*v8.Value, error)
}

type RouterImpl struct {
	app    *App
	routes []Route
}

func (a *App) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	request, _ := a.NewRequest(r, fmt.Sprintf("/api/v1/apps/%s/", a.Name))
	response, _ := a.NewResponse(w)

	response.mutex.Lock()

	app := a.value
	handle, _ := app.Get("handle")

	next, _ := a.context.Create(func(in v8.CallbackArgs) (*v8.Value, error) {
		res := response.value

		status, _ := res.Get("status")
		i, _ := in.Context.Create(404)
		status.Call(res, i)

		send, _ := res.Get("send")
		t, _ := in.Context.Create("404 Not Found")
		send.Call(res, t)

		return nil, nil
	})

	handle.Call(app, request.value, response.value, next)

	response.mutex.Lock()
	response.mutex.Unlock()
}

func (a *App) injectRouter() error {
	if jso, err := a.context.Create(func(in v8.CallbackArgs) (*v8.Value, error) {
		router := &RouterImpl{
			app:    a,
			routes: make([]Route, 0, 4),
		}
		if jso, err := a.context.Create(router); err != nil {
			return nil, err
		} else {
			return jso, nil
		}
	}); err != nil {
		return err
	} else if err := a.context.Global().Set("Router", jso); err != nil {
		return err
	}

	return nil
}

func createRouteNextHandler(router Router, routes []Route, i int, in v8.CallbackArgs) *v8.Value {
	if len(routes) <= i+1 {
		return in.Arg(2)
	} else {
		r := routes[i+1:]
		next, _ := in.Context.Create(func(in2 v8.CallbackArgs) (*v8.Value, error) {
			//err := in2.Arg(0)
			routeHandler(router, r, in)
			return nil, nil
		})
		return next
	}
}

func routeHandler(router Router, routes []Route, in v8.CallbackArgs) (*v8.Value, error) {
	req := in.Arg(0)
	res := in.Arg(1)
	next := in.Arg(2)

	for i, route := range routes {
		if route.Method == "" {
			// app.use
			if path, _ := req.Get("path"); strings.Index(path.String(), route.Path) == 0 {
				if route.Handler.IsKind(v8.KindFunction) {
					if _, err := route.Handler.Call(nil, req, res, createRouteNextHandler(router, routes, i, in)); err != nil {
						log.Fatal(err)
					}
					return nil, nil
				} else {
					v, _ := route.Handler.Get("handle")
					copy, _ := req.Get("copy")
					pathSuffix, _ := in.Context.Create(path.String()[len(strings.TrimSuffix(route.Path, "/")):])
					req, _ = copy.Call(req, pathSuffix)
					if _, err := v.Call(nil, req, res, createRouteNextHandler(router, routes, i, in)); err != nil {
						log.Fatal(err)
					}
					return nil, nil
				}
			}
		} else if v, _ := req.Get("method"); route.Method == "*" || route.Method == v.String() {
			// app.all, app.get, app.post, etc
			if v, _ := req.Get("path"); v.String() == route.Path {
				if route.Handler.IsKind(v8.KindFunction) {
					if _, err := route.Handler.Call(nil, req, res, createRouteNextHandler(router, routes, i, in)); err != nil {
						log.Fatal(err)
					}
					return nil, nil
				} else {
					v, _ := route.Handler.Get("handle")
					if _, err := v.Call(nil, req, res, createRouteNextHandler(router, routes, i, in)); err != nil {
						log.Fatal(err)
					}
					return nil, nil
				}
			}
		}
	}

	next.Call(nil)
	return nil, nil
}

func (a *App) Handle(in v8.CallbackArgs) (*v8.Value, error) {
	return routeHandler(a, a.Routes(), in)
}

func (a *App) AppendRoute(route *Route) {
	a.routes = append(a.routes, *route)
}

func (a *App) Routes() []Route {
	return a.routes
}

func (r *RouterImpl) Handle(in v8.CallbackArgs) (*v8.Value, error) {
	return routeHandler(r, r.Routes(), in)
}

func (r *RouterImpl) AppendRoute(route *Route) {
	r.routes = append(r.routes, *route)
}

func (r *RouterImpl) Routes() []Route {
	return r.routes
}

func (a *App) Use(in v8.CallbackArgs) (*v8.Value, error) {
	pathname := in.Args[0]
	handlers := in.Args[1:]

	route := Route{
		Method:  "",
		Path:    pathname.String(),
		Handler: handlers[0],
	}

	a.routes = append(a.routes, route)

	return nil, nil
}

func (r *RouterImpl) Use(in v8.CallbackArgs) (*v8.Value, error) {
	pathname := in.Args[0]
	handlers := in.Args[1:]

	route := Route{
		Method:  "",
		Path:    pathname.String(),
		Handler: handlers[0],
	}

	r.routes = append(r.routes, route)

	return nil, nil
}

func (r *RouterImpl) Get(in v8.CallbackArgs) (*v8.Value, error) {
	pathname := in.Args[0]
	handlers := in.Args[1:]

	route := Route{
		Method:  http.MethodGet,
		Path:    pathname.String(),
		Handler: handlers[0],
	}

	r.routes = append(r.routes, route)

	return nil, nil
}

type Request struct {
	Path        string `json:"path"`
	OriginalURL string `json:"originalUrl"`
	Method      string `json:"method"`
	req         *http.Request
	value       *v8.Value
}

func (a *App) NewRequest(req *http.Request, mountPoint string) (*Request, error) {
	request := &Request{
		Path:        req.URL.Path[len(strings.TrimSuffix(mountPoint, "/")):],
		OriginalURL: req.URL.String(),
		Method:      req.Method,
	}
	if value, err := a.context.Create(request); err != nil {
		return nil, err
	} else {
		request.req = req
		request.value = value
		return request, nil
	}
}

func (r *Request) Copy(in v8.CallbackArgs) (*v8.Value, error) {
	path := in.Arg(0).String()

	request := &Request{
		Path:        path,
		OriginalURL: r.OriginalURL,
		Method:      r.Method,
		req:         r.req,
	}

	if value, err := in.Context.Create(request); err != nil {
		return nil, err
	} else {
		request.value = value
		return request.value, nil
	}
}

type Response struct {
	res   http.ResponseWriter
	mutex *sync.Mutex
	value *v8.Value
}

func (a *App) NewResponse(res http.ResponseWriter) (*Response, error) {
	response := &Response{}
	if value, err := a.context.Create(response); err != nil {
		return nil, err
	} else {
		response.res = res
		response.mutex = &sync.Mutex{}
		response.value = value
		return response, nil
	}
}

func (r *Response) Status(in v8.CallbackArgs) (*v8.Value, error) {
	status := in.Args[0].Float64()

	r.res.WriteHeader(int(status))

	return nil, nil
}

func (r *Response) Send(in v8.CallbackArgs) (*v8.Value, error) {
	data := in.Arg(0)

	if _, err := r.res.Write([]byte(data.String())); err != nil {
		return nil, err
	}

	r.mutex.Unlock()

	return nil, nil
}

func (r *Response) Json(in v8.CallbackArgs) (*v8.Value, error) {
	data := in.Arg(0)

	if b, err := data.MarshalJSON(); err != nil {
		return nil, err
	} else {
		if _, err := r.res.Write(b); err != nil {
			return nil, err
		}

		r.mutex.Unlock()

		return nil, nil
	}

}
