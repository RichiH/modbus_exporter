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

// Package modbus contains all the modbus related components
package modbus

import (
	"time"

	"github.com/goburrow/modbus"
	"github.com/lupoDharkael/modbus_exporter/config"
)

// RegisterData receives the data of the modbus systems and start querying the
// modbus slaves in regular intervals in order to expose the data to prometheus
func RegisterData(slaves []config.ParsedSlave, conf config.ListSlaves) error {
	for _, slave := range slaves {
		var client modbus.Client
		// creates the client (TCP-IP or Serial)
		switch config.CheckPortSlave(conf[slave.Name]) {
		case config.IP:
			handler := modbus.NewTCPClientHandler(conf[slave.Name].Port)
			if conf[slave.Name].Timeout != 0 {
				handler.Timeout = time.Duration(conf[slave.Name].Timeout) * time.Millisecond
			}
			handler.SlaveId = conf[slave.Name].ID
			// Connect manually so that multiple requests are handled in one connection session
			if err := handler.Connect(); err != nil {
				return err
			}
			defer handler.Close()
			client = modbus.NewClient(handler)
		case config.Serial:
			handler := modbus.NewRTUClientHandler(conf[slave.Name].Port)
			if conf[slave.Name].Baudrate != 0 {
				handler.BaudRate = conf[slave.Name].Baudrate
			}
			if conf[slave.Name].Databits != 0 {
				handler.DataBits = conf[slave.Name].Databits
			}
			if conf[slave.Name].Parity != "" {
				handler.Parity = conf[slave.Name].Parity
			}
			if conf[slave.Name].Stopbits != 0 {
				handler.StopBits = conf[slave.Name].Stopbits
			}
			if conf[slave.Name].Timeout != 0 {
				handler.Timeout = time.Duration(conf[slave.Name].Timeout) * time.Second
			}
			handler.SlaveId = conf[slave.Name].ID
			// Connect manually so that multiple requests are handled in one connection session
			if err := handler.Connect(); err != nil {
				return err
			}
			defer handler.Close()
			client = modbus.NewClient(handler)
		}
		scrapeSlave(slaves, client)
	}
	return nil
}

func scrapeSlave(slaves []config.ParsedSlave, c modbus.Client) {

}
