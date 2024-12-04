#!/bin/bash

set -e

cd /workspace/influxdb2-sample-data

influx bucket create -n air-sensor
influx write -b air-sensor -f air-sensor-data/air-sensor-data.lp
