package syslog

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
	"errors"
	"log"
	"regexp"
	"strconv"
	"time"

	"github.com/influxdata/influxdb/client/v2"
	"gopkg.in/mcuadros/go-syslog.v2"

	"github.com/faryon93/metricd/recorder"
	"github.com/faryon93/metricd/util"
	"strings"
)

// --------------------------------------------------------------------------------------
//  constants
// --------------------------------------------------------------------------------------

const (
	TAG_PREFIX   = "tag_"
	VALUE_PREFIX = "val_"
)

// --------------------------------------------------------------------------------------
//  types
// --------------------------------------------------------------------------------------

type Recorder struct {
	Listen      string      `toml:"listen"`
	Regex       string      `toml:"regex"`
	Database    string      `toml:"database"`
	Measurement string      `toml:"measurement"`
	Tags        []util.Pair `toml:"tags"`

	// private runtime variables
	matcher  *regexp.Regexp
	messages syslog.LogPartsChannel
	syslog   *syslog.Server
	influx   client.Client
}

// ----------------------------------------------------------------------------------
//  public members
// ----------------------------------------------------------------------------------

// Returns the database to use.
func (r *Recorder) GetDatabase() string {
	return r.Database
}

// Returns the measurement to use.
func (r *Recorder) GetMeasurement() string {
	return r.Measurement
}

// Returns nil when the configuration is valid.
// Otherswise an error with a description is returned.
func (r *Recorder) IsValid() error {
	if len(r.Listen) < 1 {
		return errors.New("invalid listen property \"" + r.Listen + "\"")
	}

	// TODO: make sure at least on val_ group exists in regex

	return nil
}

// Creates the UDP server which listens for incoming syslog messages.
func (r *Recorder) Setup(influx client.Client) error {
	var err error

	// compile the regex from config
	r.matcher, err = regexp.Compile(r.Regex)
	if err != nil {
		return err
	}

	// configure the syslog server
	r.messages = make(syslog.LogPartsChannel)
	r.syslog = syslog.NewServer()
	r.syslog.SetFormat(syslog.RFC3164)
	r.syslog.SetHandler(syslog.NewChannelHandler(r.messages))

	// boot the udp server to start reception of log messages
	err = r.syslog.ListenUDP(r.Listen)
	if err != nil {
		return err
	}

	err = r.syslog.Boot()
	if err != nil {
		return err
	}

	r.influx = influx

	return nil
}

// Destroys the syslog server.
func (r *Recorder) Teardown() error {
	return r.syslog.Kill()
}

// Write the give points to influx database.
func (r *Recorder) Write(points client.BatchPoints) error {
	if len(points.Points()) < 1 {
		log.Println("skipping empty batch")
		return nil
	}

	return r.influx.Write(points)
}

// Processes all incomming syslog messages and transforms them
// into influxdb points.
func (r *Recorder) Run() {
	for message := range r.messages {
		// parse the syslog message and make sure everything exists
		timestamp := time.Now()
		content, ok := message["content"].(string)
		if !ok {
			log.Println("[syslog] missing message field \"content\"")
			continue
		}

		// check if the received log messages matches the
		// configured regex
		matches := r.matcher.FindStringSubmatch(content)
		if len(matches) < len(r.matcher.SubexpNames()) {
			continue
		}

		// process the message
		tags, values, err := r.process(matches)
		if err != nil {
			log.Println("[syslog] failed to process message:", err.Error())
			continue
		}

		points, err := recorder.PointUs(r, tags, values, timestamp)
		if err != nil {
			log.Println("[syslog] failed to create point:", err.Error())
			continue
		}

		// write points to the database
		err = r.Write(points)
		if err != nil {
			log.Println("[syslog] write point failed:", err.Error())
			continue
		}
	}
}

// ----------------------------------------------------------------------------------
//  private members
// ----------------------------------------------------------------------------------

// Processes a log messages.
func (r *Recorder) process(matches []string) (recorder.Tags, recorder.Values, error) {
	// maps which are used to construct the new datapoint
	tags := util.GetPairSlice(r.Tags).Map()
	values := make(map[string]interface{})

	// process all regex caputure groups and add to the coresponding
	// map in oder to insert the data into the datapoint
	for i, name := range r.matcher.SubexpNames() {
		if i > 0 && len(name) > 0 {
			val := matches[i]

			// we are processing a tag
			if strings.HasPrefix(name, TAG_PREFIX) {
				tags[strings.TrimPrefix(name, TAG_PREFIX)] = val

			// we are processing a value
			} else if strings.HasPrefix(name, VALUE_PREFIX) {
				// convert to floating point value
				value, err := strconv.ParseFloat(val, 32)
				if err != nil {
					return nil, nil, err
				}

				values[strings.TrimPrefix(name, VALUE_PREFIX)] = float32(value)
			} else {
				return nil, nil, errors.New("unknown capture group nameing prefix")
			}
		}
	}

	return tags, values, nil
}
