package snmp

import (
    "log"
    "time"

    "github.com/faryon93/metricd/config"

    "github.com/influxdata/influxdb/client/v2"
    "github.com/alouca/gosnmp"
)

func Watcher(influxdb client.Client, snmpConf config.SnmpConf) {
    for {
        // esthablish snmp connection
        snmp, err := gosnmp.NewGoSNMP(snmpConf.Host, snmpConf.Community, gosnmp.Version2c, 1)
        if err != nil {
            log.Fatal(err)
        }

        // gather all datapoints
        values := make(map[string]interface{})
        for _, oid := range snmpConf.DataPoints {
            // query the snmp oid for its values
            resp, err := snmp.Get(oid.Key())
            if err != nil {
                log.Println("failed to fetch snmp oid:", err.Error())
                continue
            }

            // make sure the response is not empty
            if len(resp.Variables) <= 0 {
                log.Println("empty response for oid", oid.Key())
                continue
            }

            values[oid.Val()] = int(resp.Variables[0].Value.(uint64))
        }

        bp, _ := client.NewBatchPoints(client.BatchPointsConfig{
            Database:  snmpConf.Database,
            Precision: "ms",
        })

        // construct the new datapoint
        pt, _ := client.NewPoint(
            snmpConf.Measurement,
            config.GetPairSlice(snmpConf.Tags).Map(),
            values,
            time.Now(),
        )
        bp.AddPoint(pt)

        // write the datapoints
        err = influxdb.Write(bp)
        if err != nil {
            log.Println("failed to write datapoint:", err.Error())
        }

        // sleep until next execution
        time.Sleep(time.Duration(snmpConf.SampleTime) * time.Millisecond)
    }
}