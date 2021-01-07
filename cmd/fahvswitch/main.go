package main

import (
	"bytes"
	"flag"
	"fmt"
	"github.com/guckykv/freeathome-go-fahapi/fahapi"
	"github.com/tkanos/gonfig"
	"log"
	"os"
	"strings"
)

var defaultSysAP = "00000000-0000-0000-0000-000000000000"

type Configuration struct {
	Host     string `env:"FHAPI_HOST"`     // local IP of the SysAP
	Username string `env:"FHAPI_USER"`     // username comes from free@home app
	Password string `env:"FHAPI_PASSWORD"` // pw is the same like you have used in the free@home app
}

var (
	configuration = Configuration{}

	configFile    = flag.String("c", "./.fahapi-config.json", "configuration file")
	deviceIdParam = flag.String("id", "", "deviceIdParam of virtual device")
	verbose       = flag.Bool("v", false, "verbose output")
	quiet         = flag.Bool("q", false, "no output")
	debug         = flag.Bool("d", false, "debug: read all changes from the SysAp but doesn't connect or write to InfluxDB")

	buf      bytes.Buffer
	logger   = log.New(&buf, "", log.LstdFlags)
	logLevel = 1 // 0: quiet / 1: normal / 2: verbose (show also all trigger outs) / 3: debug
)

func main() {
	initialize()
	logLevel = 0 // api shouldn't show any messages
	fahapi.ConfigureApi(configuration.Host, configuration.Username, configuration.Password, handleVSwitchUnit, handleVSwitchMessage, logger, logLevel)

	var device *fahapi.Device
	if *deviceIdParam == "" {
		log.Fatalf("Need virtual DeviceId as parameter\n")
	}
	var err error
	if device, err = fahapi.GetDevice(defaultSysAP, *deviceIdParam); err != nil {
		log.Fatalf("Can't load device with ID %s: %s\n", *deviceIdParam, err)
	}

	logger.Printf("Handle virtual device %s \"%s\" (%s)\n", *device.DisplayName, *device.NativeId, *deviceIdParam)

	fahapi.ReadAndHydradteAllDevices()

	err = fahapi.StartWebSocketLoop(300)
	if err != nil {
		log.Fatal(err)
	}

	os.Exit(0)
}

func handleVSwitchMessage(message fahapi.WebsocketMessage) {
	datapoints := message.ZeroSysAp.Datapoints

	for updDatapoint, value := range datapoints {
		split := strings.Split(updDatapoint, "/")
		if len(split) != 3 {
			logger.Fatalf("illegal message %x: illegal datapoint format %s", message, updDatapoint)
		}
		deviceId := split[0]

		if deviceId == *deviceIdParam {
			channelId := split[1]
			if channelId == "ch0000" {
				datapointId := split[2]
				if datapointId == "idp0000" {
					// proxy: copy input value to output
					setValue(deviceId, channelId, "odp0000", value)
				}
			}
		}
	}
}

func setValue(deviceId, channelId, datapointId, value string) {
	var ok bool
	var err error
	if ok, err = fahapi.PutDatapoint(defaultSysAP, deviceId, channelId, datapointId, value); err != nil {
		logger.Printf("error: %s\n", err)
		return
	}
	if !ok {
		logger.Printf("Can't set datapoint %s.%s.%s to %s\n", deviceId, channelId, datapointId, value)
	}
	logger.Printf("Set %s to new value: %s\n", deviceId, value)
}

func handleVSwitchUnit(unitKeys []string) {
	var deviceKey = *deviceIdParam + ".ch0000" // device.channel of the virtual device (in this easy example the channel is always "ch0000"

	for _, key := range unitKeys {
		if key == deviceKey {
			logger.Printf(fahapi.UnitMap[key].String())
			switchActUnit := fahapi.CastSAU(fahapi.UnitMap[key])
			if !switchActUnit.OnSet {
				// todo refresh virtual device
				logger.Printf("todo refresh\n")
			}
		}
	}
}

func usage() {
	fmt.Printf("usage %s:\n", os.Args[0])
	flag.PrintDefaults()
	fmt.Printf("\n  Example: \"fahvswitch --c ~/.fahapi-config.json --id=6000XXXXXXX\"\n")
	fmt.Printf("  Use:     \"fahinflux --c=\" if you want to skip the configfile and use env vars only\n")
	fmt.Printf("  Configuration file needs the following fields:" + `
	Host        or as env: "FHAPI_HOST"     // local IP of the SysAP
	Username    or as env: "FHAPI_USER"     // username comes from free@home app: a3XXXXXX-XXXX-XXXX-XXXX-XXXXXXXXXXb9
	Password    or as env: "FHAPI_PASSWORD" // pw is the same like you have used in your free@home app
`)
}

func initialize() {
	flag.Usage = usage
	flag.Parse()

	err := gonfig.GetConf(*configFile, &configuration)
	if err != nil {
		log.Fatal("GetConfig: " + err.Error())
	}

	logger.SetOutput(os.Stdout)

	if *quiet {
		logLevel = 0
	} else if *verbose {
		logLevel = 2
	}
	if *debug {
		logLevel = 3
	}
}
