# fahcli

---

*VERY early version of a shell program to call the free@home API.*

That's more like a technical preview what might possible.
Everything is subject to change.

---

## fahapi

This tool is build on top of the package [guckykv/freeathome-go-fahapi](https://github.com/guckykv/freeathome-go-fahapi).


## Current features

* List all DeviceIDs: `fahcli device`
* Show the structure of one given DeviceId: `fahcli device <DeviceId>`
* List all channels of one device: `fahcli channel <DeviceId>`
* Show one channel of a device: `fahcli channel <DeviceId>.<ChannelId>` or `fahcli channel <DeviceId> <ChannelId>`
* Getting the value of one (or more) datapoint(s): `fahcli get <DeviceId>.<ChannelId>.<Datapoint>`
* Setting the value of one (or more) datapoint(s): `fahcli set <DeviceId>.<ChannelId>.<Datapoint>=<Value>`
* Creating virtual devices: `fahcli virtual <serial> <ttl> <name> <type>`

**Examples:**

Switch a light on:

```shell
$ fahcli -c ~/.fahapi-config.json set ABB2XXXXXXX1.ch0011.idp0000=1
```
If you put the config file in your homedir, you can ommit the `-c` option.

See the status of the switch:

```shell
$ fahcli set ABB2XXXXXXX1.ch0011.idp0000
```

Create a new **virtual device**:

```shell
$ fahcli virtual abc123 300 "ein Name" BinarySensor
```

As result you get the DeviceId. Now you can view the device...
```shell
$ fahcli device <DeviceID>
```
Add option `-p` to get a pretty printed JSON result.

Your device should have only one ChannelID `ch0000`.
Look for the output DatapointID with pairingID `1`. 
Probably is that `odp0000`.

Then you can read the state of the sensor:
```shell
$ fahcli get <DeviceID>.<ChannelID>.<DatapointID>
```

And change the value via the `set` command:
```shell
$ fahcli set 6000CF034D2A.ch0000.odp0000=0
```

If you start [fahinflux](../fahinflux) in another window, you will
see the updates you trigger with `fahcli` in the output stream of `fahinflux`.

Call of fahcli:
```shell
$  cmd/fahcli/fahcli set 6000CF034D2A.ch0000.odp0000=1
$  cmd/fahcli/fahcli set 6000CF034D2A.ch0000.odp0000=0
```
In parallel `fahinflux -d` window:
```shell
2021/01/05 20:40:15 6000CF034D2A.ch0000 abc123   20:40:15:             /                  [SeSwitch]  Binary sensor: ON
2021/01/05 20:40:18 6000CF034D2A.ch0000 abc123   20:40:18:             /                  [SeSwitch]  Binary sensor: OFF
```

`abc123` was the serial while creating the virtual device.


To see more options try
```shell
$ fahcli --help`
```


## Configuration

Setup a Json file with the following structure:

```json
{
  "Host": "192.168.XX.YY",
  "Username": "a3XXXXXX-XXXX-XXXX-XXXX-XXXXXXXXXXb9",
  "Password": "XXXXXXXXX"
}
```

You can use the same config file like for `fahinflux`. It's no problem that there are unneded field for this command.

You can also use environment variables to configure these values (`FHAPI_HOST`, `FHAPI_USER`, `FHAPI_PASSWORD`).
