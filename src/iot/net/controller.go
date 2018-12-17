package net

import (
	"fmt"

	"github.com/behrsin/go-v8"
)

type Controller struct {
	device *DeviceProxy
}

func NewController(in v8.FunctionArgs) (*Controller, error) {
	if device, ok := currentDevices[in.Context]; !ok {
		return nil, fmt.Errorf("not a constructor")
	} else {
		c := &Controller{device}
		return c, nil
	}
}

func (c *Controller) V8FuncToString(in v8.FunctionArgs) (*v8.Value, error) {
	return c.device.V8FuncToString(in)
}

func (c *Controller) V8GetDevice(in v8.GetterArgs) (*v8.Value, error) {
	return c.device.V8GetDevice(in)
}

func (c Controller) V8GetTest(in v8.GetterArgs) (*v8.Value, error) {
	return c.device.V8GetDevice(in)
}

type TestConstructor struct {
	Test2 string `v8:"test2"`
}

func NewTestConstructor(in v8.FunctionArgs) (*TestConstructor, error) {
	return &TestConstructor{}, nil
}

func (c *TestConstructor) V8FuncToString(in v8.FunctionArgs) (*v8.Value, error) {
	return in.Context.Create("Hello World")
}
