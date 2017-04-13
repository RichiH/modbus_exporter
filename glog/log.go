// Copyright 2017 Alejandro Sirgo Rica
//
// This file is part of Modbus_exporter.
//
//     Modbus_exporter is free software: you can redistribute it and/or modify
//     it under the terms of the GNU General Public License as published by
//     the Free Software Foundation, either version 3 of the License, or
//     (at your option) any later version.
//
//     Modbus_exporter is distributed in the hope that it will be useful,
//     but WITHOUT ANY WARRANTY; without even the implied warranty of
//     MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
//     GNU General Public License for more details.
//
//     You should have received a copy of the GNU General Public License
//     along with Modbus_exporter.  If not, see <http://www.gnu.org/licenses/>.

// Package glog manages the global loging from every modbus query
package glog

import "github.com/prometheus/common/log"

// C is the channel exported by the global logging, the external packages have
// to send the errors here.
var C chan<- error

var trackLogs map[string]bool

func init() {
	trackLogs = make(map[string]bool)
	go processLogs()
}

func processLogs() {
	ch := make(chan error, 20)
	C = ch
	for err := range ch {
		// if the error has not been logged it shows it one time.
		// (this will change in the future with filtered errors in time intervals)
		if ok := trackLogs[err.Error()]; !ok {
			log.Errorln(err)
			trackLogs[err.Error()] = true
		}
	}
}
