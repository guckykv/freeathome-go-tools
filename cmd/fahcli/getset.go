package main

import (
	"fmt"
	"github.com/guckykv/freeathome-go-fahapi/fahapi"
	"strings"
)

// GET --------------------------------------

type GetCmd struct {
	Paths []string `arg help:"Paths of datapoints: <DeviceId>.<ChannelId>.<Datapoint>"`
}

func (g *GetCmd) Run(globals *Globals) (err error) {
	if err = initializeApi(globals.Configfile); err != nil {
		return err
	}

	// fmt.Println("get", g.Paths)
	for _, path := range g.Paths {
		if err = getDatapoint(path); err != nil {
			return err
		}
	}
	return nil
}

func getDatapoint(path string) (err error) {
	parts := strings.Split(path, ".")
	if len(parts) != 3 {
		return fmt.Errorf("illegal datapoint path format: %s", path)
	}
	var value string
	if value, err = fahapi.GetDatapoint(defaultSysAP, parts[0], parts[1], parts[2]); err != nil {
		return err
	}
	fmt.Println(value)
	return nil
}

// SET --------------------------------------

type SetCmd struct {
	Assigns struct {
		Assigns map[string]string `arg help:"Setting values via <DeviceId>.<ChannelId>.<Datapoint>=<Value>"`
	} `arg`
}

func (set *SetCmd) Run(globals *Globals) (err error) {
	if err = initializeApi(globals.Configfile); err != nil {
		return err
	}

	for path, value := range set.Assigns.Assigns {
		if err = setDatapoint(path, value); err != nil {
			return err
		}
	}
	return nil
}

func setDatapoint(path string, value string) (err error) {
	parts := strings.Split(path, ".")
	if len(parts) != 3 {
		return fmt.Errorf("illegal datapoint path format: %s", path)
	}
	var ok bool
	if ok, err = fahapi.PutDatapoint(defaultSysAP, parts[0], parts[1], parts[2], value); err != nil {
		return err
	}
	if !ok {
		return fmt.Errorf("failure at setting value %s to datapoint %s", value, path)
	}
	return nil
}
