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
	"encoding/binary"
	"fmt"
	"io/ioutil"
	"log"
	"math"
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
		err := request.scrape(module, &Handler{
			Type:      config.ModbusProtocolTCPIP,
			KeepAlive: module.KeepAlive,
			Handler:   handler})
		if err != nil {
			return nil, err
		}
	case config.ModbusProtocolSerial:
		handler := modbus.NewRTUClientHandler(targetAddress)
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
		err := request.scrape(module, &Handler{
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
		[]string{"module", "name"},
	)

	request.modbusAnalogIn = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "modbus_analog_input_total",
			Help: "Modbus analog input registers.",
		},
		[]string{"module", "name"},
	)
	request.modbusDigitalOut = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "modbus_digital_output_total",
			Help: "Modbus digital output registers.",
		},
		[]string{"module", "name"},
	)

	request.modbusAnalogOut = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "modbus_analog_output_total",
			Help: "Modbus analog output registers.",
		},
		[]string{"module", "name"},
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

func (r *scrapeRequest) scrape(module config.Module, hc *Handler) error {
	// TODO: Should we reuse this?
	c := modbus.NewClient(hc.Handler)
	var (
		values []float64
		err    error
	)

	if hc.Type == config.ModbusProtocolTCPIP {
		if len(module.DigitalInput) != 0 {
			values, err = getModbusData(module.DigitalInput,
				c.ReadDiscreteInputs, config.DigitalInput)
			if err != nil {
				return fmt.Errorf("[%s:%s] %s",
					module.Name, config.DigitalInput.String(), err)
			}
			for i, v := range values {
				r.modbusDigitalIn.WithLabelValues(
					module.Name,
					module.DigitalInput[i].Name,
				).Set(v)
			}

		}
		if len(module.DigitalOutput) != 0 {
			values, err = getModbusData(module.DigitalOutput,
				c.ReadCoils, config.DigitalOutput)
			if err != nil {
				return fmt.Errorf("[%s:%s] %s",
					module.Name, config.DigitalOutput.String(), err)
			}
			for i, v := range values {
				r.modbusDigitalOut.WithLabelValues(
					module.Name,
					module.DigitalOutput[i].Name,
				).Set(v)
			}
		}
		if len(module.AnalogInput) != 0 {
			values, err = getModbusData(module.AnalogInput,
				c.ReadInputRegisters, config.AnalogInput)
			if err != nil {
				return fmt.Errorf("[%s:%s] %s",
					module.Name, config.AnalogInput.String(), err)
			}
			for i, v := range values {
				r.modbusAnalogIn.WithLabelValues(
					module.Name,
					module.AnalogInput[i].Name,
				).Set(v)
			}
		}

		if len(module.AnalogOutput) != 0 {
			values, err = getModbusData(module.AnalogOutput,
				c.ReadHoldingRegisters, config.AnalogOutput)
			if err != nil {
				return fmt.Errorf("[%s:%s] %s",
					module.Name, config.AnalogOutput.String(), err)
			}
			for i, v := range values {
				r.modbusAnalogOut.WithLabelValues(
					module.Name,
					module.AnalogOutput[i].Name,
				).Set(v)
			}
		}
	}

	return nil
}

// modbus read function type
type modbusFunc func(address, quantity uint16) ([]byte, error)

// getModbusData returns the list of values from a target
func getModbusData(definitions []config.MetricDef, f modbusFunc, t config.RegType) ([]float64, error) {
	if len(definitions) == 0 {
		return []float64{}, nil
	}

	results := []float64{}

	// number of maximum values per query
	var div uint16
	switch t {
	case config.DigitalInput, config.DigitalOutput:
		div = 2000 // max registers for a digital query
	case config.AnalogInput, config.AnalogOutput:
		div = 125 // max registers for an analog query
	}

	for _, definition := range definitions {
		// TODO: We could cache the results to not repeat overlapping ones.
		modBytes, err := f(uint16(definition.Address), div)
		if err != nil {
			return []float64{}, err
		}

		result, err := parseModbusData(definition, modBytes)
		if err != nil {
			return []float64{}, err
		}

		results = append(results, result)
	}

	return results, nil
}

// InsufficientRegistersError is returned in Parse() whenever not enough
// registers are provided for the given data type.
type InsufficientRegistersError struct {
	e string
}

// Error implements the Golang error interface.
func (e *InsufficientRegistersError) Error() string {
	return fmt.Sprintf("insufficient amount of registers provided: %v", e.e)
}

// Parse parses the given byte slice based on the specified Modbus data type and
// returns the parsed value as a float64 (Prometheus exposition format).
//
// TODO: Handle Endianness.
func parseModbusData(d config.MetricDef, rawData []byte) (float64, error) {
	switch d.DataType {
	case config.ModbusFloat16:
		if len(rawData) < 2 {
			return float64(0), &InsufficientRegistersError{fmt.Sprintf("expected at least 1, got %v", len(rawData))}
		}
		panic("implement")
	case config.ModbusFloat32:
		if len(rawData) < 4 {
			return float64(0), &InsufficientRegistersError{fmt.Sprintf("expected at least 2, got %v", len(rawData))}
		}
		return float64(math.Float32frombits(binary.BigEndian.Uint32(rawData[:4]))), nil
	case config.ModbusInt16:
		{
			if len(rawData) < 2 {
				return float64(0), &InsufficientRegistersError{fmt.Sprintf("expected at least 1, got %v", len(rawData))}
			}
			i := binary.BigEndian.Uint16(rawData)
			return float64(int16(i)), nil
		}
	case config.ModbusUInt16:
		{
			if len(rawData) < 2 {
				return float64(0), &InsufficientRegistersError{fmt.Sprintf("expected at least 1, got %v", len(rawData))}
			}
			i := binary.BigEndian.Uint16(rawData)
			return float64(i), nil
		}
	case config.ModbusBool:
		{
			// TODO: Maybe we don't need two registers for bool.
			if len(rawData) < 2 {
				return float64(0), &InsufficientRegistersError{fmt.Sprintf("expected at least 1, got %v", len(rawData))}
			}

			if d.BitOffset == nil {
				return float64(0), fmt.Errorf("expected bit position on boolean data type")
			}

			data := binary.BigEndian.Uint16(rawData)

			if data&(uint16(1)<<uint16(*d.BitOffset)) > 0 {
				return float64(1), nil
			}
			return float64(0), nil
		}
	default:
		return 0, fmt.Errorf("unknown modbus data type")
	}
}
