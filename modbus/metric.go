package modbus

import (
	"github.com/RichiH/modbus_exporter/config"
)

type metric struct {
	Name       string
	Help       string
	Labels     map[string]string
	Value      float64
	MetricType config.MetricType
}
