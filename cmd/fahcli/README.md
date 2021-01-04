# fahcli

---

*VERY early version of a shell programm to call the free@home API.*

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

**Example:**

Switch a light on:

`fahcli -c ~/.fahapi-config.json set ABB2XXXXXXX1.ch0011.idp0000=1`

To see more options try `fahcli -h`

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
