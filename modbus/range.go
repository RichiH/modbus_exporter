package modbus

import "github.com/RichiH/modbus_exporter/config"

// Range defines a Modbus range that includes a Modbus function and associated metric definitions.
// metric definitions is a slice of continuous or semi-continuous(based on sensitivity) definition interval.
type Range struct {
	F           modbusFunc
	definitions [][]config.MetricDef
}

// RangeMap represents a mapping of Modbus function codes to corresponding Range objects.
type RangeMap map[uint64]Range
