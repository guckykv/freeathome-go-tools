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

	Device  DevCmd  `cmd help:"List devices or show one device."`
	Channel ChanCmd `cmd help:"Show channel of one device."`
	Virtual VirtCmd `cmd help:"Virtual device handling"`
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
		kong.Description("A command line tool for reading and writing to the free@home local API"),
		kong.UsageOnError(),
		kong.ConfigureHelp(kong.HelpOptions{
			Compact: false,
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

	fahapi.ConfigureApi(configuration.Host, configuration.Username, configuration.Password, nil, nil, logger, logLevel)
	return nil
}

func unmarshall(v interface{}) ([]byte, error) {
	if cli.Globals.Pretty {
		return json2.MarshalIndent(v, "", "  ")
	} else {
		return json2.Marshal(v)
	}
}
