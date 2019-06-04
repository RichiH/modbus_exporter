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
	"net"
	"strconv"
	"strings"

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

// RegType is a helper type to obtain the name of the register types
type RegType int

const (
	// DigitalInput identifies the digital input value
	DigitalInput RegType = iota
	// DigitalOutput identifies the digital output value
	DigitalOutput
	// AnalogInput identifies the analog input value
	AnalogInput
	// AnalogOutput identifies the analog output value
	AnalogOutput
)

func (r RegType) String() string {
	var s string
	switch r {
	case DigitalInput:
		s = "DIn"
	case DigitalOutput:
		s = "DOut"
	case AnalogInput:
		s = "AIn"
	case AnalogOutput:
		s = "AOut"
	}
	return s
}

// ListTargets is the list of configurations of the targets from the configuration
// file.
type ListTargets map[string]*Module

// Module defines the configuration parameters of a single target.
// Parity Values => N (None), E (Even), O (Odd)
//
// Default serial:
// Baudrate: 19200, Databits: 8, Stopbits: 1, Parity: E
type Module struct {
	Name          string         `yaml:"name"`
	Protocol      ModbusProtocol `yaml:"protocol"`
	Timeout       int            `yaml:"timeout"`
	Baudrate      int            `yaml:"baudrate"`
	Databits      int            `yaml:"databits"`
	Stopbits      int            `yaml:"stopbits"`
	Parity        string         `yaml:"parity"`
	ID            byte           `yaml:"id"`
	KeepAlive     bool           `yaml:"keepAlive"`
	DigitalInput  []MetricDef    `yaml:"digitalIn"`
	DigitalOutput []MetricDef    `yaml:"digitalOut"`
	AnalogInput   []MetricDef    `yaml:"analogIn"`
	AnalogOutput  []MetricDef    `yaml:"analogOut"`
}

// RegisterAddr specifies the register in the possible output of _digital
// output_, _digital input, _ananlog input, _analog output_.
type RegisterAddr uint16

// ModbusDataType is an Enum, representing the possible data types a register
// value can be interpreted as.
type ModbusDataType string

func (t *ModbusDataType) validate() error {
	if t == nil {
		return fmt.Errorf("expected data type not to be nil")
	}

	for _, possibelType := range possibelModbusDataTypes {
		if *t == possibelType {
			return nil
		}
	}

	return fmt.Errorf("expected one of the following data types %v but got '%v'", possibelModbusDataTypes, *t)
}

const (
	ModbusFloat16 ModbusDataType = "float16"
	ModbusInt16   ModbusDataType = "int16"
	ModbusUInt16  ModbusDataType = "uint16"
	ModbusBool    ModbusDataType = "bool"
)

var possibelModbusDataTypes = []ModbusDataType{
	ModbusFloat16,
	ModbusInt16,
	ModbusUInt16,
	ModbusBool,
}

// MetricDef defines how to construct Prometheus metrics based on one or more
// Modbus registers.
type MetricDef struct {
	Name string `yaml:"name"`

	Address RegisterAddr `yaml:"address"`
	// Index within the register byte slice, only applicable for ModbusBool.
	Index int8 `yaml:"index"`

	DataType ModbusDataType `yaml:"dataType"`
	// Bit offset within the input register to parse. Only valid for boolean data
	// type. The two bytes of a register are interpreted in network order (big
	// endianness). Boolean is determined via `register&(1<<offset)>0`.
	BitOffset *int `yaml:"bitOffset,omitempty"`
}

// Validate semantically validates the given metric definition.
func (d *MetricDef) validate() error {
	if err := d.DataType.validate(); err != nil {
		return fmt.Errorf("invalid metric definition %v: %v", d.Name, err)
	}

	// TODO: Does it have to be used with bools though? Or should there be a default?
	if d.BitOffset != nil && d.DataType != ModbusBool {
		return fmt.Errorf("bitPosition can only be used with boolean data type")
	}

	return nil
}

