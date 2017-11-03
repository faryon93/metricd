package recorder
// metricd
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

    "github.com/influxdata/influxdb/client/v2"
)

// --------------------------------------------------------------------------------------
//  types
// --------------------------------------------------------------------------------------

type Recorder interface {
    GetDatabase() string
    GetMeasurement() string
    IsValid() (error)

    Setup(client client.Client) error
    Teardown() error
    Run()
}

type Tags map[string]string
type Values map[string]interface{}

// --------------------------------------------------------------------------------------
//  public functions
// --------------------------------------------------------------------------------------

func Point(r Recorder, precision string, tags Tags, values Values, timestamp time.Time) (client.BatchPoints, error) {
    bp, _ := client.NewBatchPoints(client.BatchPointsConfig{
        Precision: precision,
        Database:  r.GetDatabase(),
    })

    // construct the new databpoint for influxdb
    pt, _ := client.NewPoint(
        r.GetMeasurement(),
        tags, values, timestamp,
    )
    bp.AddPoint(pt)

    return bp, nil
}

func PointUs(r Recorder, tags Tags, values Values, timestamp time.Time) (client.BatchPoints, error) {
    return Point(r, "us", tags, values, timestamp)
}