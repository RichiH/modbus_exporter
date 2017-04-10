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
			if err := handler.Connect(); err != nil {
				return err
			}
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
			if err := handler.Connect(); err != nil {
				return err
			}
			client = modbus.NewClient(handler)
		}
		// starts the data scrapping routine
		go scrapeSlave(slave, client)
	}
	return nil
}

func scrapeSlave(slave config.ParsedSlave, c modbus.Client) {
	// fetches new data in constant intervals
	for _ = range time.NewTicker(time.Second * 5).C {
		if len(slave.DigitalInput) != 0 {
			values := getModbusData(slave.DigitalInput, c.ReadDiscreteInputs)
			for _, v := range values {
				modbusDigital.WithLabelValues(
					slave.Name, config.DigitalInput.String()).Set(v)
			}
		}
		if len(slave.DigitalOutput) != 0 {
			values := getModbusData(slave.DigitalOutput, c.ReadCoils)
			for _, v := range values {
				modbusDigital.WithLabelValues(
					slave.Name, config.DigitalOutput.String()).Set(v)
			}
		}
		if len(slave.AnalogInput) != 0 {
			values := getModbusData(slave.AnalogInput, c.ReadInputRegisters)
			for _, v := range values {
				modbusAnalog.WithLabelValues(
					slave.Name, config.AnalogInput.String()).Set(v)
			}
		}
		if len(slave.AnalogOutput) != 0 {
			values := getModbusData(slave.AnalogOutput, c.ReadHoldingRegisters)
			for _, v := range values {
				modbusAnalog.WithLabelValues(
					slave.Name, config.AnalogOutput.String()).Set(v)
			}
		}
	}
}

// modbus read function type
type modbusFunc func(address, quantity uint16) ([]byte, error)

// getModbusData returns the list of values from a slave
func getModbusData(registers []config.Register, f modbusFunc) []float64 {
	// results contains the values to be returned
	results := make([]float64, 0, 125)
	// temporal slice for every modbus query
	var modBytes []byte
	// saves first and last register value to be obtained
	first := registers[0].Value
	last := registers[len(registers)-1].Value
	// error needed to evade the
	var err error
	// tracking of the actual index in the registers received as parameter
	regIndex := 0

	for i := ((last - first) / 125); i >= 0; i-- {
		if i > 0 {
			modBytes, err = f(first, 125)
			if err != nil {
				temp := make([]float64, 125)
				results = append(results, temp...)
				continue
			}
			first += 125
		} else if (last - first) != 0 {
			modBytes, err = f(first, last-first)
			if err != nil {
				temp := make([]float64, (last - first))
				results = append(results, temp...)
				continue
			}
		}
		for indexRes := 0; indexRes <= len(modBytes); indexRes += 2 {
			if registers[0].Value+uint16(indexRes/2) == registers[regIndex].Value {
				data := float64(modBytes[indexRes])*256 + float64(modBytes[indexRes+1])
				results = append(results, data)
				regIndex++
				continue
			}
		}
	}
	return results
}
