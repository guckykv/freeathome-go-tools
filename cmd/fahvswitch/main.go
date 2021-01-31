package main

import (
	"bytes"
	"flag"
	"fmt"
	"github.com/guckykv/freeathome-go-fahapi/fahapi"
	"github.com/tkanos/gonfig"
	"log"
	"os"
	"os/user"
	"path/filepath"
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

	configFile   = flag.String("c", "~/.fahapi-config.json", "configuration file")
	listVirtuals = flag.Bool("l", false, "list all devices with native id (virtuals)")
	useNative    = flag.Bool("n", false, "arguments are the native IDs, not the virtual internal free@home ids")
	verbose      = flag.Bool("v", false, "verbose output")
	quiet        = flag.Bool("q", false, "no output")
	debug        = flag.Bool("d", false, "debug: even more logging")

	buf      bytes.Buffer
	logger   = log.New(&buf, "", log.LstdFlags)
	logLevel = 1 // 0: quiet / 1: normal / 2: verbose (show also all trigger outs) / 3: debug

	refreshTime   = 300
	virtualIdList []string
)

func main() {
	initialize()
	logLevel = 0 // api shouldn't show any messages
	fahapi.ConfigureApi(configuration.Host, configuration.Username, configuration.Password, handleVSwitchUnit, handleVSwitchMessage, logger, logLevel)

	if len(flag.Args()) == 0 && !*listVirtuals {
		log.Fatalf("Need virtual DeviceId as parameter\n")
	}

	fahapi.ReadAndHydradteAllDevices()

	if *listVirtuals {
		for _, unit := range fahapi.UnitMap {
			if unit.GetUnitData().NativeId != nil {
				fmt.Println(unit.String())
			}
		}
		os.Exit(0)
	}

	vidList := handleArgs(flag.Args())
	virtualIdList = filterType(vidList, fahapi.UntTypeSwitchActuator)

	err := fahapi.StartWebSocketLoop(refreshTime)
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

		if _, ok := inArray(virtualIdList, &deviceId); ok {
			channelId := split[1]
			if channelId == "ch0000" {
				datapointId := split[2]
				if datapointId == "idp0000" {
					// proxy: copy input value to output
					setValueInSysAP(deviceId, channelId, "odp0000", value)
				}
			}
		}
	}
}

func handleVSwitchUnit(unitKeys []string) {
	// currently only used for refreshing devices. For updating the SysAP the method "handleVSwitchMessage" is used.

	for _, vid := range virtualIdList {
		deviceKey := vid + ".ch0000" // device.channel of the virtual device (in this easy example the channel is always "ch0000"

		for _, key := range unitKeys {
			if key == deviceKey {
				logger.Printf(fahapi.UnitMap[key].String())
				switchActUnit := fahapi.CastSAU(fahapi.UnitMap[key])

				if !switchActUnit.OnSet {
					// if I get a message for this unit, but the state of "On" hasn't changed (OnSet==false)
					// than this is the periodic refresh al call. I use this to refresh the virtual device at the SysAP.

					/*
						 * at the moment the free@home API always return (400 Bad Request) for this update PUT.
						 * so this refresh is disabled

						nativeId := *switchActUnit.NativeId
						var reqBody *fahapi.VirtualDevice
						reqBody = &fahapi.VirtualDevice{
							Properties: fahapi.VirtualDeviceProperties{
								Ttl: strconv.Itoa(refreshTime + 10),
							},
						}

						if virtualId, err := fahapi.PutVirtualDevice(defaultSysAP, nativeId, reqBody); err != nil {
							if virtualId != vid {
								logger.Fatalf("Refresh device %s (%s): Getting back wrong id %s: %s\n", vid, nativeId, virtualId, err)
							}
							logger.Printf("Refresh device %s (%s) failed: %s\n", virtualId, nativeId, err)
						} else {
							if logLevel > 1 {
								logger.Printf("Refresh device %s (%s) done\n", virtualId, nativeId)
							}
						}
					*/
				}
			}
		}
	}
}

func setValueInSysAP(deviceId, channelId, datapointId, value string) {
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

func inArray(slice []string, val *string) (int, bool) {
	if val == nil {
		return -1, false
	}
	for i, item := range slice {
		if item == *val {
			return i, true
		}
	}
	return -1, false
}

func usage() {
	fmt.Printf("usage %s: VIRTID\n", os.Args[0])
	flag.PrintDefaults()
	fmt.Printf("\n  Example: \"fahvswitch --c ./.fahapi-config.json 6000XXXXXX1 6000XXXXXX2\"\n")
	fmt.Printf("           Add as many virtual IDs as you want as arguments to the command.\n")
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

	if strings.HasPrefix(*configFile, "~/") {
		usr, _ := user.Current()
		*configFile = filepath.Join(usr.HomeDir, (*configFile)[2:])
	}

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

func handleArgs(argumentList []string) (vidList []string) {
	var device *fahapi.Device

	if *useNative {
		for _, unit := range fahapi.UnitMap {
			if i, ok := inArray(argumentList, unit.GetUnitData().NativeId); ok {
				vidList = append(vidList, unit.GetUnitData().SerialNumber)
				argumentList = append(argumentList[:i], argumentList[i+1:]...)
			}
		}
		if len(argumentList) > 0 {
			log.Fatalf("Can't find all native devices: %v\n", argumentList)
		}
	}

	var err error

	for _, vid := range vidList {
		if device, err = fahapi.GetDevice(defaultSysAP, vid); err != nil {
			log.Fatalf("Can't load device with ID %s: %s\n", vid, err)
		} else {
			logger.Printf("Handle virtual device %s \"%s\" (%s)\n", *device.DisplayName, *device.NativeId, vid)
		}
	}

	return vidList
}

func filterType(vidList []string, allowedType fahapi.UnitTypeConst) []string {
	var outList []string

	for _, vid := range vidList {
		deviceKey := vid + ".ch0000" // device.channel of the virtual device (in this easy example the channel is always "ch0000"
		unitdData := fahapi.UnitMap[deviceKey].GetUnitData()
		if unitdData.Type == allowedType {
			outList = append(outList, vid)
		} else {
			logger.Printf("Skip virtual device %s (%s). Illegal type: %s", vid, *unitdData.NativeId, unitdData.Type)
		}
	}

	return outList
}
