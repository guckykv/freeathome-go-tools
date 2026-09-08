package main

import (
	"bytes"
	"context"
	"flag"
	"fmt"
	"github.com/guckykv/freeathome-go-fahapi/fahapi"
	"github.com/tkanos/gonfig"
	"log"
	"os"
	"os/signal"
	"os/user"
	"path/filepath"
	"strings"
	"syscall"
)

type Configuration struct {
	Host        string `env:"FHAPI_HOST"`     // local IP of the SysAP
	Username    string `env:"FHAPI_USER"`     // username comes from free@home app
	Password    string `env:"FHAPI_PASSWORD"` // pw is the same like you have used in the free@home app
	InfluxUrl   string `env:"INFLUX_URL"`     // complete url with schema, host, and port
	InfluxDB    string `env:"INFLUX_DB"`      // database (1.8.x) or bucket (2.x) name
	InfluxToken string `env:"INFLUX_TOKEN"`   // at influxdb 1.8.x this can be "username:password"
	InfluxOrg   string `env:"INFLUX_ORG"`     // required by influxdb 2.x; leave empty for 1.8.x
}

var (
	configuration = Configuration{}

	configFile  = flag.String("c", "~/.fahapi-config.json", "configuration file")
	noWebsocket = flag.Bool("n", false, "no websocket connection; read and update data only once and quit")
	verbose     = flag.Bool("v", false, "verbose output")
	quiet       = flag.Bool("q", false, "no output")
	debug       = flag.Bool("d", false, "debug: read all changes from the SysAp but doesn't connect or write to InfluxDB")

	fahClient *fahapi.Client

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

	fahClient = fahapi.New(fahapi.Config{
		Host:         configuration.Host,
		Username:     configuration.Username,
		Password:     configuration.Password,
		UnitCallback: websocketCallback,
		Logger:       logger,
		LogLevel:     logLevel,
	})

	if !*debug {
		InitializeInfluxDB(configuration.InfluxUrl, configuration.InfluxToken, configuration.InfluxOrg, configuration.InfluxDB)
		defer CloseInfluxDB()
	}

	if err := fahClient.ReadAndHydrateAllDevices(); err != nil {
		log.Fatal(err)
	}

	if !*noWebsocket {
		// Signal handling belongs to the application, not to the library.
		ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
		defer stop()

		hangup := make(chan os.Signal, 1)
		signal.Notify(hangup, syscall.SIGHUP)
		go func() {
			for range hangup {
				fahClient.TreatAllUnitsAsUpdated(true)
			}
		}()

		if err := fahClient.StartWebSocketLoop(ctx, 300); err != nil {
			log.Fatal(err)
		}
	}
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
	InfluxDB    or as env: "INFLUX_DB"      // database (1.8.x) or bucket (2.x) name
	InfluxToken or as env: "INFLUX_TOKEN"   // at influxdb 1.8.x this can be "username:password"
	InfluxOrg   or as env: "INFLUX_ORG"     // required by influxdb 2.x; leave empty for 1.8.x
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
