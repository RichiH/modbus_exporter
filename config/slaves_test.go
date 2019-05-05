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
	"strconv"
	"testing"
)

func TestCheckPort(t *testing.T) {
	tests := []struct {
		input         string
		protocol      ModbusProtocol
		expectedError error
	}{
		{
			"localhost:8080",
			ModbusProtocolTCPIP,
			nil,
		},
		{
			"192.168.0.23:8080",
			ModbusProtocolTCPIP,
			nil,
		},
		{
			"192.168.0.3333.043",
			"",
			&ModbusProtocolValidationError{},
		},
		{":7070", "", &ModbusProtocolValidationError{}},
		{"300.34.23.2:6767", "", &ModbusProtocolValidationError{}},
		{"/dev/ttyS4sw34", "", &ModbusProtocolValidationError{}},
		{"/dev", "", &ModbusProtocolValidationError{}},
		{"/dev/ttyUSB0", ModbusProtocolSerial, nil},
		{"/dev/ttyS0", ModbusProtocolSerial, nil},
	}
	for i, loopTest := range tests {
		test := loopTest

		t.Run(strconv.Itoa(i), func(t *testing.T) {
			protocol, err := CheckPortTarget(test.input)

			if test.expectedError == nil {
				if test.protocol != protocol {
					t.Fatalf("expected protocol %v but got %v", test.protocol, protocol)
				}

				if err != nil {
					t.Fatalf("expected no error but got %v", err)
				}

				return
			}
		})
	}
}

func TestValidate(t *testing.T) {
	var (
		slavesBad = [...]Module{
			{Parity: "abc"},
			{Parity: "N"},
			{Stopbits: 4},
			{Baudrate: -1},
			{Databits: 50},
			{Baudrate: -1},
		}
		regDefTest = []MetricDef{
			{
				Name:     "test",
				Address:  34,
				DataType: "int16",
			},
		}
		slavesGood = [...]Module{
			{DigitalOutput: regDefTest, Protocol: ModbusProtocolTCPIP},
		}
	)

	for _, s := range slavesGood {
		if err := s.validate(); err != nil {
			t.Errorf("validation of %v expected to pass but received the error:\n"+
				"%s", s.PrettyString(), err)
		}
	}
	for _, s := range slavesBad {
		if err := s.validate(); err == nil {
			t.Errorf("validation of %v expected to fail but it didn't.",
				s.PrettyString())
		}
	}
}

func BenchmarkPrettyPrint(b *testing.B) {
	s := Module{Parity: "O", Stopbits: 1, Databits: 7}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		s.PrettyString()
	}
}
