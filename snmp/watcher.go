package snmp
// metricd - metric collector for influxdb
// Copyright (C) 2017 Maximilian Pachl

// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.

// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.

// You should have received a copy of the GNU General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.


// --------------------------------------------------------------------------------------
//  imports
// --------------------------------------------------------------------------------------

import (
    "log"
    "time"

    "github.com/faryon93/metricd/config"
    "github.com/faryon93/metricd/util"

    "github.com/influxdata/influxdb/client/v2"
    "github.com/alouca/gosnmp"
)


// --------------------------------------------------------------------------------------
//  public functions
// --------------------------------------------------------------------------------------

func Watcher(influxdb client.Client, snmpConf config.SnmpConf) {
    // esthablish snmp connection
    snmp, err := gosnmp.NewGoSNMP(snmpConf.Host, snmpConf.Community, gosnmp.Version2c, 1)
    if err != nil {
        log.Fatal(err)
    }

    for {
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
            util.GetPairSlice(snmpConf.Tags).Map(),
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