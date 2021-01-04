# fahinflux

Connect to SysAp, get all changes for specific channel types and send them to an influxDB.
So you can use grafana to draw them.

## fahapi

This tool is build on top of the package [guckykv/freeathome-go-fahapi](https://github.com/guckykv/freeathome-go-fahapi).


## Configuration

For a short introduction call `fahinflux -h` so you get a list of
the config variables and a short description.

You can store them into a json file, or set the corresponding environment variables.

Example configuration:

```json
{
  "Host": "192.168.XX.YY",
  "Username": "a3XXXXXX-XXXX-XXXX-XXXX-XXXXXXXXXXb9",
  "Password": "XXXXXXXXX",
  "InfluxUrl": "http://192.168.XX.YY:8086",
  "InfluxDB": "smarthome"
}
```

## Options

If you start `fahinflux` without any further option it will establish
a websocket connection to the SysAP and runs forever (till you kill it).

If you only want to read and write the state once to influx, use option `-n`.
Than the command will read in the whole configuration, write the data to influxDB and then quits.

### Output Control

Normaly the programm send every change it gets from the SysAp to stdout.
With `-q` you stop that.
With `-v` you also get every 5 minutes the complete state of all (configured) devices.

If you send a SIGHUP signal (`kill -1 PID`) to the running command, you force it
to print out the current device state.

## InfluxDB

All room temperature controller messages go to table `rtc`.

The window or door sensors go to `bc`.

And the weather station go to `weather`.
