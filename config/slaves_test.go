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
	s := new(Slave)
	for i, v := range valuesTest {
		s.Port = v
		pType := CheckPortSlave(s)
		if expectedTest[i] != pType {
			t.Errorf("at port %s, expected %s, got %s.", v, expectedTest[i], pType)
		}
	}
}

var (
	slavesBad = [...]Slave{
		Slave{Port: "localhost:8080", Parity: "abc"},
		Slave{Port: "localhost:8080", Parity: "N"},
		Slave{Port: "localhost:8080", Stopbits: 4},
		Slave{Port: "localhost:8080", Baudrate: -1},
		Slave{Port: "localhost:8080", Databits: 50},
		Slave{Port: "localhost:8080", Baudrate: -1},
	}
	regDefTest = []string{"34 = test"}
	slavesGood = [...]Slave{
		Slave{Port: "localhost:8080", DigitalOutput: regDefTest},
	}
)

func TestValidate(t *testing.T) {
	for _, s := range slavesGood {
		if err := ValidateSlave(&s, "TestedSlave"); err != nil {
			t.Errorf("validation of %v expected to pass but received the error:\n"+
				"%s", s.PrettyString(), err)
		}
	}
	for _, s := range slavesBad {
		if err := ValidateSlave(&s, "TestedSlave"); err == nil {
			t.Errorf("validation of %v expected to fail but it didn't.",
				s.PrettyString())
		}
	}
}

func BenchmarkPrettyPrint(b *testing.B) {
	s := Slave{Port: "localhost:8080", Parity: "O", Stopbits: 1, Databits: 7}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		s.PrettyString()
	}
}
