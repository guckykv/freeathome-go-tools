package main

import (
	"fmt"
	"github.com/guckykv/freeathome-go-fahapi/fahapi"
	"github.com/influxdata/influxdb-client-go/v2"
	"github.com/influxdata/influxdb-client-go/v2/api"
)

var writeApi api.WriteAPI

type influxConfiguration struct {
	active bool // should we track all points to influx - default: on
	url    string
	token  string // for influxdb 1.8.x use "username:password" as token
	org    string
	bucket string
}

var influxConfig influxConfiguration

// if you call this function, all changes of RTC, WindowSensor, and WeatherStation will be logged to influxdb
func InitializeInfluxDB(url, token, org, bucket string) {
	influxConfig = influxConfiguration{
		active: true,
		url:    url,
		token:  token,
		org:    org,
		bucket: bucket,
	}
}

func WriteData2Influx(keys []string) {
	if influxConfig.active {
		count := 0
		client := openInflux()
		for _, key := range keys {
			if writePoints(fahapi.UnitMap[key]) {
				count++
			}
		}
		flushAndCloseInflux(client)
		//fmt.Printf("%d influx points written\n", count)
	}
}

func writePoints(unit fahapi.Unit) bool {
	point := newInfluxPoint(unit)
	if point != nil {
		writeApi.WritePoint(point)
		return true
	}
	return false
}

func openInflux() influxdb2.Client {
	client := influxdb2.NewClientWithOptions(influxConfig.url, influxConfig.token, influxdb2.DefaultOptions().SetBatchSize(50))
	writeApi = client.WriteAPI(influxConfig.org, influxConfig.bucket)
	errorsCh := writeApi.Errors()
	go func() {
		for err := range errorsCh {
			fmt.Printf("influx write error: %s\n", err.Error())
		}
	}()
	return client
}

func flushAndCloseInflux(client influxdb2.Client) {
	writeApi.Flush()
	client.Close()
}
