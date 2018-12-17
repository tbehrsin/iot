package apps

import (
	"encoding/json"
	"fmt"
	"iot/net"
	"log"
	"net/http"
	"sort"

	"github.com/behrsin/go-v8"
	"github.com/masterminds/semver"
)

type Registry struct {
	network   *net.Network
	isolate   *v8.Isolate
	inspector *v8.Inspector
	apps      map[string]*App
}

func NewRegistry(n *net.Network, inspectorCallbacks v8.InspectorCallbacks) (*Registry, error) {
	isolate := v8.NewIsolate()
	inspector := isolate.NewInspector(inspectorCallbacks)

	r := &Registry{
		apps:      make(map[string]*App),
		isolate:   isolate,
		inspector: inspector,
		network:   n,
	}

	return r, nil
}

func (r *Registry) Inspector() *v8.Inspector {
	return r.inspector
}

func (r *Registry) ServeHTTP(res http.ResponseWriter, req *http.Request) {
	for name, app := range r.apps {
		log.Println(name, app)
	}
}

func (r *Registry) Install(name string, tarball string) (string, error) {
	if resp, err := http.Get(tarball); err != nil {
		return "", err
	} else {
		defer resp.Body.Close()

		extractTarball(name, resp.Body)
	}

	return "", nil
}

func (r *Registry) Add(name string, version string) (*App, error) {
	if resp, err := http.Get(fmt.Sprintf("https://registry.npmjs.org/%s/", name)); err != nil {
		return nil, err
	} else {
		defer resp.Body.Close()

		var p map[string]interface{}
		if err := json.NewDecoder(resp.Body).Decode(&p); err != nil {
			return nil, err
		}

		var resolved string

		for v, tag := range p["dist-tags"].(map[string]interface{}) {
			if version == v {
				resolved = tag.(string)
				break
			}
		}

		if resolved == "" {
			versions := make([]*semver.Version, 0)

			for v, _ := range p["versions"].(map[string]interface{}) {
				if sv, err := semver.NewVersion(v); err != nil {
					return nil, err
				} else {
					versions = append(versions, sv)
				}
			}

			sort.Sort(semver.Collection(versions))

			if c, err := semver.NewConstraint(version); err != nil {
				return nil, err
			} else {
				for i := len(versions) - 1; i >= 0; i-- {
					if c.Check(versions[i]) {
						resolved = versions[i].String()
						break
					}
				}
			}
		}

		if resolved == "" {
			// throw an error, unable to resolve a version
		}

		tarball := p["versions"].(map[string]interface{})[resolved].(map[string]interface{})["dist"].(map[string]interface{})["tarball"].(string)

		log.Printf("Installing %s@%s...", name, resolved)
		if _, err := r.Install(name, tarball); err != nil {
			return nil, err
		}

		return nil, nil
	}

	return nil, nil
}
