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
	"fmt"
	"io/ioutil"
	"log"
	"time"

	"github.com/goburrow/modbus"
	"github.com/lupoDharkael/modbus_exporter/config"
	"github.com/lupoDharkael/modbus_exporter/glog"
)

func init() {
	log.SetOutput(ioutil.Discard)
}

// RegisterData receives the data of the modbus systems and start querying the
// modbus slaves in regular intervals in order to expose the data to prometheus
func RegisterData(slaves []config.ParsedSlave, conf config.ListSlaves) error {
	for _, slave := range slaves {
		//var client modbus.Client
		// creates the client (TCP-IP or Serial)
		switch config.CheckPortSlave(conf[slave.Name]) {
		case config.IP:
			handler := modbus.NewTCPClientHandler(conf[slave.Name].Port)
			// diable logger
			handler.Logger = log.New(ioutil.Discard, "", log.LstdFlags)
			//handler.Logger.SetFlags(0)
			if conf[slave.Name].Timeout != 0 {
				handler.Timeout = time.Duration(conf[slave.Name].Timeout) * time.Millisecond
			}
			handler.SlaveId = conf[slave.Name].ID
			if conf[slave.Name].KeepAlive {
				if err := handler.Connect(); err != nil {
					return fmt.Errorf("unable to connect with slave %s at %s",
						slave.Name, conf[slave.Name].Port)
				}
			}
			// starts the data scrapping routine
			go scrapeSlave(slave, &Handler{
				Type:      config.IP,
				KeepAlive: conf[slave.Name].KeepAlive,
				Handler:   handler})
		case config.Serial:
			handler := modbus.NewRTUClientHandler(conf[slave.Name].Port)
			// diable logger
			handler.Logger = log.New(ioutil.Discard, "", log.LstdFlags)
			//handler.Logger.SetFlags(0)
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
				handler.Timeout = time.Duration(conf[slave.Name].Timeout) * time.Millisecond
			}
			handler.SlaveId = conf[slave.Name].ID
			if err := handler.Connect(); err != nil {
				return fmt.Errorf("unable to connect with slave %s at %s",
					slave.Name, conf[slave.Name].Port)
			}
			// starts the data scrapping routine
			go scrapeSlave(slave, &Handler{
				Type:      config.Serial,
				KeepAlive: false,
				Handler:   handler})
		}
	}
	return nil
}

// Handler is an API helper to manage a modbus handler
type Handler struct {
	//Handler         modbus.ClientHandler
	Type      config.PortType
	KeepAlive bool
	Handler   interface {
		modbus.ClientHandler
		Connect() error
		Close() error
	}
}

// Connect starts the connection
func (hc *Handler) Connect() error {
	return hc.Handler.Connect()
}

// Close closes the connection
func (hc *Handler) Close() error {
	return hc.Handler.Close()
}

