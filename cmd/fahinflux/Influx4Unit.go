package main

import (
	"time"

	"github.com/guckykv/freeathome-go-fahapi/fahapi"
	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	"github.com/influxdata/influxdb-client-go/v2/api/write"
)

func newInfluxPoint(unit fahapi.Unit) *write.Point {
	switch unit.GetUnitData().Type {
	case fahapi.UntTypeWeatherStationBrightness:
		return newInfluxPointWSB(fahapi.CastWSB(unit))
	case fahapi.UntTypeWeatherStationWind:
		return newInfluxPointWSW(fahapi.CastWSW(unit))
	case fahapi.UntTypeWeatherStationRain:
		return newInfluxPointWSR(fahapi.CastWSR(unit))
	case fahapi.UntTypeWeatherStationTemperature:
		return newInfluxPointWST(fahapi.CastWST(unit))
	case fahapi.UntTypeRoomTemperatureController:
		return newInfluxPointRTC(fahapi.CastRTC(unit))
	case fahapi.UntTypeWindowDoorSensor:
		return newInfluxPointWDS(fahapi.CastWDS(unit))
	case fahapi.UntTypeBlindActuator:
		return newInfluxPointBAU(fahapi.CastBAU(unit))
	case fahapi.UntTypeShutterActuator:
		return newInfluxPointSHAU(fahapi.CastSHAU(unit))
	case fahapi.UntTypeAwningActuator:
		return newInfluxPointAAU(fahapi.CastAAU(unit))
	}
	return nil
}

/*
 * Don't use the "LastUpdate" field of UnitData for the influxDB timestamp.
 * Use the current time instead. Reason:
 * 1) Normaly we create the point in the same moment the update happens (so current time is equal to LastUpdate
 * 2) In case of full update (we want that all all points are created at least every X minutes) we have
 *    to use the current time and not the LastUpdate (because there was no real update of the values)
 */

func newInfluxPointBAU(bau *fahapi.BlindActuatorUnit) *write.Point {
	tags := map[string]string{
		"floor": bau.Floor,
		"room":  bau.Room,
	}
	fields := map[string]interface{}{
		"absolutePosBlinds": bau.AbsolutePosBlinds,
	}
	point := influxdb2.NewPoint(
		"blinds",
		tags,
		fields,
		time.Now(),
	)
	return point
}

func newInfluxPointSHAU(shau *fahapi.ShutterActuatorUnit) *write.Point {
	tags := map[string]string{
		"floor": shau.Floor,
		"room":  shau.Room,
	}
	fields := map[string]interface{}{
		"absolutePosBlinds": shau.AbsolutePosBlinds,
		"absolutePosSlats": shau.AbsolutePosSlats,
	}
	point := influxdb2.NewPoint(
		"blinds",
		tags,
		fields,
		time.Now(),
	)
	return point
}

func newInfluxPointAAU(aau *fahapi.AwningActuatorUnit) *write.Point {
	tags := map[string]string{
		"floor": aau.Floor,
		"room":  aau.Room,
	}
	fields := map[string]interface{}{
		"absolutePosBlinds": aau.AbsolutePosBlinds,
	}
	point := influxdb2.NewPoint(
		"blinds",
		tags,
		fields,
		time.Now(),
	)
	return point
}

func newInfluxPointRTC(rtc *fahapi.RoomTemperatureControllerUnit) *write.Point {
	tags := map[string]string{
		"floor": rtc.Floor,
		"room":  rtc.Room,
	}
	fields := map[string]interface{}{
		"actualDegree": rtc.ActualDegree,
		"targetDegree": rtc.TargetDegree,
		"active":       rtc.Active,
		"capacity":     rtc.Capacity,
	}
	point := influxdb2.NewPoint(
		"rtr",
		tags,
		fields,
		time.Now(),
	)
	return point
}

func newInfluxPointWSB(ws *fahapi.WeatherStationBrightnessUnit) *write.Point {
	alarm := 0
	if ws.LuminanceAlarm {
		alarm = 1
	}
	fields := map[string]interface{}{
		"luminance":      ws.Luminance,
		"luminanceAlarm": alarm,
	}
	tags := map[string]string{}
	point := influxdb2.NewPoint(
		"weather",
		tags,
		fields,
		time.Now(),
	)
	return point
}

func newInfluxPointWSR(ws *fahapi.WeatherStationRainUnit) *write.Point {
	rain := 0
	if ws.RainAlarm {
		rain = 1
	}
	fields := map[string]interface{}{
		"rain":           rain,
		"rainPercentage": ws.RainPercentage,
	}
	tags := map[string]string{}
	point := influxdb2.NewPoint(
		"weather",
		tags,
		fields,
		time.Now(),
	)
	return point
}

func newInfluxPointWST(ws *fahapi.WeatherStationTemperatureUnit) *write.Point {
	freeze := 0
	if ws.FreezeAlarm {
		freeze = 1
	}
	fields := map[string]interface{}{
		"temperature": ws.Temperature,
		"freezeAlarm": freeze,
	}
	tags := map[string]string{}
	point := influxdb2.NewPoint(
		"weather",
		tags,
		fields,
		time.Now(),
	)
	return point
}

func newInfluxPointWSW(ws *fahapi.WeatherStationWindUnit) *write.Point {
	alarm := 0
	if ws.WindAlarm {
		alarm = 1
	}
	fields := map[string]interface{}{
		"wind":      ws.Wind,
		"windAlarm": alarm,
	}
	tags := map[string]string{}
	point := influxdb2.NewPoint(
		"weather",
		tags,
		fields,
		time.Now(),
	)
	return point
}

func newInfluxPointWDS(wds *fahapi.WindowDoorSensorUnit) *write.Point {
	tags := map[string]string{
		"floor": wds.Floor,
		"room":  wds.Room,
		"name":  *wds.GetChannel().DisplayName,
	}
	open := 0
	if wds.Open {
		open = 1
	}
	fields := map[string]interface{}{
		"state": open,
	}
	point := influxdb2.NewPoint(
		"bc",
		tags,
		fields,
		time.Now(),
	)
	return point
}
