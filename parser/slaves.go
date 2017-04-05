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

// Package parser contains all the tools needed to parse the configuration files.
package parser

import (
	"fmt"
	"sort"

	"github.com/lupoDharkael/modbus_exporter/config"
	"github.com/lupoDharkael/modbus_exporter/lexer"
	"github.com/lupoDharkael/modbus_exporter/token"

	multierror "github.com/hashicorp/go-multierror"
)

// ParseSlaves parses the list of slaves returned by config.LoadSlaves(), returns
// a nil list and an error if something went wrong
func ParseSlaves(listSlaves config.ListSlaves) ([]config.ParsedSlave, error) {
	sp := new(SlaveParser)
	sp.Init(listSlaves)
	return sp.parse()
}

// SlaveParser is the parser for the type config.ListSlaves. It requires to be
// initialized with the init method. To parse the asigned config.ListSlaves you
// may use the Parse() method and it will return a list of registers in the for
// of Register datatype with a name and the value of register. If you want to
// retrieve the error from the process of parsing you have the GetReport() method.
type SlaveParser struct {
	nameTracker map[string]int
	listSlaves  config.ListSlaves
	result      []config.ParsedSlave

	err     *multierror.Error
	context *context
}

// context is a type defined to save information about the actual set of registers
// being parsed by SlaveParser.
type context struct {
	slave string
	kind  config.RegType
}

func (c *context) String() string {
	return fmt.Sprintf("%s: %s", c.slave, c.kind)
}

// Init prepares a SlaveParser to work
func (p *SlaveParser) Init(listSlaves config.ListSlaves) {
	p.nameTracker = make(map[string]int)
	p.listSlaves = listSlaves
	p.context = new(context)
	p.result = make([]config.ParsedSlave, 0, len(listSlaves))
}

// GetReport returns the errors of the parser as a common error interface
func (p *SlaveParser) GetReport() error {
	return p.err.ErrorOrNil()
}

func (p *SlaveParser) enumeratedName(name string) string {
	p.nameTracker[name]++
	number := p.nameTracker[name]
	return fmt.Sprintf("%s%v", name, number)
}

// helper method for informative error generation
func (p *SlaveParser) error(msg, line string) {
	if len(line) > 64 {
		line = line[:64] + "..."
	}
	errorWithContext := fmt.Errorf("[%s] %s in line: \"%s\"", p.context, msg, line)
	p.err = multierror.Append(p.err, errorWithContext)
}

// Parse process the unmarshaled slave configuration in order to extract the data
// of every register.
func (p *SlaveParser) parse() ([]config.ParsedSlave, error) {
	for name, slaveConf := range p.listSlaves {
		if valErr := config.ValidateSlave(slaveConf, name); valErr != nil {
			p.err = multierror.Append(p.err, valErr)
			continue
		}
		parsedSlave := new(config.ParsedSlave)
		p.context.slave = name
		parsedSlave.Name = name
		p.context.kind = config.DigitalInput
		parsedSlave.DigitalInput = p.parseRegisters(slaveConf.DigitalInput)
		p.context.kind = config.DigitalOutput
		parsedSlave.DigitalOutput = p.parseRegisters(slaveConf.DigitalOutput)
		p.context.kind = config.AnalogInput
		parsedSlave.AnalogInput = p.parseRegisters(slaveConf.AnalogInput)
		p.context.kind = config.AnalogOutput
		parsedSlave.AnalogOutput = p.parseRegisters(slaveConf.AnalogOutput)
		p.result = append(p.result, *parsedSlave)
	}
	if p.GetReport() != nil {
		p.result = nil
	}
	return p.result, p.GetReport()
}

func (p *SlaveParser) parseRegisters(regDefList []string) (regsList []config.Register) {
	if len(regDefList) == 0 {
		return nil
	}
	repeatedTracker := make(map[uint16]bool)
	scanner := new(lexer.Scanner)
	parseHelper := new(slaveExprHandler)
	regsList = make([]config.Register, 0, len(regDefList))
	// iterate over te list of slave definitions
	for _, regsDef := range regDefList {
		// init token generator
		scanner.Init(p.context.String(), []byte(regsDef))
		// stage 1 of parsing, extracts data into the parseHelper inner types in
		// each iteration
		errMsg := parseHelper.init(scanner.Scan())
		if errMsg != "" {
			p.error(errMsg, regsDef)
		}
		// iterates over all the data, the parsing stops after the first error
		// but it keeps iterating to get all the scanner errors.
		for tok, lit, _ := scanner.Scan(); tok != token.EOF; tok, lit, _ = scanner.Scan() {
			if errMsg == "" && scanner.GetReport() == nil {
				errMsg = parseHelper.handleToken(tok, lit)
				if errMsg != "" {
					p.error(errMsg, regsDef)
				}
			}
		}
		p.err = multierror.Append(p.err, scanner.GetReport())
		// early exit if we found a parsing error, skips creation of registers
		if len(parseHelper.names) > len(parseHelper.regs) {
			p.error("more names than registers", regsDef)
		}
		if p.err.ErrorOrNil() != nil {
			return nil
		}
		// creates the registers
		for i := 0; i < len(parseHelper.regs); i++ {
			if repeated := repeatedTracker[parseHelper.regs[i]]; !repeated {
				repeatedTracker[parseHelper.regs[i]] = true
			} else {
				p.error("repeated register", regsDef)
			}
			// when we have names left for the following registers
			if i < len(parseHelper.names) || parseHelper.repeatNames {
				var name regName
				// repeat the available names over the regs (ellipsis)
				if parseHelper.repeatNames {
					name = parseHelper.names[i%len(parseHelper.names)]
				} else {
					name = parseHelper.names[i]
				}
				// when the name requires automatic numeration
				if name.numeration {
					name.content = p.enumeratedName(name.content)
				}
				newReg := config.Register{Name: name.content, Value: parseHelper.regs[i]}
				regsList = append(regsList, newReg)
				// automatic naming when no names left but we find more registers
			} else {
				genName := autonaming(p.context.kind, parseHelper.regs[i])
				newReg := config.Register{Name: genName, Value: parseHelper.regs[i]}
				regsList = append(regsList, newReg)
			}
		}
	}
	// order by register value
	sort.Slice(regsList, func(i int, j int) bool {
		return regsList[i].Value < regsList[j].Value
	})
	return
}

// autonaming obtains the name for unnamed registers. It obtains the short version
// of the register type as digital/analog input/output.
func autonaming(context config.RegType, reg uint16) string {
	return fmt.Sprintf("%s_%v", context, reg)
}
