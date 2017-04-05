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

package config

import (
	"fmt"
	"net"
	"strconv"
	"strings"

	multierror "github.com/hashicorp/go-multierror"
)

// ListSlaves is the list of configurations of the slaves from the configuration
// file.
type ListSlaves map[string]*Slave

// Slave defines the configuration parameters of a single slave.
// Parity Values => N (None), E (Even), O (Odd)
//
// Default serial:
// Baudrate: 19200, Databits: 8, Stopbits: 1, Parity: E
type Slave struct {
	Port          string   `yaml:"port"`
	ID            byte     `yaml:"id"`
	Timeout       int      `yaml:"timeout"`
	Baudrate      int      `yaml:"baudrate"`
	Databits      int      `yaml:"databits"`
	Stopbits      int      `yaml:"stopbits"`
	Parity        string   `yaml:"parity"`
	DigitalInput  []string `yaml:"DigitalIn"`
	DigitalOutput []string `yaml:"DigitalOut"`
	AnalogInput   []string `yaml:"AnalogIn"`
	AnalogOutput  []string `yaml:"AnalogOut"`
}

// PrettyString prints only the initialized values
func (s *Slave) PrettyString() string {
	res := "{Port: " + s.Port
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

// PortType represents the type of the port of a Slave.
type PortType int

func (p PortType) String() string {
	return portNames[p]
}

const (
	// IP is an IPv4 port
	IP PortType = iota
	// Serial is an USB port
	Serial
	// Invalid is a not valid port
	Invalid
)

var (
	portNames    = [...]string{"IP", "serial", "invalid"}
	serialPrefix = [...]string{"/dev/ttyACM", "/dev/ttyUSB", "/dev/ttyS"}
)

// CheckPortSlave indetifies the port as a PortType in order to identify the type
// of connection to stqablish in the Modbus Manager. Returns Invalid or IP, and
// Invalid when the Port property has an inidentifiable content.
func CheckPortSlave(s *Slave) PortType {
	var prefixSerial string
	isSerial := false
	for i := 0; i < len(serialPrefix) && !isSerial; i++ {
		prefixSerial = serialPrefix[i]
		isSerial = strings.HasPrefix(s.Port, prefixSerial)
	}
	// checks if it's a correct port
	if isSerial && len(s.Port) > len(prefixSerial) {
		portNumber := s.Port[len(prefixSerial):]
		if v, err := strconv.Atoi(portNumber); err == nil && v >= 0 {
			return Serial
		}
	}
	// checks if it's a correct IP
	if i := strings.LastIndex(s.Port, ":"); i > -1 {
		_, err := strconv.Atoi(s.Port[i+1:])
		if err != nil {
			return Invalid
		}
		IPv4 := net.ParseIP(s.Port[:i]).To4()
		if IPv4 != nil || s.Port[:i] == "localhost" {
			return IP
		}
	}
	return Invalid
}

func (s *Slave) hasRegisterDefinitions() bool {
	return len(s.DigitalInput) != 0 || len(s.DigitalOutput) != 0 ||
		len(s.AnalogInput) != 0 || len(s.AnalogOutput) != 0
}

// ValidateSlave tries to find inconsistencies in the parameters of a Slave.
// The port must be valid. If present:
// -Baudrate and Timeout must be positive.
// -Stopbits must be 1 or 2.
// -Databits must be 5, 6, 7 or 8.
// -Parity has to be "N", "E" or "O". The use of no parity requires 2 stop bits.
func ValidateSlave(s *Slave, alias string) error {
	var err error
	if p := CheckPortSlave(s); p == Invalid {
		newErr := fmt.Errorf("Invalid port \"%s\" in slave \"%s\"", s.Port, alias)
		err = multierror.Append(err, newErr)
	}
	if s.Baudrate < 0 || s.Stopbits < 0 || s.Databits < 0 || s.Timeout < 0 {
		newErr := fmt.Errorf("Invalid negative value in slave \"%s\"", alias)
		err = multierror.Append(err, newErr)
	}
	// Data bits: default, 5, 6, 7 or 8
	if s.Databits != 0 && (s.Databits < 5 || s.Databits > 8) {
		newErr := fmt.Errorf("Invalid data bits value in slave \"%s\"", alias)
		err = multierror.Append(err, newErr)
	}
	// Stop bits: default, 1 or 2
	if s.Stopbits > 2 {
		newErr := fmt.Errorf("Invalid stop bits value in slave \"%s\"", alias)
		err = multierror.Append(err, newErr)
	}
	// Parity: N (None), E (Even), O (Odd)
	if s.Parity != "N" && s.Parity != "E" && s.Parity != "O" &&
		s.Parity != "" {
		newErr := fmt.Errorf("Invalid parity value in slave \"%s\" "+
			"N (None), E (Even), O (Odd)", alias)
		err = multierror.Append(err, newErr)
	}
	// The use of no parity requires 2 stop bits.
	if s.Parity == "N" && s.Stopbits != 2 {
		newErr := fmt.Errorf("The use of no parity requires 2 stop bits in "+
			"slave \"%s\"", alias)
		err = multierror.Append(err, newErr)
	}
	// track that error if we have no register definitions
	if !s.hasRegisterDefinitions() {
		noRegErr := fmt.Errorf("no register definition found in slave %s", alias)
		err = multierror.Append(err, noRegErr)
	}
	return err
}
