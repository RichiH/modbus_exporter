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

// Package modbus contains all the modbus related components
package modbus

import (
	"encoding/binary"
	"fmt"
	"math"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/log"

	"github.com/goburrow/modbus"
	"github.com/lupoDharkael/modbus_exporter/config"
)

// Collector represents a Prometheus exporter converting modbus information
// retrieved from remote targets via TCP as Prometheus style metrics.
type Collector struct {
	target    string
	subTarget byte
	module    config.Module
}

// NewExporter returns a new modbus exporter.
func NewCollector(target string, subTarget byte, module config.Module) *Collector {
	return &Collector{target, subTarget, module}
}

// Describe implements Prometheus.Collector.
func (c Collector) Describe(ch chan<- *prometheus.Desc) {
	ch <- prometheus.NewDesc("dummy", "dummy", nil, nil)
}

// Collect scrapes the given target via TCP based on the configuration of the
// specified module returning a Prometheus gatherer with the resulting metrics.
func (c Collector) Collect(ch chan<- prometheus.Metric) {

	// TODO: We should probably be reusing these, right?
	handler := modbus.NewTCPClientHandler(c.target)
	if c.module.Timeout != 0 {
		handler.Timeout = time.Duration(c.module.Timeout) * time.Millisecond
	}
	handler.SlaveId = c.subTarget
	if err := handler.Connect(); err != nil {
		err = fmt.Errorf("unable to connect with target %s via module %s",
			c.target, c.module.Name)
		log.Error(err)
		ch <- prometheus.NewInvalidMetric(prometheus.NewDesc("modbus_error", err.Error(), nil, nil), err)
	}

	// TODO: Should we reuse this?
	client := modbus.NewClient(handler)

	metrics, err := scrapeMetrics(c.module.Metrics, client)
	if err != nil {
		err = fmt.Errorf("failed to scrape metrics for module '%v': %v", c.module.Name, err.Error())
		log.Error(err)
		ch <- prometheus.NewInvalidMetric(prometheus.NewDesc("modbus_error", err.Error(), nil, nil), err)
	}

	if err := putMetrics(ch, c.module.Name, metrics); err != nil {
		err := fmt.Errorf("failed to register metrics for module %v: %v", c.module.Name, err.Error())
		log.Error(err)
		ch <- prometheus.NewInvalidMetric(prometheus.NewDesc("modbus_error", err.Error(), nil, nil), err)
	}
}

func keys(labels map[string]string) []string {
	keys := []string{}
	for k := range labels {
		keys = append(keys, k)
	}

	return keys
}

func values(labels map[string]string) []string {
	values := []string{}
	for _, v := range labels {
		values = append(values, v)
	}

	return values
}

func putMetrics(ch chan<- prometheus.Metric, moduleName string, metrics []metric) error {
	for _, m := range metrics {
		if m.Labels == nil {
			m.Labels = map[string]string{}
		}
		m.Labels["module"] = moduleName

		desc := prometheus.NewDesc(m.Name, m.Help, keys(m.Labels), prometheus.Labels{})

		var metricType prometheus.ValueType
		switch m.MetricType {
		case config.MetricTypeGauge:
			metricType = prometheus.GaugeValue
		case config.MetricTypeCounter:
			if m.Value < 0 {
				return fmt.Errorf("metric '%v': counter can not be negative, got '%v'", m.Name, m.Value)
			}
			metricType = prometheus.CounterValue
		}

		metric, err := prometheus.NewConstMetric(desc, metricType, m.Value, values(m.Labels)...)
		if err != nil {
			return fmt.Errorf("failed to create metric %v: %v", m.Name, err.Error())
		}

		ch <- metric
	}

	return nil
}

