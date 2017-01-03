package main
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
    "flag"
    "os"
    "time"
    "syscall"
    "os/signal"

    "github.com/faryon93/metricd/config"

    "github.com/influxdata/influxdb/client/v2"
    "github.com/alouca/gosnmp"
)


// --------------------------------------------------------------------------------------
//  constants
// --------------------------------------------------------------------------------------

const (
    INFLUX_PRECISION = "ms"
)


// --------------------------------------------------------------------------------------
//  global variables
// --------------------------------------------------------------------------------------

// command line parameters
var (
    configPath string
)


// --------------------------------------------------------------------------------------
//  application entry
// --------------------------------------------------------------------------------------

func main() {
    // setup commandline parser
    flag.StringVar(&configPath, "conf", "/etc/metricd/metricd.conf", "")
    flag.Parse()

	// load the configuration file
    conf, err := config.Load(configPath)
    if err != nil {
        log.Println("failed to load configuration file:", err.Error())
        os.Exit(-1)
    }

    // connect to influxdb
    influx, err := client.NewHTTPClient(client.HTTPConfig{
        Addr: conf.Influx.Address,
        Username: conf.Influx.User,
        Password: conf.Influx.Password,
    })
    if err != nil {
        log.Println("failed to connect to influxdb:", err.Error())
        os.Exit(-1)
    }
    log.Println(influx.Ping(300 * time.Millisecond))

    // setup all measurement watcher
    for name, snmp := range conf.Snmp {
        log.Printf("setting up snmp target \"%s\"", name)
        go watcher(influx, snmp)
    }

    // wait for signals to exit application
    wait(os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
}

// --------------------------------------------------------------------------------------
//  helper functions
// --------------------------------------------------------------------------------------

func watcher(influxdb client.Client, snmpConf config.SnmpConf) {
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
            Precision: INFLUX_PRECISION,
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

func wait(sig ...os.Signal) {
    signals := make(chan os.Signal)
    signal.Notify(signals, sig...)
    <- signals
}