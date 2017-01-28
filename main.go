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
    "syscall"
    "os/signal"

    "github.com/faryon93/metricd/config"
    "github.com/faryon93/metricd/snmp"
    "github.com/faryon93/metricd/accounting"
    "github.com/faryon93/metricd/mysql"

    "github.com/influxdata/influxdb/client/v2"
)


// --------------------------------------------------------------------------------------
//  constants
// --------------------------------------------------------------------------------------


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

    // setup all snmp measurement watcher
    for name, snmpConf := range conf.Snmp {
        log.Printf("setting up snmp target \"%s\"", name)
        go snmp.Watcher(influx, snmpConf)
    }

    // setup all router os traffic accounting watchers
    for name, accoutingConf := range conf.Accouting {
        log.Printf("setting up accounting target \"%s\"", name)
        go accounting.Watcher(influx, accoutingConf)
    }

    // setup mysql tasks
    for name, mysqlConf := range conf.MySql {
        log.Printf("setting up mysql target \"%s\"", name)
        go mysql.Watcher(influx, mysqlConf)
    }

    // wait for signals to exit application
    wait(os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
}

// --------------------------------------------------------------------------------------
//  helper functions
// --------------------------------------------------------------------------------------

func wait(sig ...os.Signal) {
    signals := make(chan os.Signal)
    signal.Notify(signals, sig...)
    <- signals
}