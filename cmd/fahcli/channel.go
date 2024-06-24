package main

import (
	"fmt"
	"github.com/guckykv/freeathome-go-fahapi/fahapi"
	"strings"
)

type ChanCmd struct {
	DeviceId  string `arg required help:"DeviceId or DeviceId.ChannelId"`
	ChannelId string `arg required help:"Channel Id"`
}

func (c *ChanCmd) Run(globals *Globals) (err error) {
	if err = initializeApi(globals.Configfile); err != nil {
		return
	}

	var deviceId, channelId string
	parts := strings.Split(c.DeviceId, ".")
	if len(parts) == 2 {
		if c.ChannelId != "" {
			return fmt.Errorf("channelId given several times: multipart %s and %s", c.DeviceId, c.ChannelId)
		}
		deviceId = parts[0]
		channelId = parts[1]
	} else {
		deviceId = c.DeviceId
		channelId = c.ChannelId
	}

	var device *fahapi.Device
	if device, err = fahapi.GetDevice(defaultSysAP, deviceId); err != nil {
		return
	}

	var ok bool
	var channel *fahapi.Channel
	if channel, ok = device.Channels[channelId]; !ok {
		return fmt.Errorf("channel %s not found in device %s", channelId, deviceId)
	}

	var json []byte
	if json, err = unmarshall(channel); err != nil {
		return err
	}
	fmt.Println(string(json))

	return nil
}
