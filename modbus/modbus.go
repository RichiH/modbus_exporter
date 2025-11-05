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

package modbus

import (
	"encoding/binary"
	"fmt"
	"math"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/RichiH/modbus_exporter/config"
	"github.com/goburrow/modbus"
)

// Exporter represents a Prometheus exporter converting modbus information
// retrieved from remote targets via TCP as Prometheus style metrics.
type Exporter struct {
	Config config.Config
}

// NewExporter returns a new modbus exporter.
func NewExporter(config config.Config) *Exporter {
	return &Exporter{config}
}

// GetConfig loads the config file
func (e *Exporter) GetConfig() *config.Config {
	return &e.Config
}

// Scrape scrapes the given target via TCP based on the configuration of the
// specified module returning a Prometheus gatherer with the resulting metrics.
func (e *Exporter) Scrape(targetAddress string, subTarget byte, moduleName string) (prometheus.Gatherer, error) {
	reg := prometheus.NewRegistry()

	module := e.Config.GetModule(moduleName)
	if module == nil {
		return nil, fmt.Errorf("failed to find '%v' in config", moduleName)
	}

	// TODO: We should probably be reusing these, right?
	handler := modbus.NewTCPClientHandler(targetAddress)
	if module.Timeout != 0 {
		handler.Timeout = time.Duration(module.Timeout) * time.Millisecond
	}
	handler.SlaveId = subTarget
	if err := handler.Connect(); err != nil {
		return nil, fmt.Errorf("unable to connect with target %s via module %s",
			targetAddress, module.Name)
	}

	if module.Workarounds.SleepAfterConnect > 0 {
		time.Sleep(module.Workarounds.SleepAfterConnect)
	}

	// TODO: Should we reuse this?
	c := modbus.NewClient(handler)

	// Close tcp connection.
	defer handler.Close()

	metrics, err := scrapeMetrics(module.Metrics, c)
	if err != nil {
		return nil, fmt.Errorf("failed to scrape metrics for module '%v': %v", moduleName, err.Error())
	}

	if err := registerMetrics(reg, moduleName, metrics); err != nil {
		return nil, fmt.Errorf("failed to register metrics for module %v: %v", moduleName, err.Error())
	}

	return reg, nil
}

func registerMetrics(reg prometheus.Registerer, moduleName string, metrics []metric) error {
	registeredGauges := map[string]*prometheus.GaugeVec{}
	registeredCounters := map[string]*prometheus.CounterVec{}

	for _, m := range metrics {
		if m.Labels == nil {
			m.Labels = map[string]string{}
		}
		m.Labels["module"] = moduleName

		switch m.MetricType {
		case config.MetricTypeGauge:
			// Make sure not to register the same metric twice.
			collector, ok := registeredGauges[m.Name]

			if !ok {
				collector = prometheus.NewGaugeVec(prometheus.GaugeOpts{
					Name: m.Name,
					Help: m.Help,
				}, keys(m.Labels))

				if err := reg.Register(collector); err != nil {
					return fmt.Errorf("failed to register metric %v: %v", m.Name, err.Error())
				}

				registeredGauges[m.Name] = collector
			}

			collector.With(m.Labels).Set(m.Value)
		case config.MetricTypeCounter:
			// Make sure not to register the same metric twice.
			collector, ok := registeredCounters[m.Name]

			if !ok {
				collector = prometheus.NewCounterVec(prometheus.CounterOpts{
					Name: m.Name,
					Help: m.Help,
				}, keys(m.Labels))

				if err := reg.Register(collector); err != nil {
					return fmt.Errorf("failed to register metric %v: %v", m.Name, err.Error())
				}

				registeredCounters[m.Name] = collector
			}

			// Prometheus client library panics among other things
			// if the counter value is negative. The below construct
			// recovers from such panic and properly returns the error
			// with meta data attached.
			var err error

			func() {
				defer func() {
					if r := recover(); r != nil {
						err = r.(error)
					}
				}()

				collector.With(m.Labels).Add(m.Value)
			}()

			if err != nil {
				return fmt.Errorf(
					"metric '%v', type '%v', value '%v', labels '%v': %v",
					m.Name, m.MetricType, m.Value, m.Labels, err,
				)
			}
		}

	}

	return nil
}

func keys(m map[string]string) []string {
	keys := []string{}
	for k := range m {
		keys = append(keys, k)
	}

	return keys
}

