package main

import (
	"fmt"
	"github.com/guckykv/freeathome-go-fahapi/fahapi"
)

type DevCmd struct {
	DeviceId struct {
		Short    bool   `help:"List only the IDs" short:"s"`
		DeviceId string `arg optional help:"Show given device. Without id list all device ids."`
	} `arg`
}

func (dev *DevCmd) Run(globals *Globals) error {
	var err error
	if err = initializeApi(globals.Configfile); err != nil {
		return err
	}
	if dev.DeviceId.DeviceId != "" {
		// print structure of one device
		var device *fahapi.Device
		if device, err = fahapi.GetDevice(defaultSysAP, dev.DeviceId.DeviceId); err != nil {
			return err
		}
		var json []byte
		json, err = unmarshall(device)
		if err != nil {
			fmt.Printf("Error unmarshall object: %v", err)
			return err
		}
		fmt.Printf("%s\n", json)
	} else {
		// print plain list of all devices
		var dl *fahapi.Devicelist
		if dl, err = fahapi.GetDeviceList(); err != nil {
			return err
		}
		for _, devId := range dl.AdditionalProperties {
			fmt.Printf("%s\n", devId)
		}
	}

	return nil
}
