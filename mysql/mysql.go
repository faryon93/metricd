package mysql
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

// ----------------------------------------------------------------------------------
//  imports
// ----------------------------------------------------------------------------------

import (
	"log"
	"database/sql"
	"time"

	"github.com/faryon93/metricd/config"

	"github.com/influxdata/influxdb/client/v2"
	_ "github.com/go-sql-driver/mysql"
)


// ----------------------------------------------------------------------------------
//  public functions
// ----------------------------------------------------------------------------------

func Watcher(influxdb client.Client, conf config.MySqlConf) {
	// open mysql connection
	db, err := sql.Open("mysql", conf.SqlUri)
	if err != nil {
		log.Println("failed to connect to mysql server:", err.Error())
		return
	}
	defer db.Close()

	// check connection
	err = db.Ping()
	if err != nil {
		log.Println("failed to connect to mysql server:", err.Error())
		return
	}

	for {
		startTime := time.Now()
		bp, _ := client.NewBatchPoints(client.BatchPointsConfig{
            Database:  conf.Database,
            Precision: "s",
        })

		values := make(map[string]interface{})
		for _, query := range conf.Queries {
			// execute the query
			rows, err := db.Query(query.Val())
			if err != nil {
				log.Printf("failed to execute query: %s", err.Error())
				continue
			}

			// parse result			
			if rows.Next() {
				var value int
				err = rows.Scan(&value)
				if err != nil {
					log.Printf("failed to read query: %s", err.Error())
					continue
				}

				values[query.Key()] = value
			}
		}

		// construct the new datapoint
		pt, _ := client.NewPoint(
			conf.Measurement,
			config.GetPairSlice(conf.Tags).Map(),
			values,
			startTime,
		)
		bp.AddPoint(pt)

		// write the datapoints to influx
        err = influxdb.Write(bp)
        if err != nil {
            log.Println("[Accouting] failed to write:", err.Error())
        }

		// sleep until next execution
        time.Sleep((time.Duration(conf.SampleTime) * time.Millisecond) - time.Since(startTime))
	}
}