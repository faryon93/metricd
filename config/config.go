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
	"github.com/BurntSushi/toml"
)


// --------------------------------------------------------------------------------------
//  types
// --------------------------------------------------------------------------------------

type Config struct {
	Influx struct {
		Address string `toml:"address"`
		User string `toml:"user"`
		Password string `toml:"password"`
	} `toml:"influx"`
	Snmp map[string]SnmpConf
	Accouting map[string]AccoutingConf
	MySql map[string]MySqlConf
}

type SnmpConf struct {
	Host string `toml:"host"`
	Community string `toml:"community"`
	Database string `toml:"database"`
	Measurement string `toml:"measurement"`
	SampleTime int `toml:"sample_time"`
	DataPoints []Pair `toml:"datapoints"`
	Tags []Pair `toml:"tags"`
}

type AccoutingConf struct {
	Database string `toml:"database"`
	Measurement string `toml:"measurement"`

	Host string `toml:"host"`
	Network string `toml:"network"`
	Exclude []string `toml:"exclude"`
}

type MySqlConf struct {
	Database string `toml:"database"`
	Measurement string `toml:"measurement"`
	SampleTime int `toml:"sample_time"`

	SqlUri string `toml:"sql_uri"`	
	Queries []Pair `toml:"queries"`
	Tags []Pair `toml:"tags"`
}


// --------------------------------------------------------------------------------------
//  public functions
// --------------------------------------------------------------------------------------

func Load(path string) (*Config, error) {
	// decode the config file to struct
	var config Config
	if _, err := toml.DecodeFile(path, &config); err != nil {
		return nil, err
	}

	return &config, nil
}

