# metricd - metric collector for influxdb

## sample configuration

```
[influx]
address="http://192.168.0.32:8086"
user="user"
password="password"

[snmp.if_stats]
host = "1921.68.0.1"
community = "public"
database = "router"
measurement = "if_stats"
sample_time = 1000
tags = [
    "network:adsl-uplink"
]
datapoints = [
	".1.3.6.1.2.1.31.1.1.1.10.1:tx_bytes",
	".1.3.6.1.2.1.31.1.1.1.6.1:rx_bytes",
]
```