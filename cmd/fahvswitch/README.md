# fahvswitch

## Be a proxy for virtual devices

Small test program which take the deviceId of a virtual "SwitchingActuator"
and copy the input datapoint 0000 to the output datapoint 0000.
So it works like a very simple actuator (lamp).

That's only a simple showcase and everything might change quickly.

Example:

```shell
$ DEVICEID=$(fahcli -d virtual --create abcVAct 7200 "Virtual Test Actuator" SwitchingActuator)
$ fahvswitch $DEVICEID
$ # or use the native id
$ fahvswitch -n abcVAct
```

Creating a virtual device und calling `fahvswitch` with the device id (or `-n` and the native id).

## List all virtual devices

To list all your virtual devices in your system use the `-l` switch.

Example: 
```shell
$ fahvswitch -l
600023AFXXXX.ch0000 abc777   16:38:40:             /                  [AcSwitch]  ein name: OFF
60009B33XXXX.ch0000 abc1234  16:38:40: Etage 2     / Atelier          [SeWindow]  Binärsensor: zu
6000D4FDXXXX.ch0000 abcVAc   16:38:40: Etage 2     / Atelier          [AcSwitch]  Virtual Test Actuator: ON
6000CF03XXXX.ch0000 abc99    16:38:40:             /                  [SeSwitch]  Binary sensor: OFF
```