// PrettyString prints only the initialized values
func (s *Module) PrettyString() string {
	res := ""
	if s.ID != 0 {
		res += fmt.Sprintf(", ID: %v", s.ID)
	}
	if s.Timeout != 0 {
		res += fmt.Sprintf(", Timeout: %v", s.Timeout)
	}
	if s.Baudrate != 0 {
		res += fmt.Sprintf(", Baudrate: %v", s.Baudrate)
	}
	if s.Databits != 0 {
		res += fmt.Sprintf(", Databits: %v", s.Databits)
	}
	if s.Stopbits != 0 {
		res += fmt.Sprintf(", Stopbits: %v", s.Stopbits)
	}
	if s.Parity != "" {
		res += fmt.Sprintf(", Parity: %v", s.Parity)
	}
	res += "}"
	return res
}

// ModbusProtocol specifies the protocol used to retrieve modbus data.
type ModbusProtocol string

const (
	// ModbusProtocolTCPIP represents modbus via TCP/IP.
	ModbusProtocolTCPIP = "tcp/ip"
	// ModbusProtocolSerial represents modbus via Serial.
	ModbusProtocolSerial = "serial"
)

var (
	serialPrefix = [...]string{"/dev/ttyACM", "/dev/ttyUSB", "/dev/ttyS"}
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

// CheckPortTarget indetifies the port as a PortType in order to identify the type
// of connection to stqablish in the Modbus Manager. Returns Invalid or IP, and
// Invalid when the Port property has an inidentifiable content.
func CheckPortTarget(t string) (ModbusProtocol, *ModbusProtocolValidationError) {
	var prefixSerial string
	isSerial := false
	for i := 0; i < len(serialPrefix) && !isSerial; i++ {
		prefixSerial = serialPrefix[i]
		isSerial = strings.HasPrefix(t, prefixSerial)
	}
	// checks if it's a correct port
	if isSerial && len(t) > len(prefixSerial) {
		portNumber := t[len(prefixSerial):]
		if v, err := strconv.Atoi(portNumber); err == nil && v >= 0 {
			return ModbusProtocolSerial, nil
		}
	}
	// checks if it's a correct IP
	if i := strings.LastIndex(t, ":"); i > -1 {
		_, err := strconv.Atoi(t[i+1:])
		if err != nil {
			return "", &ModbusProtocolValidationError{fmt.Sprintf("failed to validate IP '%v': %v", t, err)}
		}
		IPv4 := net.ParseIP(t[:i]).To4()
		if IPv4 != nil || t[:i] == "localhost" {
			return ModbusProtocolTCPIP, nil
		}
	}
	return "", &ModbusProtocolValidationError{fmt.Sprintf("failed to extract modbus protocol from '%v'", t)}
}

func (s *Module) hasRegisterDefinitions() bool {
	return len(s.DigitalInput) != 0 || len(s.DigitalOutput) != 0 ||
		len(s.AnalogInput) != 0 || len(s.AnalogOutput) != 0
}

// Validate tries to find inconsistencies in the parameters of a Target.
// The port must be valid. If present:
// -Baudrate and Timeout must be positive.
// -Stopbits must be 1 or 2.
// -Databits must be 5, 6, 7 or 8.
// -Parity has to be "N", "E" or "O". The use of no parity requires 2 stop bits.
func (s *Module) validate() error {
	var err error

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
		if s.Stopbits > 2 {
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
		// The use of no parity requires 2 stop bits.
		// if s.Parity == "N" && s.Stopbits != 2 {
		// 	newErr := fmt.Errorf("the use of no parity requires 2 stop bits in "+
		// 		"target \"%s\"", s.Name)
		// 	err = multierror.Append(err, newErr)
		// }
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
	if !s.hasRegisterDefinitions() {
		noRegErr := fmt.Errorf("no register definition found in target %s", s.Name)
		err = multierror.Append(err, noRegErr)
	}

	for _, defs := range [][]MetricDef{s.DigitalInput, s.DigitalOutput, s.AnalogInput, s.AnalogOutput} {
		for _, def := range defs {
			if err := def.validate(); err != nil {
				return fmt.Errorf("failed to validate target %v: %v", s.Name, err)
			}
		}
	}

	return err
}
