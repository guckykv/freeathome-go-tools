# fahvswitch

Small test program which take the deviceId of a virtual "SwitchingActuator"
and copy the input datapoint 0000 to the output datapoint 0000.
So it works like a very simple actuator (lamp).

That's only a simple showcase and everything might change quickly.

Example:

```shell
$ DEVICEID=$(fahcli -d virtual --create abcVAct 7200 "Virtual Test Actuator" SwitchingActuator)
$ fahvswitch --id=$DEVICEID
```

Creating a virtual device und calling `fahvswitch` with the device id.
