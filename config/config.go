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

package config

import (
	"fmt"

	multierror "github.com/hashicorp/go-multierror"
)

// Config represents the configuration of the modbus exporter.
type Config struct {
	Modules []Module `yaml:"modules"`
}

// validate semantically validates the given config.
func (c *Config) validate() error {
	for _, t := range c.Modules {
		if err := t.validate(); err != nil {
			return err
		}
	}

	return nil
}

// HasModule returns whether the given config has a module with the given name.
func (c *Config) HasModule(n string) bool {
	return c.GetModule(n) != nil
}

// GetModule returns the module matching the given string or nil if none was
// found.
func (c *Config) GetModule(n string) *Module {
	for _, m := range c.Modules {
		if m.Name == n {
			return &m
		}
	}

	return nil
}

// ListTargets is the list of configurations of the targets from the configuration
// file.
type ListTargets map[string]*Module

// Module defines the configuration parameters of a single module.
// Parity Values => N (None), E (Even), O (Odd)
//
// Default serial settings (inherited from goburrow/serial defaults):
// Baudrate: 19200, Databits: 8, Parity: E, Stopbits: 1
type Module struct {
	Name     string         `yaml:"name"`
	Protocol ModbusProtocol `yaml:"protocol"`
	Timeout  int            `yaml:"timeout"`
	Baudrate int            `yaml:"baudrate"`
	Databits int            `yaml:"databits"`
	Stopbits int            `yaml:"stopbits"`
	Parity   string         `yaml:"parity"`
	Metrics  []MetricDef    `yaml:"metrics"`
}

// RegisterAddr specifies the register in the possible output of _digital
// output_, _digital input, _ananlog input, _analog output_.
type RegisterAddr uint32

// ModbusDataType is an Enum, representing the possible data types a register
// value can be interpreted as.
type ModbusDataType string

func (t *ModbusDataType) validate() error {
	possibleModbusDataTypes := []ModbusDataType{
		ModbusBool,
		ModbusInt16,
		ModbusUInt16,
		ModbusFloat16,
		ModbusInt32,
		ModbusUInt32,
		ModbusFloat32,
		ModbusInt64,
		ModbusUInt64,
		ModbusFloat64,
	}

	if t == nil {
		return fmt.Errorf("expected data type not to be nil")
	}

	for _, possibleType := range possibleModbusDataTypes {
		if *t == possibleType {
			return nil
		}
	}

	return fmt.Errorf("expected one of the following data types %v but got '%v'",
		possibleModbusDataTypes,
		*t)
}

const (
	ModbusBool    ModbusDataType = "bool"
	ModbusFloat16 ModbusDataType = "float16"
	ModbusInt16   ModbusDataType = "int16"
	ModbusUInt16  ModbusDataType = "uint16"
	ModbusInt32   ModbusDataType = "int32"
	ModbusUInt32  ModbusDataType = "uint32"
	ModbusFloat32 ModbusDataType = "float32"
	ModbusInt64   ModbusDataType = "int64"
	ModbusUInt64  ModbusDataType = "uint64"
	ModbusFloat64 ModbusDataType = "float64"
)

// EndiannessType is an Enum, representing the possible endianness types a register
// value can have.
type EndiannessType string

func (endianness *EndiannessType) validate() error {
	possibleEndiannessTypes := []EndiannessType{
		EndiannessBigEndian,
		EndiannessLittleEndian,
		EndiannessMixedEndian,
		EndiannessYolo,
	}

	for _, possibleEndianness := range possibleEndiannessTypes {
		if *endianness == possibleEndianness {
			return nil
		}
	}
	return fmt.Errorf("expected one of the following endianness types %v but got '%v'",
		possibleEndiannessTypes,
		*endianness)
}

const (
	// EndiannessBigEndian (1 2 3 4)
	EndiannessBigEndian EndiannessType = "big"
	// EndiannessLittleEndian (4 3 2 1)
	EndiannessLittleEndian EndiannessType = "little"
	// EndiannessMixedEndian (2 1 4 3)
	EndiannessMixedEndian EndiannessType = "mixed"
	// EndiannessYolo (3 4 1 2)
	EndiannessYolo EndiannessType = "yolo"
)

// MetricType specifies the Prometheus metric type, see
// https://prometheus.io/docs/concepts/metric_types/ for details.
type MetricType string

func (t *MetricType) validate() error {
	possibleMetricTypes := []MetricType{
		MetricTypeGauge,
		MetricTypeCounter,
	}

	if t == nil {
		return fmt.Errorf("expected metric type not to be nil")
	}

	for _, possibleType := range possibleMetricTypes {
		if *t == possibleType {
			return nil
		}
	}

	return fmt.Errorf("expected one of the following metric types %v but got '%v'",
		possibleMetricTypes,
		*t)
}

const (
	MetricTypeGauge   MetricType = "gauge"
	MetricTypeCounter MetricType = "counter"
)

