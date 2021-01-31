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

type Configuration struct {
	Host        string `env:"FHAPI_HOST"`     // local IP of the SysAP
	Username    string `env:"FHAPI_USER"`     // username comes from free@home app
	Password    string `env:"FHAPI_PASSWORD"` // pw is the same like you have used in the free@home app
	InfluxUrl   string `env:"INFLUX_URL"`     // complete url with schema, host, and port
	InfluxDB    string `env:"INFLUX_DB"`      // complete url with schema, host, and port
	InfluxToken string `env:"INFLUX_TOKEN"`   // at influxdb 1.8.x this can be "username:password"
}

var (
	configuration = Configuration{}

	configFile  = flag.String("c", "~/.fahapi-config.json", "configuration file")
	noWebsocket = flag.Bool("n", false, "no websocket connection; read and update data only once and quit")
	verbose     = flag.Bool("v", false, "verbose output")
	quiet       = flag.Bool("q", false, "no output")
	debug       = flag.Bool("d", false, "debug: read all changes from the SysAp but doesn't connect or write to InfluxDB")

	buf      bytes.Buffer
	logger   = log.New(&buf, "", log.LstdFlags)
	logLevel = 1 // 0: quiet / 1: normal / 2: verbose (show also all trigger outs) / 3: debug
)

func main() {
	initialize()

	websocketCallback := WriteData2Influx
	if *debug {
		websocketCallback = nil
	}

	fahapi.ConfigureApi(configuration.Host, configuration.Username, configuration.Password, websocketCallback, nil, logger, logLevel)

	if !*debug {
		InitializeInfluxDB(configuration.InfluxUrl, configuration.InfluxToken, "", configuration.InfluxDB)
	}

	fahapi.ReadAndHydradteAllDevices()

	if !*noWebsocket {
		err := fahapi.StartWebSocketLoop(300)
		if err != nil {
			log.Fatal(err)
		}
	}

	os.Exit(0)
}

func usage() {
	fmt.Printf("usage %s:\n", os.Args[0])
	flag.PrintDefaults()
	fmt.Printf("\n  Example: \"fahinflux --c ~/.fahapi-config.json\"\n")
	fmt.Printf("  Use:     \"fahinflux --c=\" if you want to skip the configfile and use env vars only\n")
	fmt.Printf("  Configuration file needs the following fields:" + `
	Host        or as env: "FHAPI_HOST"     // local IP of the SysAP
	Username    or as env: "FHAPI_USER"     // username comes from free@home app: a3XXXXXX-XXXX-XXXX-XXXX-XXXXXXXXXXb9
	Password    or as env: "FHAPI_PASSWORD" // pw is the same like you have used in your free@home app
	InfluxUrl   or as env: "INFLUX_URL"     // complete url with schema, host, and port
	InfluxDB    or as env: "INFLUX_DB"      // database (bucket) name
	InfluxToken or as env: "INFLUX_TOKEN"   // at influxdb 1.8.x this can be "username:password"
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
