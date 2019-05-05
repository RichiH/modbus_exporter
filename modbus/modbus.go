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

	"github.com/prometheus/client_golang/prometheus"

	"github.com/goburrow/modbus"
	"github.com/lupoDharkael/modbus_exporter/config"
)

func init() {
	log.SetOutput(ioutil.Discard)
}

// Exporter represents a Prometheus exporter converting modbus information
// retrieved from remote targets via TCP as Prometheus style metrics.
type Exporter struct {
	config config.Config
}

// NewExporter returns a new modbus exporter.
func NewExporter(config config.Config) *Exporter {
	return &Exporter{config}
}

// Scrape scrapes the given target via TCP based on the configuration of the
// specified module returning a Prometheus gatherer with the resulting metrics.
func (e *Exporter) Scrape(targetAddress, moduleName string) (prometheus.Gatherer, error) {
	reg := prometheus.NewRegistry()

	var module config.Module

	for _, m := range e.config.Modules {
		if m.Name == moduleName {
			module = m
		}
	}

	// TODO: Not a nice way of checking whether the module was found.
	if module.Name == "" {
		return nil, fmt.Errorf("failed to find %v in config", moduleName)
	}

	request := newScrapeRequest(reg)

	protocol, err := config.CheckPortTarget(targetAddress)
	if err != nil {
		return nil, err
	}

	if protocol != module.Protocol {
		return nil, fmt.Errorf(
			"target address protocol and module protocol don't match '%v', '%v'",
			protocol,
			module.Protocol,
		)
	}

	//var client modbus.Client
	// creates the client (TCP-IP or Serial)
	switch module.Protocol {
	case config.ModbusProtocolTCPIP:
		// TODO: We should probably be reusing these, right?
		handler := modbus.NewTCPClientHandler(targetAddress)
		// diable logger
		handler.Logger = log.New(ioutil.Discard, "", log.LstdFlags)
		//handler.Logger.SetFlags(0)
		if module.Timeout != 0 {
			handler.Timeout = time.Duration(module.Timeout) * time.Millisecond
		}
		handler.SlaveId = module.ID
		if module.KeepAlive {
			if err := handler.Connect(); err != nil {
				return nil, fmt.Errorf("unable to connect with target %s via module %s",
					targetAddress, module.Name)
			}
		}
		// starts the data scrapping routine
		err := request.scrapeSlave(module, &Handler{
			Type:      config.ModbusProtocolTCPIP,
			KeepAlive: module.KeepAlive,
			Handler:   handler})
		if err != nil {
			return nil, err
		}
	case config.ModbusProtocolSerial:
		handler := modbus.NewRTUClientHandler(targetAddress)
		// diable logger
		handler.Logger = log.New(ioutil.Discard, "", log.LstdFlags)
		//handler.Logger.SetFlags(0)
		if module.Baudrate != 0 {
			handler.BaudRate = module.Baudrate
		}
		if module.Databits != 0 {
			handler.DataBits = module.Databits
		}
		if module.Parity != "" {
			handler.Parity = module.Parity
		}
		if module.Stopbits != 0 {
			handler.StopBits = module.Stopbits
		}
		if module.Timeout != 0 {
			handler.Timeout = time.Duration(module.Timeout) * time.Millisecond
		}
		handler.SlaveId = module.ID
		if err := handler.Connect(); err != nil {
			return nil, fmt.Errorf("unable to connect with target %s via module %s",
				targetAddress, module.Name)
		}
		// starts the data scrapping routine
		err := request.scrapeSlave(module, &Handler{
			Type:      config.ModbusProtocolSerial,
			KeepAlive: false,
			Handler:   handler})
		if err != nil {
			return nil, err
		}
	}

	return reg, nil
}

func newScrapeRequest(reg prometheus.Registerer) *scrapeRequest {
	request := &scrapeRequest{}

	request.modbusDigitalIn = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "modbus_digital_input_total",
			Help: "Modbus digital input registers.",
		},
		[]string{"slave", "name"},
	)

	request.modbusAnalogIn = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "modbus_analog_input_total",
			Help: "Modbus analog input registers.",
		},
		[]string{"slave", "name"},
	)
	request.modbusDigitalOut = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "modbus_digital_output_total",
			Help: "Modbus digital output registers.",
		},
		[]string{"slave", "name"},
	)

	request.modbusAnalogOut = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "modbus_analog_output_total",
			Help: "Modbus analog output registers.",
		},
		[]string{"slave", "name"},
	)

	reg.MustRegister(
		request.modbusDigitalIn,
		request.modbusDigitalOut,
		request.modbusAnalogIn,
		request.modbusAnalogOut,
	)

	return request
}

