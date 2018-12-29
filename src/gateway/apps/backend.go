package apps

import (
	"api"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

type LocalBackend struct {
	path string
}

func NewLocalBackend(path string) api.Backend {
	return &LocalBackend{path}
}

func (m *LocalBackend) ReadFile(path string) ([]byte, error) {
	p := filepath.Join(m.path, strings.TrimPrefix(path, "/"))
	return ioutil.ReadFile(p)
}

func (m *LocalBackend) IsDir(path string) bool {
	p := filepath.Join(m.path, strings.TrimPrefix(path, "/"))
	if fi, err := os.Lstat(p); err == nil && fi.Mode().IsDir() {
		return true
	}
	return false
}

func (m *LocalBackend) IsExist(path string) bool {
	p := filepath.Join(m.path, strings.TrimPrefix(path, "/"))
	if _, err := os.Stat(p); os.IsExist(err) {
		return true
	}
	return false
}
