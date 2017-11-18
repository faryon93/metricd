# metricd - metric collector for influxdb

This is a simple metric collector service which was built for me personal needs.
Only InfluxDB is supported as a data storage backend. The following modules are
available for metric aggregation:

## Mikrotik RouterOS Accounting
This modules uses the accounting functionality of Mikrotiks RouterOS. It provides per client traffic statistics.
The clients ip adddress is stored as the tag *host*.

```
[accouting.test]
database = "db"                 # influx database name
measurement = "measurement"     # influx measurement name
sample_time = 1000              # sample time in ms
host = "192.168.178.1"          # router os host ip address
network = "192.168.178.0/24"    # network to obtain traffic stats from
```

## MySQL
The MySQL module can be used to execute arbitrary SQL queries against an mysql server and store the results.
The queries are executed every `sample_time` milliseconds.

```
[mysql.test]
database = "database"                                   # influx database name
measurement = "measurement"                             # influx measurement name
sample_time = 60000                                     # sample time in ms
sql_uri = "user:pass@tcp(127.0.0.1:3306)/database"      # database uri
queries = [                                             # query with value name
        "val1:SELECT AVG(time) FROM table",
        "val2:SELECT count(*) FROM table",
]
tags = ["foo:bar"]                                      # static tags
```

## Syslog
The syslog module creates an RFC3164 compatible syslog server and parses the log messages with a standard golang regex.
Named capture groups with the prefix *tag_* are stored as tags and capture groups with the prefix *val_* are parsed as float numbers and stored as point values.

```
[syslog.test]
database = "database"                   # influxdb database name
measurement = "measurement"             # influxdb measurement name
listen = "0.0.0.0:514"                  # listen address for syslog server
regex = "(key:\\s+(?P<val_value>.+)"    # regex to parse log message
tags = ["foo:bar"]                      # static tags
```

## SNMP
Cyclically acquires information from an SNMP target.

```
[snmp.test]
host = "1921.68.178.1"
community = "public"
database = "database"
measurement = "measurement"
sample_time = 1000
tags = ["foo:bar"]
datapoints = [
	".1.3.6.1.2.1.31.1.1.1.10.1:tx_bytes",
	".1.3.6.1.2.1.31.1.1.1.6.1:rx_bytes",
]
```