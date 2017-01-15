package accounting
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
    "time"
    "log"
    "net"
    "bufio"
    "strings"

    "github.com/faryon93/metricd/config"
    "github.com/faryon93/metricd/util"

    "github.com/influxdata/influxdb/client/v2"
)


// --------------------------------------------------------------------------------------
//  constants
// --------------------------------------------------------------------------------------

const (
    SAMPLE_TIME = 1000

    HOST_TAG = "host"
    TX_BYTES_FIELD = "tx_bytes"
    RX_BYTES_FIELD = "rx_bytes"
)


// --------------------------------------------------------------------------------------
//  public functions
// --------------------------------------------------------------------------------------

func Watcher(influxdb client.Client, conf config.AccoutingConf) {
    _, network, err := net.ParseCIDR(conf.Network)
    if err != nil {
        log.Println("[Accouting] failed to parse net:", err.Error())
        return
    }

    // Find all already present host keys, so datapoints are
    // create even when the host does not issue any traffic.
    // We need to fill in zero values if no traffic is issued.
    q := client.Query{
        Command:  "SHOW TAG VALUES FROM iaas_traffic WITH KEY=\"" + HOST_TAG + "\"",
        Database: conf.Database,
    }
    response, err := influxdb.Query(q)
    if err != nil {
        log.Println("[Accouting] failed to get persistent hosts:", err.Error())
        log.Println("[Accouting] exiting plugin")
        return
    }
    if response.Error() != nil {
        log.Println("[Accouting] failed to get persistent hosts:", response.Error().Error())
        log.Println("[Accouting] exiting plugin")
        return
    }

    // maps containing our taffic statistics
    // for the current cycle
    txBytes := make(map[string]int)
    rxBytes := make(map[string]int)

    // make sure all info will be available
    if len(response.Results) > 0 && len(response.Results[0].Series) > 0 {
        // prefill the traffic maps
        for _, row := range response.Results[0].Series[0].Values {
            // row[0] is the column name, row[1] column value
            host := row[1].(string)

            // initialize the hosts stats with zero
            txBytes[host] = 0
            rxBytes[host] = 0
        }
    }

    // begin with the cyclic sampling
    for {
        startTime := time.Now()

        // gather the current ippairs from the router
        accouting, err := util.HttpGet("http://" + conf.Host + "/accounting/ip.cgi")
        if err != nil {
            log.Println("[Accouting] failed to download ip accounting:", err.Error())
        }

        // loop over the lines, which represent a pair of communication partners
        scanner := bufio.NewScanner(strings.NewReader(accouting))
        for scanner.Scan() {
            pair, err := parseIpPair(scanner.Text())
            if err != nil {
                log.Println("[Accouting] failed to parse ip pair:", err.Error())
                continue
            }

            // the source ip is on the net to watch -> upload
            if network.Contains(pair.SrcIp) {
                txBytes[pair.SrcIp.String()] += int(pair.Bytes)

            // the ip on the watch net receives traffic -> download
            } else if network.Contains(pair.DstIp) {
                rxBytes[pair.DstIp.String()] += int(pair.Bytes)
            }
        }

        bp, _ := client.NewBatchPoints(client.BatchPointsConfig{
            Database:  conf.Database,
            Precision: "s",
        })

        // write a point for each host
        for _, host := range hosts(txBytes, rxBytes) {
            // construct the new datapoint
            pt, _ := client.NewPoint(
                conf.Measurement,
                map[string]string{HOST_TAG: host},
                map[string]interface{}{
                    TX_BYTES_FIELD: txBytes[host],
                    RX_BYTES_FIELD: rxBytes[host]},
                time.Now(),
            )
            bp.AddPoint(pt)

            // reset the traffic counters
            txBytes[host] = 0
            rxBytes[host] = 0
        }

        // write the datapoints to influx
        err = influxdb.Write(bp)
        if err != nil {
            log.Println("[Accouting] failed to write:", err.Error())
        }

        // sleep until next execution
        time.Sleep((SAMPLE_TIME * time.Millisecond) - time.Since(startTime))
    }
}

// --------------------------------------------------------------------------------------
//  private functions
// --------------------------------------------------------------------------------------

func hosts(x map[string]int, y map[string]int) ([]string) {
    // fake implementation of a set
    t := make(map[string]interface{})
    for key := range x {
        t[key] = 0
    }

    for key := range y {
        t[key] = 0
    }

    // get all the keys
    keys := make([]string, 0)
    for key := range t {
        keys = append(keys, key)
    }

    return keys
}