func scrapeSlave(slave config.ParsedSlave, hc *Handler) { //c modbus.Client) {
	// fetches new data in constant intervals
	c := modbus.NewClient(hc.Handler)
	var (
		err          error
		values       []float64
		connIsClosed bool
	)
	for _ = range time.NewTicker(config.ScrapeInterval).C {
		switch {
		// if the last query went ok
		case err == nil || (!hc.KeepAlive && hc.Type == config.IP):
			if len(slave.DigitalInput) != 0 {
				values, err = getModbusData(slave.DigitalInput,
					c.ReadDiscreteInputs, config.DigitalInput)
				if err != nil {
					glog.C <- fmt.Errorf("[%s:%s] %s",
						slave.Name, config.DigitalInput.String(), err)
				}
				for i, v := range values {
					modbusDigitalIn.WithLabelValues(
						slave.Name,
						slave.DigitalInput[i].Name,
					).Set(v)
				}

			}
			if len(slave.DigitalOutput) != 0 {
				values, err = getModbusData(slave.DigitalOutput,
					c.ReadCoils, config.DigitalOutput)
				if err != nil {
					glog.C <- fmt.Errorf("[%s:%s] %s",
						slave.Name, config.DigitalOutput.String(), err)
				}
				for i, v := range values {
					modbusDigitalOut.WithLabelValues(
						slave.Name,
						slave.DigitalOutput[i].Name,
					).Set(v)
				}
			}
			if len(slave.AnalogInput) != 0 {
				values, err = getModbusData(slave.AnalogInput,
					c.ReadInputRegisters, config.AnalogInput)
				if err != nil {
					glog.C <- fmt.Errorf("[%s:%s] %s",
						slave.Name, config.AnalogInput.String(), err)
				}
				for i, v := range values {
					modbusAnalogIn.WithLabelValues(
						slave.Name,
						slave.AnalogInput[i].Name,
					).Set(v)
				}
			}

			if len(slave.AnalogOutput) != 0 {
				values, err = getModbusData(slave.AnalogOutput,
					c.ReadHoldingRegisters, config.AnalogOutput)
				if err != nil {
					glog.C <- fmt.Errorf("[%s:%s] %s",
						slave.Name, config.AnalogOutput.String(), err)
				}
				for i, v := range values {
					modbusAnalogOut.WithLabelValues(
						slave.Name,
						slave.AnalogOutput[i].Name,
					).Set(v)
				}
			}
			// in case of non failure evades the fallthrough which starts a reconnection
			if !(err != nil &&
				(hc.Type == config.Serial || (hc.KeepAlive && hc.Type == config.IP))) {
				continue
			}
			fallthrough
		// when we need to reconnect
		case err != nil &&
			(hc.Type == config.Serial || (hc.KeepAlive && hc.Type == config.IP)):
			if !connIsClosed {
				hc.Close()
				connIsClosed = true
			}
			err = hc.Connect()
			if err == nil {
				connIsClosed = false
			}
		}
	}
}

// modbus read function type
type modbusFunc func(address, quantity uint16) ([]byte, error)

// getModbusData returns the list of values from a slave
func getModbusData(registers []config.Register, f modbusFunc, t config.RegType) ([]float64, error) {
	// results contains the values to be returned
	results := make([]float64, 0, 125)
	// saves first and last register value to be obtained
	first := registers[0].Value
	last := registers[len(registers)-1].Value
	// error needed to evade the
	var err error
	// tracking of the actual index in the registers received as parameter
	regIndex := 0
	// range of elements to be queried
	rangeN := (last - first) + 1
	// number of maximum values per query
	var div uint16
	switch t {
	case config.DigitalInput, config.DigitalOutput:
		div = 2000 // max registers for a digital query
	case config.AnalogInput, config.AnalogOutput:
		div = 125 // max registers for an analog query
	}
	for it := int(rangeN / div); it >= 0; it-- {
		// Temporal slice for every modbus query.
		modBytes := []byte{}
		// The number of the first register loaded into `modBytes`.
		modBytesFirstRegister := first

		if it > 0 {
			// query the maximum number of registers
			modBytes, err = f(first, div)
			first += div
		} else {
			// query the last elements denoted by the incremented 'first' and the last
			modBytes, err = f(first, (last-first)+1)
		}

		if err != nil {
			results = make([]float64, len(registers))
			break
		}

		// i < int(div-1) make sure not to try to access anything outside
		// the maximum length of digital or analog return.
		for i := 0; i < int(rangeN) && i < int(div-1); i++ {
			// Check if we are already done.
			if regIndex >= len(registers) {
				break
			}

			switch t {
			case config.DigitalInput, config.DigitalOutput:
				if modBytesFirstRegister+uint16(i) == registers[regIndex].Value {
					data := float64((modBytes[i/8] >> uint16(i) % 8) & 1)
					results = append(results, data)
					regIndex++
				}
			case config.AnalogInput, config.AnalogOutput:
				if modBytesFirstRegister+uint16(i) == registers[regIndex].Value {
					data := float64(modBytes[i*2])*256 + float64(modBytes[(i*2)+1])
					results = append(results, data)
					regIndex++
				}
			}
		}
	}
	return results, err
}
