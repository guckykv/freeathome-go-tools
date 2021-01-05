package main

import (
	"fmt"
	"github.com/guckykv/freeathome-go-fahapi/fahapi"
	"strconv"
	"time"
)

type VirtCmd struct {
	Serial string `arg required help:"Internal serial string"`
	Ttl    int    `arg required help:"Ttl (time to live) Use 0 for delete, -1 for forever or NNN seconds"`
	Create bool   `optional short:"n" help:"Create a new virtual device (needs Type and Name)"`
	Name   string `arg optional help:"Display Name for the new device"`
	Type   string `arg optional help:"One of type VirtualDeviceType"`
}

func (virtCmd *VirtCmd) Run(globals *Globals) (err error) {
	if err = initializeApi(globals.Configfile); err != nil {
		return
	}

	var reqBody *fahapi.VirtualDevice

	if virtCmd.Create {
		if reqBody, err = virtCmd.vCreate(globals); err != nil {
			return
		}
	} else {
		reqBody = &fahapi.VirtualDevice{
			Properties: fahapi.VirtualDeviceProperties{
				Ttl: strconv.Itoa(virtCmd.Ttl),
			},
		}
	}

	var virtualSerial string
	if virtualSerial, err = fahapi.PutVirtualDevice(defaultSysAP, virtCmd.Serial, reqBody); err != nil {
		return
	}

	fmt.Println(virtualSerial)
	return
}

func (virtCmd *VirtCmd) vCreate(globals *Globals) (*fahapi.VirtualDevice, error) {
	if virtCmd.Name == "" {
		return nil, fmt.Errorf("virtual: missing name for creating new virtual device")
	}
	if virtCmd.Type == "" {
		return nil, fmt.Errorf("virtual: missing type for creating new virtual device")
	}

	reqBody := fahapi.VirtualDevice{
		Properties: fahapi.VirtualDeviceProperties{
			Displayname: virtCmd.Name,
			Ttl:         strconv.Itoa(virtCmd.Ttl),
		},
		Type: fahapi.VirtualDeviceType(virtCmd.Type),
	}

	return &reqBody, nil
}

func vDebug(virtualSerial string) (err error) {
	var virtualDev *fahapi.Device
	virtualDev, err = fahapi.GetDevice(defaultSysAP, virtualSerial)

	time.Sleep(13 * time.Second)

	var json []byte
	json, err = unmarshall(virtualDev)
	fmt.Println(string(json))

	_, err = fahapi.PutDatapoint(defaultSysAP, virtualSerial, "ch0000", "odp0000", "1")
	return err
}
