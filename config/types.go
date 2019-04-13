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

// ParsedSlave contains all the I/O registers of one slave
type ParsedSlave struct {
	Name          string
	DigitalInput  []Register
	DigitalOutput []Register
	AnalogInput   []Register
	AnalogOutput  []Register
}

// Register is the representation of a single register
type Register struct {
	Name    string
	Address uint16
}
