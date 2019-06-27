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
