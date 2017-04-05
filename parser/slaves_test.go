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

package parser

// TODO create a REAL test
// var input = config.ListSlaves{
// 	"arduino": config.Slave{
// 		Coils:          []string{"2:28 = led$..."},
// 		InputRegisters: []string{"2,3,4,5 = relayA$, relayA$, relayB$, relay$"},
// 	},
// 	"raspberry pi": config.Slave{
// 		Coils:            []string{"200:204 = light..."},
// 		HoldingRegisters: []string{"55:60"},
// 		DiscreteInputs:   []string{"60 = Motor"},
// 	},
// }
//
// func TestSlaveParser(t *testing.T) {
// 	sp := new(SlaveParser)
// 	sp.Init(&input)
// 	if list, err := sp.parse(); err == nil {
// 		fmt.Println(list)
// 	} else {
// 		fmt.Printf("error %v\n", err)
// 	}
// }
