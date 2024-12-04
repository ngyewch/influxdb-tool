# influxdb-tool

InfluxDB tool.

## Report

```
influxdb-tool report --config-file config.yml --output-file report.html
```

### Sample config file

```
---
org: my_org
bucket: my_bucket
startTime: '2024-12-01T00:00:00+08:00'
endTime: '2024-12-02T00:00:00+08:00'
timezone: Asia/Singapore
timeTooltipFormat: yyyy-MM-dd HH:mm
aggregateWindow: 1h
tagOrder:
  - location
  - scientist
  - measurement
  - field
aliases:
  measurement: _measurement
  field: _field
tags:
  measurement: census
timeDisplayFormats:
  year: yyyy
  quarter: yyyy QQQ
  month: yyyy-MM
  day: yyyy-MM-dd
  hour: yyyy-MM-dd HH:mm
queries:
  - tags:
      field: ants
  - tags:
      field: bees
```