type scrapeRequest struct {
	modbusDigitalIn  *prometheus.GaugeVec
	modbusAnalogIn   *prometheus.GaugeVec
	modbusDigitalOut *prometheus.GaugeVec
	modbusAnalogOut  *prometheus.GaugeVec
}

// Handler is an API helper to manage a modbus handler
//
// TODO: Can we get rid of this?
type Handler struct {
	//Handler         modbus.ClientHandler
	Type      config.ModbusProtocol
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

func (r *scrapeRequest) scrapeSlave(slave config.Module, hc *Handler) error {
	// TODO: Should we reuse this?
	c := modbus.NewClient(hc.Handler)
	var (
		values []float64
		err    error
	)

	if hc.Type == config.ModbusProtocolTCPIP {
		if len(slave.DigitalInput) != 0 {
			values, err = getModbusData(slave.DigitalInput,
				c.ReadDiscreteInputs, config.DigitalInput)
			if err != nil {
				return fmt.Errorf("[%s:%s] %s",
					slave.Name, config.DigitalInput.String(), err)
			}
			for i, v := range values {
				r.modbusDigitalIn.WithLabelValues(
					slave.Name,
					slave.DigitalInput[i].Name,
				).Set(v)
			}

		}
		if len(slave.DigitalOutput) != 0 {
			values, err = getModbusData(slave.DigitalOutput,
				c.ReadCoils, config.DigitalOutput)
			if err != nil {
				return fmt.Errorf("[%s:%s] %s",
					slave.Name, config.DigitalOutput.String(), err)
			}
			for i, v := range values {
				r.modbusDigitalOut.WithLabelValues(
					slave.Name,
					slave.DigitalOutput[i].Name,
				).Set(v)
			}
		}
		if len(slave.AnalogInput) != 0 {
			values, err = getModbusData(slave.AnalogInput,
				c.ReadInputRegisters, config.AnalogInput)
			if err != nil {
				return fmt.Errorf("[%s:%s] %s",
					slave.Name, config.AnalogInput.String(), err)
			}
			for i, v := range values {
				r.modbusAnalogIn.WithLabelValues(
					slave.Name,
					slave.AnalogInput[i].Name,
				).Set(v)
			}
		}

		if len(slave.AnalogOutput) != 0 {
			values, err = getModbusData(slave.AnalogOutput,
				c.ReadHoldingRegisters, config.AnalogOutput)
			if err != nil {
				return fmt.Errorf("[%s:%s] %s",
					slave.Name, config.AnalogOutput.String(), err)
			}
			for i, v := range values {
				r.modbusAnalogOut.WithLabelValues(
					slave.Name,
					slave.AnalogOutput[i].Name,
				).Set(v)
			}
		}
	}

	return nil
}

// modbus read function type
type modbusFunc func(address, quantity uint16) ([]byte, error)

// getModbusData returns the list of values from a slave
// TODO: rename registers to definitions.
func getModbusData(registers []config.MetricDef, f modbusFunc, t config.RegType) ([]float64, error) {
	if len(registers) == 0 {
		return []float64{}, nil
	}

	// results contains the values to be returned
	results := make([]float64, 0, 125)

	// saves first and last register value to be obtained
	first := registers[0].Address
	var last config.RegisterAddr
	for _, def := range registers {
		if def.Address > last {
			last = def.Address
		}
	}

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
	for it := int(uint16(rangeN) / div); it >= 0; it-- {
		// Temporal slice for every modbus query.
		modBytes := []byte{}
		// The number of the first register loaded into `modBytes`.
		modBytesFirstRegister := first

		if it > 0 {
			// query the maximum number of registers
			modBytes, err = f(uint16(first), div)
			first += config.RegisterAddr(div)
		} else {
			// query the last elements denoted by the incremented 'first' and the last
			modBytes, err = f(uint16(first), uint16(last-first)+1)
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
				if modBytesFirstRegister+config.RegisterAddr(i) == registers[regIndex].Address {
					// TODO: Use metric definition parse.
					data := float64((modBytes[i/8] >> uint16(i) % 8) & 1)
					results = append(results, data)
					regIndex++
				}
			case config.AnalogInput, config.AnalogOutput:
				if modBytesFirstRegister+config.RegisterAddr(i) == registers[regIndex].Address {
					data, err := registers[regIndex].Parse([2]byte{modBytes[i*2], modBytes[(i*2)+1]})
					if err != nil {
						return []float64{}, err
					}
					results = append(results, data)
					regIndex++
				}
			}
		}
	}
	return results, err
}
