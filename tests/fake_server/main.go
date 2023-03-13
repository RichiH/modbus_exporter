// Copyright 2019 Richard Hartmann
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"log"
	"time"

	"github.com/tbrandon/mbserver"
)

const address = "127.0.0.1:1502"

func main() {
	serv := mbserver.NewServer()
	serv.HoldingRegisters[22] = uint16(240)
	serv.HoldingRegisters[23] = uint16(250)
	serv.Coils[24] = byte(1)
	err := serv.ListenTCP(address)
	if err != nil {
		log.Printf("%v\n", err)
	}

	log.Printf("listening on %v", address)

	defer serv.Close()

	// Wait forever
	for {
		time.Sleep(1 * time.Second)
	}
}
