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

import "testing"

var (
	valuesTest = [...]string{
		"localhost:8080",
		"192.168.0.23:8080",
		"192.168.0.3333.043",
		":7070",
		"300.34.23.2:6767",
		"/dev/ttyS4sw34",
		"/dev",
		"/dev/ttyUSB0",
		"/dev/ttyS0",
	}
	expectedTest = [...]PortType{
		IP,
		IP,
		Invalid,
		Invalid,
		Invalid,
		Invalid,
		Invalid,
		Serial,
		Serial,
	}
)

func TestCheckPort(t *testing.T) {
	s := new(Target)
	for i, v := range valuesTest {
		s.Port = v
		pType := CheckPortTarget(*s)
		if expectedTest[i] != pType {
			t.Errorf("at port %s, expected %s, got %s.", v, expectedTest[i], pType)
		}
	}
}

var (
	slavesBad = [...]Target{
		Target{Port: "localhost:8080", Parity: "abc"},
		Target{Port: "localhost:8080", Parity: "N"},
		Target{Port: "localhost:8080", Stopbits: 4},
		Target{Port: "localhost:8080", Baudrate: -1},
		Target{Port: "localhost:8080", Databits: 50},
		Target{Port: "localhost:8080", Baudrate: -1},
	}
	regDefTest = []MetricDef{
		{
			Name:     "test",
			Address:  34,
			DataType: "int16",
		},
	}
	slavesGood = [...]Target{
		Target{Port: "localhost:8080", DigitalOutput: regDefTest},
	}
)

func TestValidate(t *testing.T) {
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
	s := Target{Port: "localhost:8080", Parity: "O", Stopbits: 1, Databits: 7}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		s.PrettyString()
	}
}
