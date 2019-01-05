package net

import (
	"api"
	"encoding/json"
	"fmt"

	"github.com/behrsin/go-v8"
)

type Controller struct {
	value      *v8.Value
	context    *v8.Context
	device     *DeviceProxy
	state      map[string]interface{}
	stateCache *v8.Value
	app        api.Application
	Name       string `v8:"name"`
	Index      string `v8:"index"`
}

func NewController(in v8.FunctionArgs) (*Controller, error) {
	if device, ok := currentDevices[in.Context]; !ok {
		return nil, fmt.Errorf("not a constructor")
	} else {
		app := in.Context.GetIsolate().GetData("app").(api.Application)
		c := &Controller{
			in.This,
			in.Context,
			device,
			map[string]interface{}{},
			nil,
			app,
			fmt.Sprintf("%s %s", device.device.GetManufacturer(), device.device.GetModel()),
			"/index.html",
		}
		in.Context.GetIsolate().AddShutdownHook(c.onShutdownIsolate)
		return c, nil
	}
}

func (c *Controller) V8FuncToString(in v8.FunctionArgs) (*v8.Value, error) {
	return c.device.V8FuncToString(in)
}

func (c *Controller) V8GetDevice(in v8.GetterArgs) (*v8.Value, error) {
	return c.device.V8GetDevice(in)
}

func (c *Controller) V8GetState(in v8.GetterArgs) (*v8.Value, error) {
	if c.stateCache != nil {
		return c.stateCache, nil
	}

	if buf, err := json.Marshal(c.state); err != nil {
		return nil, err
	} else if v, err := c.context.ParseJSON(string(buf)); err != nil {
		return nil, err
	} else {
		c.stateCache = v
		return v, nil
	}
}

func (c *Controller) V8FuncSetState(in v8.FunctionArgs) (*v8.Value, error) {
	var state map[string]interface{}
	force, _ := in.Arg(1).Bool()

	if buf, err := json.Marshal(in.Arg(0)); err != nil {
		return nil, err
	} else if err := json.Unmarshal(buf, &state); err != nil {
		return nil, err
	} else if err := c.SetState(state, force); err != nil {
		return nil, err
	}

	return nil, nil
}

func (c *Controller) SetState(state map[string]interface{}, force bool) error {
	c.stateCache = nil
	var out map[string]interface{}
	if force {
		for k, v := range state {
			c.state[k] = v
		}
	} else if buf, err := json.Marshal(state); err != nil {
		return err
	} else if s, err := c.context.ParseJSON(string(buf)); err != nil {
		return err
	} else if v, err := c.value.CallMethod("onSetState", s); err != nil {
		return err
	} else if s, err := json.Marshal(v); err != nil {
		return err
	} else if err := json.Unmarshal([]byte(s), &out); err != nil {
		return err
	} else {
		for k, v := range out {
			c.state[k] = v
		}
	}

	if err := c.device.Holder().Publish(); err != nil {
		return err
	}

	return nil
}

func (c *Controller) Application() api.Application {
	return c.app
}

func (c *Controller) onShutdownIsolate(i *v8.Isolate) {
	c.state = nil
	c.value = nil
	c.context = nil
	c.stateCache = nil
	c.app = nil
	c.device.Holder().Publish()
	c.device.controller = nil
	c.device.value = nil
}
