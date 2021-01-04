package main

import (
	"bytes"
	json2 "encoding/json"
	"fmt"
	"github.com/alecthomas/kong"
	"github.com/guckykv/freeathome-go-fahapi/fahapi"
	"github.com/tkanos/gonfig"
	"log"
	"os"
	"strings"
)

type Globals struct {
	Debug      bool   `help:"Enable debug mode." short:"d"`
	Configfile string `type:"path" help:"Specify configuration file." short:"c" default:"~/.fahapi-config.json"`
	// Format     string      `short:"f" help:"Output format. json or text" default:"json"`
	Pretty  bool        `short:"p" help:"Pretty format for json" default:"false"`
	Version VersionFlag `name:"version" help:"Print version information and quit"`
}

type CLI struct {
	Globals

	Device  DevCmd  `cmd help:"LIst devices or show one device."`
	Channel ChanCmd `cmd help:"List channels of one device."`
	Get     GetCmd  `cmd help:"Get the value of a datapoint."`
	Set     SetCmd  `cmd help:"Set the value of a datapoint."`
}

type VersionFlag string

func (v VersionFlag) Decode(ctx *kong.DecodeContext) error { return nil }
func (v VersionFlag) IsBool() bool                         { return true }
func (v VersionFlag) BeforeApply(app *kong.Kong, vars kong.Vars) error {
	fmt.Println(vars["version"])
	app.Exit(0)
	return nil
}

// --------------------------------------

var defaultSysAP = "00000000-0000-0000-0000-000000000000"

var globals = Globals{
	Version: VersionFlag("0.0.1"),
}

type Configuration struct {
	Host     string `env:"FHAPI_HOST"`     // local IP of the SysAP
	Username string `env:"FHAPI_USER"`     // username comes from free@home app
	Password string `env:"FHAPI_PASSWORD"` // pw is the same like you have used in the free@home app
}

var configuration = Configuration{}
var cli CLI

func main() {
	cli = CLI{
		Globals: globals,
	}

	ctx := kong.Parse(&cli,
		kong.Name("fahcli"),
		kong.Description("A shell-like example app."),
		kong.UsageOnError(),
		kong.ConfigureHelp(kong.HelpOptions{
			Compact: true,
			Summary: true,
		}),
		kong.Vars{
			"version": "0.0.1",
		},
	)
	err := ctx.Run(&cli.Globals)
	ctx.FatalIfErrorf(err)
}

func initializeApi(configfile string) error {
	err := gonfig.GetConf(configfile, &configuration)
	if err != nil {
		log.Fatal("GetConfig: " + err.Error())
		return err
	}

	var (
		buf      bytes.Buffer
		logger   = log.New(&buf, "", log.LstdFlags)
		logLevel = 1 // 0: quiet / 1: normal / 2: verbose (show also all trigger outs)
	)
	logger.SetOutput(os.Stdout)

	fahapi.ConfigureApi(configuration.Host, configuration.Username, configuration.Password, nil, logger, logLevel)
	return nil
}

func unmarshall(v interface{}) ([]byte, error) {
	if cli.Globals.Pretty {
		return json2.MarshalIndent(v, "", "  ")
	} else {
		return json2.Marshal(v)
	}
}

// DEVICE --------------------------------------

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
			log.Printf("Error unmarshall object: %v", err)
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

// CHANNEL --------------------------------------

type ChanCmd struct {
	DeviceId  string `arg required help:"DeviceId or DeviceId.ChannelId"`
	ChannelId string `arg optional help:"Channel Id"`
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

// GET --------------------------------------

type GetCmd struct {
	Paths []string `arg name:"path" help:"Paths of datapoints: <DeviceId>.<ChannelId>.<Datapoint>"`
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