func scrapeMetrics(definitions []config.MetricDef, c modbus.Client) ([]metric, error) {
	metrics := []metric{}

	if len(definitions) == 0 {
		return []metric{}, nil
	}

	for _, definition := range definitions {
		var f modbusFunc

		// Here we are parcing Modbus Address from config file
		// for function code and register address
		modFunction, err := strconv.ParseUint(fmt.Sprint(definition.Address)[0:1], 10, 64)
		if err != nil {
			return []metric{}, fmt.Errorf("modbus function code parcing failed: %v", modFunction)
		}

		// And here we are parcing Modbus Address from config file
		// for register address
		modAddress, err := strconv.ParseUint(fmt.Sprint(definition.Address)[1:], 10, 64)
		if err != nil {
			return []metric{}, fmt.Errorf("modbus register address parcing failed  %v", modAddress)
		}

		if modAddress > 65535 {
			return []metric{}, fmt.Errorf("modbus register address is out of range: %v", definition.Address)
		}

		switch modFunction {
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
				"metric: '%v', address '%v': metric address should be within the range of 10 - 465535."+
					"'1xxxxx' for read coil / digital output, '2xxxxx' for read discrete inputs / digital input,"+
					"'3xxxxx' read holding registers / analog output, '4xxxxx' read input registers / analog input",
				definition.Name, definition.Address,
			)
		}

		m, err := scrapeMetric(definition, f, modAddress)
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
func scrapeMetric(definition config.MetricDef, f modbusFunc, modAddress uint64) (metric, error) {
	// For now we are not caching any results, thus we can request the
	// minimum necessary amount of registers per request dependint in the dataType.
	// For future reference, the maximum for digital in/output is 2000 registers,
	// the maximum for analog in/output is 125.
	var div uint16
	switch definition.DataType {
	case config.ModbusFloat16,
		config.ModbusInt16,
		config.ModbusBool,
		config.ModbusUInt16:
		div = uint16(1)
	case config.ModbusFloat32,
		config.ModbusInt32,
		config.ModbusUInt32:
		div = uint16(2)
	default:
		div = uint16(4)
	}

	// TODO: We could cache the results to not repeat overlapping ones.

	modBytes, err := f(uint16(modAddress), div)
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
	return fmt.Sprintf("insufficient amount of register data provided: %v", e.e)
}

// Parse parses the given byte slice based on the specified Modbus data type and
// returns the parsed value as a float64 (Prometheus exposition format).
//
// TODO: Handle Endianness.
func parseModbusData(d config.MetricDef, rawData []byte) (float64, error) {
	switch d.DataType {
	case config.ModbusBool:
		{
			if d.BitOffset == nil {
				return float64(0), fmt.Errorf("expected bit position on boolean data type")
			}

			// Convert byte to uint16
			data := binary.BigEndian.Uint16(rawData)

			if data&(uint16(1)<<uint16(*d.BitOffset)) > 0 {
				return float64(1), nil
			}
			return float64(0), nil
		}
	case config.ModbusFloat16:
		{
			if len(rawData) != 2 {
				return float64(0), &InsufficientRegistersError{fmt.Sprintf("expected 2 bytes, got %v", len(rawData))}
			}
			panic("implement")
		}
	case config.ModbusInt16:
		{
			if len(rawData) != 2 {
				return float64(0), &InsufficientRegistersError{fmt.Sprintf("expected 2 bytes, got %v", len(rawData))}
			}
			rawDataWithEndianness, err := convertEndianness16b(d.Endianness, rawData)
			if err != nil {
				return float64(0), err
			}
			data := binary.BigEndian.Uint16(rawDataWithEndianness)
			return scaleValue(d.Factor, float64(int16(data))), nil
		}
	case config.ModbusUInt16:
		{
			if len(rawData) != 2 {
				return float64(0), &InsufficientRegistersError{fmt.Sprintf("expected 2 bytes, got %v", len(rawData))}
			}
			rawDataWithEndianness, err := convertEndianness16b(d.Endianness, rawData)
			if err != nil {
				return float64(0), err
			}
			data := binary.BigEndian.Uint16(rawDataWithEndianness)
			return scaleValue(d.Factor, float64(data)), nil
		}
	case config.ModbusInt32:
		{
			if len(rawData) != 4 {
				return float64(0), &InsufficientRegistersError{fmt.Sprintf("expected 4 bytes, got %v", len(rawData))}
			}
			rawDataWithEndianness, err := convertEndianness32b(d.Endianness, rawData)
			if err != nil {
				return float64(0), err
			}
			data := binary.BigEndian.Uint32(rawDataWithEndianness)
			return scaleValue(d.Factor, float64(int32(data))), nil
		}
	case config.ModbusUInt32:
		{
			if len(rawData) != 4 {
				return float64(0), &InsufficientRegistersError{fmt.Sprintf("expected 4 bytes, got %v", len(rawData))}
			}
			rawDataWithEndianness, err := convertEndianness32b(d.Endianness, rawData)
			if err != nil {
				return float64(0), err
			}
			data := binary.BigEndian.Uint32(rawDataWithEndianness)
			return scaleValue(d.Factor, float64(data)), nil
		}
	case config.ModbusFloat32:
		{
			if len(rawData) != 4 {
				return float64(0), &InsufficientRegistersError{fmt.Sprintf("expected 4 bytes, got %v", len(rawData))}
			}
			rawDataWithEndianness, err := convertEndianness32b(d.Endianness, rawData)
			if err != nil {
				return float64(0), err
			}
			data := binary.BigEndian.Uint32(rawDataWithEndianness)
			return scaleValue(d.Factor, float64(math.Float32frombits(data))), nil
		}
	case config.ModbusInt64:
		{
			if len(rawData) != 8 {
				return float64(0), &InsufficientRegistersError{fmt.Sprintf("expected 8 bytes, got %v", len(rawData))}
			}
			rawDataWithEndianness, err := convertEndianness64b(d.Endianness, rawData)
			if err != nil {
				return float64(0), err
			}
			data := binary.BigEndian.Uint64(rawDataWithEndianness)
			return scaleValue(d.Factor, float64(int64(data))), nil
		}
	case config.ModbusUInt64:
		{
			if len(rawData) != 8 {
				return float64(0), &InsufficientRegistersError{fmt.Sprintf("expected 8 bytes, got %v", len(rawData))}
			}
			rawDataWithEndianness, err := convertEndianness64b(d.Endianness, rawData)
			if err != nil {
				return float64(0), err
			}
			data := binary.BigEndian.Uint64(rawDataWithEndianness)
			return scaleValue(d.Factor, float64(data)), nil
		}
	case config.ModbusFloat64:
		{
			if len(rawData) != 8 {
				return float64(0), &InsufficientRegistersError{fmt.Sprintf("expected 8 bytes, got %v", len(rawData))}
			}
			rawDataWithEndianness, err := convertEndianness64b(d.Endianness, rawData)
			if err != nil {
				return float64(0), err
			}
			data := binary.BigEndian.Uint64(rawDataWithEndianness)
			return scaleValue(d.Factor, math.Float64frombits(data)), nil
		}
	default:
		{
			return 0, fmt.Errorf("unknown modbus data type")
		}
	}
}

// Scales value by factor
func scaleValue(f *float64, d float64) float64 {
	if f == nil {
		return d
	}
	return (d * float64(*f))
}

// Converts an array of 16 bits from an endianness to the default big Endian
func convertEndianness16b(rawEndianness config.EndiannessType, rawData []byte) ([]byte, error) {
	if len(rawData) != 2 {
		return []byte{uint8(0), uint8(0)}, fmt.Errorf("expected 2 bytes, got %v", len(rawData))
	}
	var data []byte
	switch rawEndianness {
	case config.EndiannessLittleEndian:
		data = []byte{
			rawData[1],
			rawData[0]}
	// default: BigEndian
	default:
		data = []byte{
			rawData[0],
			rawData[1]}
	}
	return data, nil
}

// Converts an array of 32 bits from an endianness to the default big Endian
func convertEndianness32b(rawEndianness config.EndiannessType, rawData []byte) ([]byte, error) {
	if len(rawData) != 4 {
		return []byte{uint8(0), uint8(0), uint8(0), uint8(0)},
			fmt.Errorf("expected 4 bytes, got %v", len(rawData))
	}
	var data []byte
	switch rawEndianness {
	case config.EndiannessLittleEndian:
		data = []byte{
			rawData[3],
			rawData[2],
			rawData[1],
			rawData[0]}
	case config.EndiannessMixedEndian:
		data = []byte{
			rawData[1],
			rawData[0],
			rawData[3],
			rawData[2]}
	case config.EndiannessYolo:
		data = []byte{
			rawData[2],
			rawData[3],
			rawData[0],
			rawData[1]}
	// default: BigEndian
	default:
		data = []byte{
			rawData[0],
			rawData[1],
			rawData[2],
			rawData[3]}
	}
	return data, nil
}

// Converts an array of 64 bits from an endianness to the default big Endian
func convertEndianness64b(rawEndianness config.EndiannessType, rawData []byte) ([]byte, error) {
	if len(rawData) != 8 {
		return []byte{uint8(0), uint8(0), uint8(0), uint8(0), uint8(0), uint8(0), uint8(0), uint8(0)},
			fmt.Errorf("expected 8 bytes, got %v", len(rawData))
	}
	var data []byte
	switch rawEndianness {
	case config.EndiannessLittleEndian:
		data = []byte{
			rawData[7],
			rawData[6],
			rawData[5],
			rawData[4],
			rawData[3],
			rawData[2],
			rawData[1],
			rawData[0]}
	case config.EndiannessMixedEndian:
		data = []byte{
			rawData[1],
			rawData[0],
			rawData[3],
			rawData[2],
			rawData[5],
			rawData[4],
			rawData[7],
			rawData[6]}
	case config.EndiannessYolo:
		data = []byte{
			rawData[6],
			rawData[7],
			rawData[4],
			rawData[5],
			rawData[2],
			rawData[3],
			rawData[0],
			rawData[1]}
	// default: BigEndian
	default:
		data = []byte{
			rawData[0],
			rawData[1],
			rawData[2],
			rawData[3],
			rawData[4],
			rawData[5],
			rawData[6],
			rawData[7]}
	}
	return data, nil
}
