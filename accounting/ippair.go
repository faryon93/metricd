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
    "net"
    "strings"
    "errors"
    "strconv"
    "fmt"
)

// --------------------------------------------------------------------------------------
//  constants
// --------------------------------------------------------------------------------------

const (
    FIELD_SPERATOR = " "
)

const (
    FIELD_SRC_IP = 0
    FIELD_DST_IP = 1
    FIELD_BYTES = 2
    FIELD_PACKETS = 3
)


// --------------------------------------------------------------------------------------
//  types
// --------------------------------------------------------------------------------------

type IpPair struct {
    SrcIp net.IP    // Source IP Address
    DstIp net.IP    // Destination IP Addres
    Packets uint64  // Transmitted Packets (Src -> Dst)
    Bytes uint64    // Transmitted Packets (Src -> Dst)
}


// --------------------------------------------------------------------------------------
//  constructors
// --------------------------------------------------------------------------------------

func parseIpPair(raw string) (*IpPair, error) {
    // split up the line
    parsed := strings.Split(raw, FIELD_SPERATOR)
    if len(parsed) < 4 {
        return nil, errors.New("invalid number of fields found")
    }

    // parse int values
    packets, err := strconv.Atoi(parsed[FIELD_PACKETS])
    if err != nil {
        return nil, err
    }

    bytes, err := strconv.Atoi(parsed[FIELD_BYTES])
    if err != nil {
        return nil, err
    }

    // construct object
    return &IpPair{
        SrcIp: net.ParseIP(parsed[FIELD_SRC_IP]),
        DstIp: net.ParseIP(parsed[FIELD_DST_IP]),
        Packets: uint64(packets),
        Bytes: uint64(bytes),
    }, nil
}


// --------------------------------------------------------------------------------------
//  public members
// --------------------------------------------------------------------------------------

func (p *IpPair) IsInsideNetwork(network string) (bool) {
    _, ipnet, err := net.ParseCIDR(network)
    if err != nil {
        return false
    }

    return ipnet.Contains(p.SrcIp) || ipnet.Contains(p.DstIp)
}

func (p *IpPair) String() string {
    return fmt.Sprintf("%s -> %s: %d, %d", p.SrcIp.String(), p.DstIp.String(),
                                                  p.Packets, p.Bytes)
}