// MetricDef defines how to construct Prometheus metrics based on one or more
// Modbus registers.
type MetricDef struct {
	// Name of the metric in the Prometheus output format.
	Name string `yaml:"name"`

	// Help text of the metric in the Prometheus output format.
	Help string `yaml:"help"`

	// Labels to be applied to the metric in the Prometheus output format.
	Labels map[string]string `yaml:"labels"`

	Address RegisterAddr `yaml:"address"`

	DataType ModbusDataType `yaml:"dataType"`

	Endianness EndiannessType `yaml:"endianness,omitempty"`

	// Bit offset within the input register to parse. Only valid for boolean data
	// type. The two bytes of a register are interpreted in network order (big
	// endianness). Boolean is determined via `register&(1<<offset)>0`.
	BitOffset *int `yaml:"bitOffset,omitempty"`

	MetricType MetricType `yaml:"metricType"`

	// Scaling factor
	Factor *float64 `yaml:"factor,omitempty"`
}

// Validate semantically validates the given metric definition.
func (d *MetricDef) validate() error {
	if err := d.DataType.validate(); err != nil {
		return fmt.Errorf("invalid metric definition %v: %v", d.Name, err)
	}

	if err := d.MetricType.validate(); err != nil {
		return fmt.Errorf("invalid metric definition %v: %v", d.Name, err)
	}

	// TODO: Does it have to be used with bools though? Or should there be a default?
	if d.BitOffset != nil && d.DataType != ModbusBool {
		return fmt.Errorf("bitPosition can only be used with boolean data type")
	}

	if d.Endianness != "" {
		if err := d.Endianness.validate(); err != nil {
			return fmt.Errorf("invalid endianness definition %v: %v", d.Name, err)
		}
	} else {
		d.Endianness = EndiannessBigEndian
	}

	if d.Factor != nil && d.DataType == ModbusBool {
		return fmt.Errorf("factor cannot be used with boolean data type")
	}

	if d.Factor != nil && *d.Factor == 0.0 {
		return fmt.Errorf("factor cannot be 0")
	}

	return nil
}

// ModbusProtocol specifies the protocol used to retrieve modbus data.
type ModbusProtocol string

const (
	// ModbusProtocolTCPIP represents modbus via TCP/IP.
	ModbusProtocolTCPIP = "tcp/ip"
	// ModbusProtocolSerial represents modbus via Serial.
	ModbusProtocolSerial = "serial"
)

// ModbusProtocolValidationError is returned on invalid or unsupported modbus
// protocol specifications.
type ModbusProtocolValidationError struct {
	e string
}

// Error implements the Golang error interface.
func (e *ModbusProtocolValidationError) Error() string {
	return e.e
}

func (t *ModbusProtocol) validate() error {
	possibleProtocols := []ModbusProtocol{
		ModbusProtocolTCPIP,
		ModbusProtocolSerial,
	}

	if t == nil {
		return fmt.Errorf("expected protocol not to be nil")
	}

	for _, possibleProtocol := range possibleProtocols {
		if *t == possibleProtocol {
			return nil
		}
	}

	return fmt.Errorf("expected one of the following protocols %v but got '%v'",
		possibleProtocols,
		*t)
}

// Validate tries to find inconsistencies in the parameters of a module.
func (s *Module) validate() error {
	var err error

	if protocolErr := s.Protocol.validate(); protocolErr != nil {
		err = multierror.Append(err, protocolErr)
	}

	switch s.Protocol {
	case ModbusProtocolSerial:
		if s.Baudrate < 0 || s.Stopbits < 0 || s.Databits < 0 || s.Timeout < 0 {
			newErr := fmt.Errorf("invalid negative value in target \"%s\"", s.Name)
			err = multierror.Append(err, newErr)
		}
		// Data bits: default, 5, 6, 7 or 8
		if s.Databits != 0 && (s.Databits < 5 || s.Databits > 8) {
			newErr := fmt.Errorf("invalid data bits value in target \"%s\"", s.Name)
			err = multierror.Append(err, newErr)
		}
		// Stop bits: default, 1 or 2
		if s.Stopbits < 0 || s.Stopbits > 2 {
			newErr := fmt.Errorf("invalid stop bits value in target \"%s\"", s.Name)
			err = multierror.Append(err, newErr)
		}
		// Parity: N (None), E (Even), O (Odd)
		if s.Parity != "N" && s.Parity != "E" && s.Parity != "O" &&
			s.Parity != "" {
			newErr := fmt.Errorf("invalid parity value in target \"%s\" "+
				"N (None), E (Even), O (Odd)", s.Name)
			err = multierror.Append(err, newErr)
		}
	// checking the absence of specific parameters for a serial connection
	case ModbusProtocolTCPIP:
		if s.Parity != "" || s.Stopbits != 0 || s.Databits != 0 || s.Baudrate != 0 {
			newErr := fmt.Errorf("invalid argument in target %s, TCP targets don't"+
				"use Parity, Stopbits, Databits or Baudrate", s.Name)
			err = multierror.Append(err, newErr)
		}
	default:
		err = multierror.Append(
			err,
			fmt.Errorf("expected one of '%v' protocol but got '%v'",
				[]string{ModbusProtocolTCPIP, ModbusProtocolSerial},
				s.Protocol,
			))
	}
	// track that error if we have no register definitions
	if len(s.Metrics) == 0 {
		noRegErr := fmt.Errorf("no metric definitions found in module %s", s.Name)
		err = multierror.Append(err, noRegErr)
	}

	for _, def := range s.Metrics {
		if err := def.validate(); err != nil {
			return fmt.Errorf("failed to validate module %v: %v", s.Name, err)
		}
	}

	return err
}
