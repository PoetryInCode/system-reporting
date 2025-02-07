#!/bin/bash

set -x

INFLUX_HOST=${INFLUX_HOST:-"http://192.168.1.83:8086/write?db=metrics"}
HOST=$(hostname)

go build -ldflags "-X solow.xyz/system-reporting/config.Device=${HOST} -X solow.xyz/system-reporting/config.InfluxHost=${INFLUX_HOST}" ./system-reporting.go
