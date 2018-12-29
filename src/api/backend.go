package api

import "reflect"

type Backend interface {
	ReadFile(path string) ([]byte, error)
	IsDir(path string) bool
	IsExist(path string) bool
}

var BackendType = reflect.TypeOf((*Backend)(nil)).Elem()

type Application interface {
	Backend() Backend
	Package() Package
	// EnableInspector()
	// DisableInspector()
	// Name()
	// Main() Module
	// Require(id string)
	// Reload()

	Terminate()
}

var ApplicationType = reflect.TypeOf((*Application)(nil)).Elem()

type Package interface {
	Name() string
	Main() string
	Public() string
}

var PackageType = reflect.TypeOf((*Package)(nil)).Elem()

type Module interface {
	// Application() Application
	// GetID() string
	// Resolve(id string) (string, error)
	// Require(id string)
	// GetParent() Module
	// GetChildren() []Module
	// GetFilename() string
	// GetDirname() string
	// GetPaths() []string
}

var ModuleType = reflect.TypeOf((*Module)(nil)).Elem()

type Controller interface {
	Application() Application
}

var ControllerType = reflect.TypeOf((*Controller)(nil)).Elem()
