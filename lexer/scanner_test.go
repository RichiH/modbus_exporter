// Copyright 2017 Alejandro Sirgo Rica
//
// This file is part of GryphOn.
//
//     GryphOn is free software: you can redistribute it and/or modify
//     it under the terms of the GNU General Public License as published by
//     the Free Software Foundation, either version 3 of the License, or
//     (at your option) any later version.
//
//     GryphOn is distributed in the hope that it will be useful,
//     but WITHOUT ANY WARRANTY; without even the implied warranty of
//     MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
//     GNU General Public License for more details.
//
//     You should have received a copy of the GNU General Public License
//     along with GryphOn.  If not, see <http://www.gnu.org/licenses/>.

// part of the code in this file belongs to the Go project, check the
// license_compliance file in this directory.

package lexer

import (
	"fmt"
	"testing"

	"github.com/lupoDharkael/modbus_exporter/token"
)

// TODO create a real test
var src = `-34:56 = sensor$, sensor$, sensor$
2:5 = led$...
66 = otherSensor
2,3,4,5 = relay$, relay$, relay$, relay$
#----
INPUT slave1[48] <= 200 || INPUT slave1[43] <= 200 && INPUT slave1[4] <= 200
    INPUT slave1[50] <= 200

!(OUTPUT slave1[46] == 200)
`

func TestScanner(t *testing.T) {
	s := new(Scanner)
	fmt.Printf("Input:\n\"%s\"\n\nToken Output:\n", src)
	s.Init("testing", []byte(src))
	for t, lit, _ := s.Scan(); t != token.EOF; t, lit, _ = s.Scan() {
		fmt.Println(t, "  ", lit)
	}
	println("\n")
	if s.GetReport() != nil {
		fmt.Println(s.GetReport())
	}
}
