package main

import (
	"bytes"
	"flag"
	"fmt"
	"github.com/tkanos/gonfig"
	heosapi "github.com/xaxes/heos-api"
	"log"
	"os"
	"os/user"
	"path/filepath"
	"strings"
)

type Configuration struct {
	HeosHost     string `env:"HEOS_HOST"`     // local IP of the Receiver
	HeosPort     string `env:"HEOS_PORT"`     // port of the receiver interface, normally 1255
}

var (
	configuration = Configuration{}

	configFile   = flag.String("c", "~/.fahapi-config.json", "configuration file")
	verbose      = flag.Bool("v", false, "verbose output")
	quiet        = flag.Bool("q", false, "no output")
	debug        = flag.Bool("d", false, "debug: even more logging")

	buf      bytes.Buffer
	logger   = log.New(&buf, "", log.LstdFlags)
	logLevel = 1 // 0: quiet / 1: normal / 2: verbose (show also all trigger outs) / 3: debug

)

func main() {
	initialize()
	example()
}

func example() {
	heos := heosapi.NewHeos(configuration.HeosHost + ":" + configuration.HeosPort)

	if err := heos.Connect(); err != nil {
		fmt.Printf("connect: %s\n", err)
		os.Exit(1)
	}
	defer func() {
		if err := heos.Disconnect(); err != nil {
			fmt.Printf("disconnect: %s\n", err)
			os.Exit(1)
		}
	}()

	resp, err := heos.Send(heosapi.Command{
		Group:   "system",
		Command: "heart_beat",
	}, map[string]string{})
	if err != nil {
		fmt.Printf("send: %s\n", err)
		os.Exit(1)
	}

	fmt.Printf("Response: %+v\n", resp)
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
