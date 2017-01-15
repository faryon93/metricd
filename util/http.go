package util
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
    "errors"
    "io/ioutil"
    "net/http"
)


// --------------------------------------------------------------------------------------
//  public functions
// --------------------------------------------------------------------------------------

func HttpGet(url string) (string, error) {
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