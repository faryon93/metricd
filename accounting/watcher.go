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
    "net/http"
    "io/ioutil"
    "bufio"
    "strings"
    "errors"

    "github.com/faryon93/metricd/config"

    "github.com/influxdata/influxdb/client/v2"
)


// --------------------------------------------------------------------------------------
//  constants
// --------------------------------------------------------------------------------------

const (
    SAMPLE_TIME = 1000
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

    for {
        startTime := time.Now()

        // gather the current ippairs from the router
        accouting, err := get("http://" + conf.Host + "/accounting/ip.cgi")
        if err != nil {
            log.Println("[Accouting] failed to download ip accounting:", err.Error())
        }

        txBytes := make(map[string]int)
        rxBytes := make(map[string]int)

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
                map[string]string {
                    "host": host,
                },
                map[string]interface{}{
                    "tx_bytes": txBytes[host],
                    "rx_bytes": rxBytes[host],
                },
                time.Now(),
            )
            bp.AddPoint(pt)
        }

        // write the datapoints to influx
        err = influxdb.Write(bp)
        if err != nil {
            log.Println("failed to write datapoint:", err.Error())
        }

        // sleep until next execution
        time.Sleep((SAMPLE_TIME * time.Millisecond) - time.Since(startTime))
    }
}

// --------------------------------------------------------------------------------------
//  private functions
// --------------------------------------------------------------------------------------

func get(url string) (string, error) {
    // make the http get request
    response, err := http.Get(url)
    if err != nil {
        return "", err
    }
    defer response.Body.Close()

    if response.StatusCode != 200 {
        return "", errors.New("invalid status code: " + response.Status)
    }

    // read the whole body
    body, err := ioutil.ReadAll(response.Body)
    if err != nil {
        return "", err
    }

    return string(body), err
}

func hosts(x map[string]int, y map[string]int) []string {
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