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

package parser

import (
	"strconv"

	"github.com/lupoDharkael/modbus_exporter/token"
)

// regName represents the information of a single register's name
type regName struct {
	content    string
	numeration bool
}

// slaveExprHandler extracts the list of names and registers from the tokenized
// input from the Scanner type when parsing the slaves configuration
type slaveExprHandler struct {
	names       []regName
	regs        []uint16
	handleToken state
	repeatNames bool
}

type state func(token.Token, string) string

// init prepares the initial status of the slaveExprHandler checking the first case
// and setting the first state function, if it fails returns an error and in
// that case the parse is unable to be finished.
func (s *slaveExprHandler) init(tok token.Token, lit string, _ int) string {
	var errMsg string
	s.names = make([]regName, 0)
	s.regs = make([]uint16, 0)
	s.repeatNames = false
	s.handleToken = nil
	// initialize  from first token
	if tok == token.INT {
		if i, err := strconv.Atoi(lit); err == nil {
			s.regs = append(s.regs, uint16(i))
		}
		s.handleToken = s.rangeOrEnumState
	} else {
		errMsg = "expected INT value as first parameter"
	}
	return errMsg
}

// all the possible states. They represent the states in the process of parsing a
// single expression from the slave's definition.

func (s *slaveExprHandler) rangeOrEnumState(tok token.Token, lit string) string {
	var errMsg string
	switch tok {
	case token.COMMA:
		s.handleToken = s.commaToIntState
	case token.COLON:
		s.handleToken = s.colonToIntState
	case token.ASSIGN:
		s.handleToken = s.equalsToStringState
	default:
		errMsg = "unexpected parameter after first register definition"
	}
	return errMsg
}

func (s *slaveExprHandler) equalsToStringState(tok token.Token, lit string) string {
	var errMsg string
	if tok == token.IDENT {
		s.names = append(s.names, regName{content: lit})
		s.handleToken = s.stringToCommaState
	} else {
		errMsg = "expected name definition after the `=` symbol"
	}
	return errMsg
}

func (s *slaveExprHandler) colonToIntState(tok token.Token, lit string) string {
	var errMsg string
	if tok == token.INT {
		// check a < b in a:b and adds all the elements of the range to the slice
		top, err := strconv.Atoi(lit)
		if err == nil && s.regs[0] < uint16(top) {
			for i := s.regs[0] + 1; i <= uint16(top); i++ {
				s.regs = append(s.regs, i)
			}
		} else {
			errMsg = "wrong value in the second member in range definition"
		}
		s.handleToken = s.intToEqualsState
	} else {
		errMsg = "expected name definition after the `=` symbol"
	}
	return errMsg
}

func (s *slaveExprHandler) commaToIntState(tok token.Token, lit string) string {
	var errMsg string
	if tok == token.INT {
		if i, err := strconv.Atoi(lit); err == nil {
			s.regs = append(s.regs, uint16(i))
		}
		s.handleToken = s.intToCommaState
	} else {
		errMsg = "expected register value after comma"
	}
	return errMsg
}

func (s *slaveExprHandler) intToEqualsState(tok token.Token, lit string) string {
	var errMsg string
	if tok == token.ASSIGN {
		s.handleToken = s.equalsToStringState
	} else {
		errMsg = "expected = value after range definition"
	}
	return errMsg
}

func (s *slaveExprHandler) commaToStringState(tok token.Token, lit string) string {
	var errMsg string
	if tok == token.IDENT {
		s.names = append(s.names, regName{content: lit})
		s.handleToken = s.stringToCommaState
	} else {
		errMsg = "expected name definition after comma"
	}
	return errMsg
}

func (s *slaveExprHandler) stringToCommaState(tok token.Token, lit string) string {
	var errMsg string
	switch tok {
	case token.COMMA:
		s.handleToken = s.commaToStringState
	case token.ELLIPSIS:
		s.repeatNames = true
		s.handleToken = s.endedState
	case token.ITERATOR:
		s.names[len(s.names)-1].numeration = true
		s.handleToken = s.dollarToCommaState
	default:
		errMsg = "invalid value in the name enumeration"
	}
	return errMsg
}

func (s *slaveExprHandler) dollarToCommaState(tok token.Token, lit string) string {
	var errMsg string
	switch tok {
	case token.COMMA:
		s.handleToken = s.commaToStringState
	case token.ELLIPSIS:
		s.repeatNames = true
		s.handleToken = s.endedState
	default:
		errMsg = "invalid value in the name enumeration"
	}
	return errMsg
}

func (s *slaveExprHandler) intToCommaState(tok token.Token, lit string) string {
	var errMsg string
	switch tok {
	case token.COMMA:
		s.handleToken = s.commaToIntState
	case token.ASSIGN:
		s.handleToken = s.equalsToStringState
	default:
		errMsg = "invalid value after register enumeration, expected comma or assign"
	}
	return errMsg
}

func (s *slaveExprHandler) endedState(tok token.Token, lit string) string {
	return "illegal arguments after ellipsis"
}
