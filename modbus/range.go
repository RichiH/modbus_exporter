package modbus

import (
	"fmt"
	"sort"

	"github.com/RichiH/modbus_exporter/config"
	"github.com/goburrow/modbus"
)

// Range defines a Modbus range that includes a Modbus function and associated metric definitions.
// metric definitions is a slice of continuous or semi-continuous(based on sensitivity) definition interval.
type Range struct {
	F           modbusFunc
	definitions [][]config.MetricDef
}

// RangeMap represents a mapping of Modbus function codes to corresponding Range objects.
type RangeMap map[uint64]Range

func generateRangeMap(definitions []config.MetricDef, c modbus.Client, sensitivity uint64, blocklist []config.RegisterRange) (rangeMap RangeMap, err error) {
	rangeMap = initializeRangeMap(c)

	sort.Slice(definitions, func(i, j int) bool {
		iAddress, _ := definitions[i].Address.GetModAddress()
		jAddress, _ := definitions[j].Address.GetModAddress()
		return iAddress < jAddress
	})

	for _, definition := range definitions {
		if definition.RangeBlacklist {
			continue
		}
		modFunction, err := definition.Address.GetModFunction()
		if err != nil {
			return rangeMap, fmt.Errorf("can't generate range map: %v", err)
		}
		r, err := validateModFunction(rangeMap, modFunction)
		if err != nil {
			return rangeMap, fmt.Errorf("can't generate range map: %v", err)
		}
		if len(r.definitions) == 0 {
			r.definitions = append(r.definitions, []config.MetricDef{})
		}
		lastDefInterval := r.definitions[len(r.definitions)-1]
		if len(lastDefInterval) == 0 {
			lastDefInterval = append(lastDefInterval, definition)
			r.definitions[len(r.definitions)-1] = lastDefInterval
			rangeMap[modFunction] = r
			continue
		}
		firstDef := lastDefInterval[0]
		lastDef := lastDefInterval[len(lastDefInterval)-1]
		firstDefAddress, err := firstDef.Address.GetModAddress()
		if err != nil {
			return rangeMap, fmt.Errorf("can't generate range map: %v", err)
		}
		lastDefAddress, err := lastDef.Address.GetModAddress()
		if err != nil {
			return rangeMap, fmt.Errorf("can't generate range map: %v", err)
		}
		modAddress, err := definition.Address.GetModAddress()
		if err != nil {
			return rangeMap, fmt.Errorf("can't generate range map: %v", err)
		}
		totalRangeOffset := uint16(modAddress-firstDefAddress) + definition.DataType.Offset()
		if modAddress-lastDefAddress > sensitivity || totalRangeOffset > 2000 || rangeCrossesBlock(lastDefAddress, modAddress, modFunction, blocklist) {
			r.definitions = append(r.definitions, []config.MetricDef{})
		}
		r.definitions[len(r.definitions)-1] = append(r.definitions[len(r.definitions)-1], definition)
		rangeMap[modFunction] = r
	}
	return rangeMap, nil
}

func initializeRangeMap(c modbus.Client) RangeMap {
	return RangeMap{
		1: {F: c.ReadCoils},
		2: {F: c.ReadDiscreteInputs},
		3: {F: c.ReadHoldingRegisters},
		4: {F: c.ReadInputRegisters},
	}
}

func validateModFunction(rangeMap RangeMap, modFunction uint64) (Range, error) {
	rangeObj, ok := rangeMap[modFunction]
	if !ok {
		return Range{}, fmt.Errorf("invalid modFunction: %v", modFunction)
	}
	return rangeObj, nil
}

func rangeCrossesBlock(start, end, modFunction uint64, blocks []config.RegisterRange) bool {
	if start > end {
		start, end = end, start
	}
	for _, b := range blocks {
		startF, err := b.Start.GetModFunction()
		if err != nil {
			continue
		}
		endF, err := b.End.GetModFunction()
		if err != nil {
			continue
		}
		if startF != endF || startF != modFunction {
			continue
		}
		sAddr, _ := b.Start.GetModAddress()
		eAddr, _ := b.End.GetModAddress()
		if sAddr > eAddr {
			sAddr, eAddr = eAddr, sAddr
		}
		if start <= eAddr && end >= sAddr {
			return true
		}
	}
	return false
}

func scrapeMetricRange(r Range) ([]metric, error) {
	var metrics []metric

	for _, definitions := range r.definitions {
		first := definitions[0]
		last := definitions[len(definitions)-1]
		firstAddress, err := first.Address.GetModAddress()
		if err != nil {
			return nil, fmt.Errorf("can't calculate mod address for %s: %v", first.Name, err)
		}
		lastAddress, err := last.Address.GetModAddress()
		if err != nil {
			return nil, fmt.Errorf("can't calculate mod address for %s: %v", last.Name, err)
		}
		lastOffset := last.DataType.Offset()
		totalOffset := uint16(lastAddress-firstAddress) + lastOffset

		modBytes, err := r.F(uint16(firstAddress), totalOffset)
		if err != nil {
			return nil, fmt.Errorf("can't read modbus registers for %s: %v", first.Name, err)
		}

		for _, definition := range definitions {
			modAddress, err := definition.Address.GetModAddress()
			if err != nil {
				return nil, fmt.Errorf("can't calculate mod address for %s: %v", definition.Name, err)
			}
			start := (modAddress - firstAddress) * 2
			end := uint16(start) + (definition.DataType.Offset() * 2)
			defBytes := modBytes[start:end]
			v, err := parseModbusData(definition, defBytes)
			if err != nil {
				return nil, fmt.Errorf("can't parse modbus data for %s: %v", definition.Name, err)
			}
			metrics = append(metrics, metric{definition.Name, definition.Help, definition.Labels, v, definition.MetricType})
		}
	}

	return metrics, nil
}