func scrapeMetrics(definitions []config.MetricDef, c modbus.Client) ([]metric, error) {
	metrics := []metric{}

	if len(definitions) == 0 {
		return []metric{}, nil
	}

	for _, definition := range definitions {
		var f modbusFunc

		switch definition.Address / 10000 {
		case 1:
			f = c.ReadCoils
		case 2:
			f = c.ReadDiscreteInputs
		case 3:
			f = c.ReadHoldingRegisters
		case 4:
			f = c.ReadInputRegisters
		default:
			return []metric{}, fmt.Errorf(
				"metric: '%v', address '%v': metric address should be within the range of 10000 - 40000."+
					"'1xxxx' for read coil / digital output, '2xxxx' for read discrete inputs / digital input,"+
					"'3xxxx' read holding registers / analog output, '4xxxx' read input registers / analog input",
				definition.Name, definition.Address,
			)
		}

		m, err := scrapeMetric(definition, f)
		if err != nil {
			return []metric{}, fmt.Errorf("metric '%v', address '%v': %v", definition.Name, definition.Address, err)
		}

		metrics = append(metrics, m)
	}

	return metrics, nil
}

// modbus read function type
type modbusFunc func(address, quantity uint16) ([]byte, error)

// scrapeMetric returns the list of values from a target
func scrapeMetric(definition config.MetricDef, f modbusFunc) (metric, error) {
	// For now we are not caching any results, thus we can request the
	// minimum necessary amount of registers per request. Our biggest data type
	// is float32, thereby 2 registers are enough. For future reference, the
	// maximum for digital in/output is 2000 registers, the maximum for analog
	// in/output is 125.
	div := uint16(2)

	// TODO: We could cache the results to not repeat overlapping ones.
	// Modulo 10000 as the first digit identifies the modbus function code
	// (1-4).
	modBytes, err := f(uint16(definition.Address%10000), div)
	if err != nil {
		return metric{}, err
	}

	v, err := parseModbusData(definition, modBytes)
	if err != nil {
		return metric{}, err
	}

	return metric{definition.Name, definition.Help, definition.Labels, v, definition.MetricType}, nil
}

// InsufficientRegistersError is returned in Parse() whenever not enough
// registers are provided for the given data type.
type InsufficientRegistersError struct {
	e string
}

// Error implements the Golang error interface.
func (e *InsufficientRegistersError) Error() string {
	return fmt.Sprintf("insufficient amount of registers provided: %v", e.e)
}

// Parse parses the given byte slice based on the specified Modbus data type and
// returns the parsed value as a float64 (Prometheus exposition format).
//
// TODO: Handle Endianness.
func parseModbusData(d config.MetricDef, rawData []byte) (float64, error) {
	switch d.DataType {
	case config.ModbusFloat16:
		if len(rawData) < 2 {
			return float64(0), &InsufficientRegistersError{fmt.Sprintf("expected at least 1, got %v", len(rawData))}
		}
		panic("implement")
	case config.ModbusFloat32:
		if len(rawData) < 4 {
			return float64(0), &InsufficientRegistersError{fmt.Sprintf("expected at least 2, got %v", len(rawData))}
		}
		return float64(math.Float32frombits(binary.BigEndian.Uint32(rawData[:4]))), nil
	case config.ModbusInt16:
		{
			if len(rawData) < 2 {
				return float64(0), &InsufficientRegistersError{fmt.Sprintf("expected at least 1, got %v", len(rawData))}
			}
			i := binary.BigEndian.Uint16(rawData)
			return float64(int16(i)), nil
		}
	case config.ModbusUInt16:
		{
			if len(rawData) < 2 {
				return float64(0), &InsufficientRegistersError{fmt.Sprintf("expected at least 1, got %v", len(rawData))}
			}
			i := binary.BigEndian.Uint16(rawData)
			return float64(i), nil
		}
	case config.ModbusBool:
		{
			// TODO: Maybe we don't need two registers for bool.
			if len(rawData) < 2 {
				return float64(0), &InsufficientRegistersError{fmt.Sprintf("expected at least 1, got %v", len(rawData))}
			}

			if d.BitOffset == nil {
				return float64(0), fmt.Errorf("expected bit position on boolean data type")
			}

			data := binary.BigEndian.Uint16(rawData)

			if data&(uint16(1)<<uint16(*d.BitOffset)) > 0 {
				return float64(1), nil
			}
			return float64(0), nil
		}
	default:
		return 0, fmt.Errorf("unknown modbus data type")
	}
}
