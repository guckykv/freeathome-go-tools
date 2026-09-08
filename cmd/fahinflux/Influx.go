package main

import (
	"fmt"
	"github.com/guckykv/freeathome-go-fahapi/fahapi"
	"github.com/influxdata/influxdb-client-go/v2"
	"github.com/influxdata/influxdb-client-go/v2/api"
)

var (
	influxClient influxdb2.Client
	writeApi     api.WriteAPI
)

type influxConfiguration struct {
	active bool // should we track all points to influx - default: on
	url    string
	token  string // for influxdb 1.8.x use "username:password" as token
	org    string
	bucket string
}

var influxConfig influxConfiguration

// InitializeInfluxDB opens the connection. Call CloseInfluxDB when done. After
// this, all changes of RTC, WindowSensor and WeatherStation are logged to
// InfluxDB.
func InitializeInfluxDB(url, token, org, bucket string) error {
	// Both of these come back from the server as "bucket not found", which
	// sends the reader looking in the wrong place. Say what is actually missing.
	if url == "" {
		return fmt.Errorf("influx: no InfluxUrl configured")
	}
	if bucket == "" {
		return fmt.Errorf("influx: no bucket configured -- set InfluxBucket (or InfluxDB), " +
			"otherwise every write fails with \"bucket not found\"")
	}

	influxConfig = influxConfiguration{
		active: true,
		url:    url,
		token:  token,
		org:    org,
		bucket: bucket,
	}

	influxClient = influxdb2.NewClientWithOptions(url, token, influxdb2.DefaultOptions().SetBatchSize(50))
	writeApi = influxClient.WriteAPI(org, bucket)

	// InfluxDB 2.x rejects a write without an organisation and reports that as
	// "bucket not found" too. 1.8.x ignores org, so this is a hint, not an error.
	if org == "" {
		logger.Printf("influx: no organisation configured -- fine for InfluxDB 1.8.x, "+
			"but 2.x will answer %q for every write. Set InfluxOrg / INFLUX_ORG.\n", "bucket not found")
	}

	go func() {
		for err := range writeApi.Errors() {
			logger.Printf("influx write error: %s\n", err)
		}
	}()

	logger.Printf("influx: writing to %s, bucket %q, org %q\n", url, bucket, org)
	return nil
}

// CloseInfluxDB flushes what is pending and closes the connection.
func CloseInfluxDB() {
	if influxClient == nil {
		return
	}
	writeApi.Flush()
	influxClient.Close()
	influxClient = nil
}

func WriteData2Influx(keys []string) {
	if !influxConfig.active {
		return
	}

	for _, key := range keys {
		writePoints(fahClient.Unit(key))
	}
	writeApi.Flush()
}

func writePoints(unit fahapi.Unit) bool {
	if unit == nil {
		return false
	}
	point := newInfluxPoint(unit)
	if point != nil {
		writeApi.WritePoint(point)
		return true
	}
	return false
}
