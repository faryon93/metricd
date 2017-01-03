package config
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
    "strings"
)


// --------------------------------------------------------------------------------------
//  types
// --------------------------------------------------------------------------------------

type Pair string
type PairSlice []Pair


// --------------------------------------------------------------------------------------
//  constructors
// --------------------------------------------------------------------------------------

func GetPairSlice(slice []Pair) PairSlice {
    return slice
}


// --------------------------------------------------------------------------------------
//  public members
// --------------------------------------------------------------------------------------

func (p Pair) Key() string {
    return strings.Split(string(p), ":")[0]
}

func (p Pair) Val() string {
    return strings.Split(string(p), ":")[1]
}

func (s PairSlice) Map() map[string]string {
    m := make(map[string]string)

    for _, entry := range s {
        m[entry.Key()] = entry.Val()
    }

    return m
}