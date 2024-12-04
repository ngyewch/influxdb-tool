FROM influxdb:2

WORKDIR /workspace/influxdb2-sample-data/air-sensor-data
RUN curl -sLO https://github.com/influxdata/influxdb2-sample-data/raw/1867562d51ef281e911e3249b5007b4d2d562114/air-sensor-data/air-sensor-data.lp

COPY /files/ /
