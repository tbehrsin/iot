package apps

import (
	"fmt"
	"path"
	"path/filepath"
	"strings"

	v8 "github.com/behrsin/go-v8"
)

type Module struct {
	App      *App
	require  *v8.Value
	value    *v8.Value
	ID       string    `v8:"id"`
	Exports  *v8.Value `v8:"exports"`
	Parent   *Module   `v8:"parent"`
	Filename string    `v8:"filename"`
	Dirname  string
	Loaded   bool      `v8:"loaded"`
	Children []*Module `v8:"children"`
	Paths    []string  `v8:"paths"`
}

func wrapScript(code string) string {
	return fmt.Sprintf("(exports, require, module, __filename, __dirname) => {\n%s\n}", code)
}

func (a *App) NewModuleFromFile(filename string, parent *Module) (*Module, error) {
	if m, ok := a.moduleCache[filename]; ok {
		return m, nil
	} else if exports, err := a.context.Create(map[string]interface{}{}); err != nil {
		return nil, err
	} else if data, err := a.backend.ReadFile(filename); err != nil {
		return nil, err
	} else if script, err := a.context.Run(wrapScript(string(data)), filename); err != nil {
		return nil, err
	} else {
		absdir := filepath.Dir(filename)

		m := &Module{
			App:      a,
			ID:       filename,
			Exports:  exports,
			Parent:   parent,
			Filename: filename,
			Dirname:  absdir,
			Loaded:   false,
			Children: []*Module{},
		}

		// a.isolate.AddShutdownHook(m)

		a.moduleCache[filename] = m

		if parent != nil {
			parent.Children = append(parent.Children, m)
		}
		if p, err := m.newPaths(); err != nil {
			return nil, err
		} else {
			m.Paths = p
		}

		if module, err := a.context.Create(m); err != nil {
			return nil, err
		} else {
			m.value = module

			if require, err := m.newRequire(); err != nil {
				return nil, err
			} else if filename, err := a.context.Create(filename); err != nil {
				return nil, err
			} else if dirname, err := a.context.Create(absdir); err != nil {
				return nil, err
			} else if _, err := script.Call(nil, exports, require, module, filename, dirname); err != nil {
				return nil, err
			}

			m.Loaded = true
		}

		return m, nil
	}
}

func (m *Module) newRequire() (*v8.Value, error) {
	if m.require != nil {
		return m.require, nil
	}

	if require, err := m.value.Get("require"); err != nil {
		return nil, err
	} else if breq, err := require.Bind(m.value); err != nil {
		return nil, err
	} else if resolve, err := m.App.context.Create(m.Resolve); err != nil {
		return nil, err
	} else if err := breq.Set("resolve", resolve); err != nil {
		return nil, err
	} else {
		m.require = breq
		return m.require, nil
	}
}

func (m *Module) Resolve(in v8.FunctionArgs) (*v8.Value, error) {
	src := m.Dirname
	dst := in.Arg(0).String()
	abs := dst

	if strings.HasPrefix(abs, ".") {
		abs = path.Join(src, abs)

		if path.Ext(abs) == "" {
			abs = abs + ".js"
		}
	} else {
		for _, p := range m.Paths {
			t := path.Join(p, abs)
			if m.App.backend.IsDir(t) {
				abs = t
				break
			}

			if path.Ext(t) == "" {
				t = t + ".js"
			}

			if m.App.backend.IsExist(t) {
				abs = t
				break
			}
		}
	}

	if m.App.backend.IsDir(abs) {
		abs = path.Join(abs, "index.js")
	}

	return in.Context.Create(abs)
}

func (m *Module) V8FuncRequire(in v8.FunctionArgs) (*v8.Value, error) {
	if p, err := m.Resolve(in); err != nil {
		return nil, err
	} else if module, err := m.App.NewModuleFromFile(p.String(), m); err != nil {
		return nil, err
	} else {
		return module.Exports, nil
	}
}

func (m *Module) newPaths() ([]string, error) {
	pre := "/"
	out := []string{}

	for _, p := range strings.Split(m.Dirname, "/") {
		pre = filepath.Join(pre, p)

		if strings.HasPrefix(p, "@") || p == "node_modules" {
			continue
		}

		out = append([]string{filepath.Join(pre, "node_modules")}, out...)
	}
	return out, nil
}

func (m *Module) ShutdownIsolate(i *v8.Isolate) {
	m.App = nil
	m.require = nil
	m.value = nil
	m.Exports = nil
	m.Parent = nil
	m.Children = nil
	m.Paths = nil
}
