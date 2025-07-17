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
	ReadFunction      modbusFunc
	metricDefinitions [][]config.MetricDef
}

// RangeMap represents a mapping of Modbus function codes to corresponding Range objects.
type RangeMap map[uint64]Range

func generateRangeMap(metricDefinitions []config.MetricDef, client modbus.Client, sensitivity uint64, blockedRanges []config.RegisterRange) (rangeMap RangeMap, err error) {
	rangeMap = initializeRangeMap(client)

	sort.Slice(metricDefinitions, func(i, j int) bool {
		iAddress, _ := metricDefinitions[i].Address.GetModAddress()
		jAddress, _ := metricDefinitions[j].Address.GetModAddress()
		return iAddress < jAddress
	})

	for _, definition := range metricDefinitions {
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
		if len(r.metricDefinitions) == 0 {
			r.metricDefinitions = append(r.metricDefinitions, []config.MetricDef{})
		}
		lastDefinitionInterval := r.metricDefinitions[len(r.metricDefinitions)-1]
		if len(lastDefinitionInterval) == 0 {
			lastDefinitionInterval = append(lastDefinitionInterval, definition)
			r.metricDefinitions[len(r.metricDefinitions)-1] = lastDefinitionInterval
			rangeMap[modFunction] = r
			continue
		}
		firstDefinition := lastDefinitionInterval[0]
		lastDefinition := lastDefinitionInterval[len(lastDefinitionInterval)-1]
		firstDefinitionAddress, err := firstDefinition.Address.GetModAddress()
		if err != nil {
			return rangeMap, fmt.Errorf("can't generate range map: %v", err)
		}
		lastDefinitionAddress, err := lastDefinition.Address.GetModAddress()
		if err != nil {
			return rangeMap, fmt.Errorf("can't generate range map: %v", err)
		}
		modAddress, err := definition.Address.GetModAddress()
		if err != nil {
			return rangeMap, fmt.Errorf("can't generate range map: %v", err)
		}
		totalRangeOffset := uint16(modAddress-firstDefinitionAddress) + definition.DataType.Offset()
		if modAddress-lastDefinitionAddress > sensitivity || totalRangeOffset > 2000 || rangeCrossesBlock(lastDefinitionAddress, modAddress, modFunction, blockedRanges) {
			r.metricDefinitions = append(r.metricDefinitions, []config.MetricDef{})
		}
		r.metricDefinitions[len(r.metricDefinitions)-1] = append(r.metricDefinitions[len(r.metricDefinitions)-1], definition)
		rangeMap[modFunction] = r
	}
	return rangeMap, nil
}

func initializeRangeMap(client modbus.Client) RangeMap {
	return RangeMap{
		1: {ReadFunction: client.ReadCoils},
		2: {ReadFunction: client.ReadDiscreteInputs},
		3: {ReadFunction: client.ReadHoldingRegisters},
		4: {ReadFunction: client.ReadInputRegisters},
	}
}

func validateModFunction(rangeMap RangeMap, modFunction uint64) (Range, error) {
	rangeObj, ok := rangeMap[modFunction]
	if !ok {
		return Range{}, fmt.Errorf("invalid modFunction: %v", modFunction)
	}
	return rangeObj, nil
}

func rangeCrossesBlock(start, end, modFunction uint64, blockedRanges []config.RegisterRange) bool {
	if start > end {
		start, end = end, start
	}
	for _, blockedRange := range blockedRanges {
		startFunction, err := blockedRange.Start.GetModFunction()
		if err != nil {
			continue
		}
		endFunction, err := blockedRange.End.GetModFunction()
		if err != nil {
			continue
		}
		if startFunction != endFunction || startFunction != modFunction {
			continue
		}
		startAddress, _ := blockedRange.Start.GetModAddress()
		endAddress, _ := blockedRange.End.GetModAddress()
		if startAddress > endAddress {
			startAddress, endAddress = endAddress, startAddress
		}
		if start <= endAddress && end >= startAddress {
			return true
		}
	}
	return false
}

func scrapeMetricRange(r Range) ([]metric, error) {
	var metrics []metric

	for _, definitions := range r.metricDefinitions {
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

		modBytes, err := r.ReadFunction(uint16(firstAddress), totalOffset)
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
