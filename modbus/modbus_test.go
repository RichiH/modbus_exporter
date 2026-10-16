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
	"errors"
	"maps"
	"math"
	"testing"

	"github.com/RichiH/modbus_exporter/config"
	"github.com/prometheus/client_golang/prometheus"
)

func TestRegisterMetrics(t *testing.T) {
	t.Run("does not fail", func(t *testing.T) {
		reg := prometheus.NewRegistry()
		moduleName := "my_module"
		metrics := []metric{}

		if err := registerMetrics(reg, moduleName, metrics); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("registers metrics with same name and same label keys", func(t *testing.T) {
		reg := prometheus.NewRegistry()
		moduleName := "my_module"
		metrics := []metric{
			{
				Name: "my_metric",
				Help: "my_help",
				Labels: map[string]string{
					"labelKey1": "labelValueA",
					"labelKey2": "labelValueA",
				},
				Value:      1,
				MetricType: config.MetricTypeCounter,
			},
			{
				Name: "my_metric",
				Help: "my_help",
				Labels: map[string]string{
					"labelKey1": "labelValueB",
					"labelKey2": "labelValueB",
				},
				Value:      2,
				MetricType: config.MetricTypeCounter,
			},
		}

		if err := registerMetrics(reg, moduleName, metrics); err != nil {
			t.Fatal(err)
		}

		metricFamilies, err := reg.Gather()
		if err != nil {
			t.Fatal(err)
		}

		if len(metricFamilies) != 1 {
			t.Fatalf("expected %v but got %v", 1, len(metricFamilies))
		}

		for _, m := range metricFamilies {
			metrics := m.Metric
			if len(metrics) != 2 {
				t.Fatalf("expected %v metrics but got %v", 2, len(metrics))
			}
		}
	})
}

func TestParseModbusData(t *testing.T) {
	offsetZero := 0
	offsetOne := 1

	tests := []struct {
		name          string
		input         func() []byte
		metricDef     func() *config.MetricDef
		expectedValue float64
	}{
		{
			name: "bool, no bit",
			input: func() []byte {
				return []byte{uint8(0)}
			},
			metricDef: func() *config.MetricDef {
				return &config.MetricDef{
					DataType:  config.ModbusBool,
					BitOffset: &offsetZero,
				}
			},
			expectedValue: 0,
		},
		{
			name: "bool, first bit",
			input: func() []byte {
				return []byte{uint8(1)}
			},
			metricDef: func() *config.MetricDef {
				return &config.MetricDef{
					DataType:  config.ModbusBool,
					BitOffset: &offsetZero,
				}
			},
			expectedValue: 1,
		},
		{
			name: "bool, second bit",
			input: func() []byte {
				return []byte{uint8(2)}
			},
			metricDef: func() *config.MetricDef {
				return &config.MetricDef{
					DataType:  config.ModbusBool,
					BitOffset: &offsetOne,
				}
			},
			expectedValue: 1,
		},
		{
			name: "bool, register, bit nine",
			input: func() []byte {
				return []byte{uint8(2), uint8(0)}
			},
			metricDef: func() *config.MetricDef {
				offsetNine := 9
				return &config.MetricDef{
					DataType:  config.ModbusBool,
					BitOffset: &offsetNine,
				}
			},
			expectedValue: 1,
		},
		{
			name: "bool, register, bit zero unset",
			input: func() []byte {
				return []byte{uint8(1), uint8(0)}
			},
			metricDef: func() *config.MetricDef {
				return &config.MetricDef{
					DataType:  config.ModbusBool,
					BitOffset: &offsetZero,
				}
			},
			expectedValue: 0,
		},
		{
			name: "int16, default endian (big endian)",
			input: func() []byte {
				return []byte{uint8(0), uint8(1)}
			},
			metricDef: func() *config.MetricDef {
				return &config.MetricDef{
					DataType: config.ModbusInt16,
				}
			},
			expectedValue: 1,
		},
		{
			name: "int16, Big endian",
			input: func() []byte {
				return []byte{uint8(0), uint8(1)}
			},
			metricDef: func() *config.MetricDef {
				return &config.MetricDef{
					DataType:   config.ModbusInt16,
					Endianness: config.EndiannessBigEndian,
				}
			},
			expectedValue: 1,
		},
		{
			name: "int16, Little endian",
			input: func() []byte {
				return []byte{uint8(1), uint8(0)}
			},
			metricDef: func() *config.MetricDef {
				return &config.MetricDef{
					DataType:   config.ModbusInt16,
					Endianness: config.EndiannessLittleEndian,
				}
			},
			expectedValue: 1,
		},
		{
			name: "uint16, default endian (big endian)",
			input: func() []byte {
				return []byte{uint8(0), uint8(1)}
			},
			metricDef: func() *config.MetricDef {
				return &config.MetricDef{
					DataType: config.ModbusUInt16,
				}
			},
			expectedValue: 1,
		},
		{
			name: "uint16, Big endian",
			input: func() []byte {
				return []byte{uint8(0), uint8(1)}
			},
			metricDef: func() *config.MetricDef {
				return &config.MetricDef{
					DataType:   config.ModbusUInt16,
					Endianness: config.EndiannessBigEndian,
				}
			},
			expectedValue: 1,
		},
		{
			name: "uint16, Little endian",
			input: func() []byte {
				return []byte{uint8(1), uint8(0)}
			},
			metricDef: func() *config.MetricDef {
				return &config.MetricDef{
					DataType:   config.ModbusUInt16,
					Endianness: config.EndiannessLittleEndian,
				}
			},
			expectedValue: 1,
		},
		{
			name: "int32, default endian (big endian)",
			input: func() []byte {
				return []byte{uint8(0), uint8(0), uint8(0), uint8(1)}
			},
			metricDef: func() *config.MetricDef {
				return &config.MetricDef{
					DataType: config.ModbusInt32,
				}
			},
			expectedValue: 1,
		},
		{
			name: "int32, Big endian",
			input: func() []byte {
				return []byte{uint8(0), uint8(0), uint8(0), uint8(1)}
			},
			metricDef: func() *config.MetricDef {
				return &config.MetricDef{
					DataType:   config.ModbusInt32,
					Endianness: config.EndiannessBigEndian,
				}
			},
			expectedValue: 1,
		},
		{
			name: "int32, Little endian",
			input: func() []byte {
				return []byte{uint8(1), uint8(0), uint8(0), uint8(0)}
			},
			metricDef: func() *config.MetricDef {
				return &config.MetricDef{
					DataType:   config.ModbusInt32,
					Endianness: config.EndiannessLittleEndian,
				}
			},
			expectedValue: 1,
		},
		{
			name: "int32, Mixed endian",
			input: func() []byte {
				return []byte{uint8(0), uint8(0), uint8(1), uint8(0)}
			},
			metricDef: func() *config.MetricDef {
				return &config.MetricDef{
					DataType:   config.ModbusInt32,
					Endianness: config.EndiannessMixedEndian,
				}
			},
			expectedValue: 1,
		},
		{
			name: "int32, Yolo endian",
			input: func() []byte {
				return []byte{uint8(0), uint8(1), uint8(0), uint8(0)}
			},
			metricDef: func() *config.MetricDef {
				return &config.MetricDef{
					DataType:   config.ModbusInt32,
					Endianness: config.EndiannessYolo,
				}
			},
			expectedValue: 1,
		},
		{
			name: "uint32, default endian (big endian)",
			input: func() []byte {
				return []byte{uint8(0), uint8(0), uint8(0), uint8(1)}
			},
			metricDef: func() *config.MetricDef {
				return &config.MetricDef{
					DataType: config.ModbusUInt32,
				}
			},
			expectedValue: 1,
		},
		{
			name: "uint32, Big endian",
			input: func() []byte {
				return []byte{uint8(0), uint8(0), uint8(0), uint8(1)}
			},
			metricDef: func() *config.MetricDef {
				return &config.MetricDef{
					DataType:   config.ModbusUInt32,
					Endianness: config.EndiannessBigEndian,
				}
			},
			expectedValue: 1,
		},
		{
			name: "uint32, Little endian",
			input: func() []byte {
				return []byte{uint8(1), uint8(0), uint8(0), uint8(0)}
			},
			metricDef: func() *config.MetricDef {
				return &config.MetricDef{
					DataType:   config.ModbusUInt32,
					Endianness: config.EndiannessLittleEndian,
				}
			},
			expectedValue: 1,
		},
		{
			name: "uint32, Mixed endian",
			input: func() []byte {
				return []byte{uint8(0), uint8(0), uint8(1), uint8(0)}
			},
			metricDef: func() *config.MetricDef {
				return &config.MetricDef{
					DataType:   config.ModbusUInt32,
					Endianness: config.EndiannessMixedEndian,
				}
			},
			expectedValue: 1,
		},
		{
			name: "uint32, Yolo endian",
			input: func() []byte {
				return []byte{uint8(0), uint8(1), uint8(0), uint8(0)}
			},
			metricDef: func() *config.MetricDef {
				return &config.MetricDef{
					DataType:   config.ModbusUInt32,
					Endianness: config.EndiannessYolo,
				}
			},
			expectedValue: 1,
		},
		{
			name: "int64, default endian (big endian)",
			input: func() []byte {
				return []byte{uint8(0), uint8(0), uint8(0), uint8(0), uint8(0), uint8(0), uint8(0), uint8(1)}
			},
			metricDef: func() *config.MetricDef {
				return &config.MetricDef{
					DataType: config.ModbusInt64,
				}
			},
			expectedValue: 1,
		},
		{
			name: "int64, Big endian",
			input: func() []byte {
				return []byte{uint8(0), uint8(0), uint8(0), uint8(0), uint8(0), uint8(0), uint8(0), uint8(1)}
			},
			metricDef: func() *config.MetricDef {
				return &config.MetricDef{
					DataType:   config.ModbusInt64,
					Endianness: config.EndiannessBigEndian,
				}
			},
			expectedValue: 1,
		},
		{
			name: "int64, Little endian",
			input: func() []byte {
				return []byte{uint8(1), uint8(0), uint8(0), uint8(0), uint8(0), uint8(0), uint8(0), uint8(0)}
			},
			metricDef: func() *config.MetricDef {
				return &config.MetricDef{
					DataType:   config.ModbusInt64,
					Endianness: config.EndiannessLittleEndian,
				}
			},
			expectedValue: 1,
		},
		{
			name: "int64, Mixed endian",
			input: func() []byte {
				return []byte{uint8(0), uint8(0), uint8(0), uint8(0), uint8(0), uint8(0), uint8(1), uint8(0)}
			},
			metricDef: func() *config.MetricDef {
				return &config.MetricDef{
					DataType:   config.ModbusInt64,
					Endianness: config.EndiannessMixedEndian,
				}
			},
			expectedValue: 1,
		},
		{
			name: "int64, Yolo endian",
			input: func() []byte {
				return []byte{uint8(0), uint8(1), uint8(0), uint8(0), uint8(0), uint8(0), uint8(0), uint8(0)}
			},
			metricDef: func() *config.MetricDef {
				return &config.MetricDef{
					DataType:   config.ModbusInt64,
					Endianness: config.EndiannessYolo,
				}
			},
			expectedValue: 1,
		},
		{
			name: "uint64, default endian (big endian)",
			input: func() []byte {
				return []byte{uint8(0), uint8(0), uint8(0), uint8(0), uint8(0), uint8(0), uint8(0), uint8(1)}
			},
			metricDef: func() *config.MetricDef {
				return &config.MetricDef{
					DataType: config.ModbusUInt64,
				}
			},
			expectedValue: 1,
		},
		{
			name: "uint64, Big endian",
			input: func() []byte {
				return []byte{uint8(0), uint8(0), uint8(0), uint8(0), uint8(0), uint8(0), uint8(0), uint8(1)}
			},
			metricDef: func() *config.MetricDef {
				return &config.MetricDef{
					DataType:   config.ModbusUInt64,
					Endianness: config.EndiannessBigEndian,
				}
			},
			expectedValue: 1,
		},
		{
			name: "uint64, Little endian",
			input: func() []byte {
				return []byte{uint8(1), uint8(0), uint8(0), uint8(0), uint8(0), uint8(0), uint8(0), uint8(0)}
			},
			metricDef: func() *config.MetricDef {
				return &config.MetricDef{
					DataType:   config.ModbusUInt64,
					Endianness: config.EndiannessLittleEndian,
				}
			},
			expectedValue: 1,
		},
		{
			name: "uint64, Mixed endian",
			input: func() []byte {
				return []byte{uint8(0), uint8(0), uint8(0), uint8(0), uint8(0), uint8(0), uint8(1), uint8(0)}
			},
			metricDef: func() *config.MetricDef {
				return &config.MetricDef{
					DataType:   config.ModbusUInt64,
					Endianness: config.EndiannessMixedEndian,
				}
			},
			expectedValue: 1,
		},
		{
			name: "uint64, Yolo endian",
			input: func() []byte {
				return []byte{uint8(0), uint8(1), uint8(0), uint8(0), uint8(0), uint8(0), uint8(0), uint8(0)}
			},
			metricDef: func() *config.MetricDef {
				return &config.MetricDef{
					DataType:   config.ModbusUInt64,
					Endianness: config.EndiannessYolo,
				}
			},
			expectedValue: 1,
		},
	}

	for _, loopTest := range tests {

		test := loopTest

		t.Run(test.name, func(t *testing.T) {
			f, err := parseModbusData(*test.metricDef(), test.input())
			if err != nil {
				t.Fatal(err)
			}

			if f != test.expectedValue {
				t.Fatalf("expected metric value to be %v but got %v", test.expectedValue, f)
			}
		})
	}
}

func TestParseModbusDataInsufficientRegisters(t *testing.T) {
	d := config.MetricDef{
		DataType: config.ModbusInt16,
	}

	_, err := parseModbusData(d, []byte{})

	if err == nil {
		t.Fatal("expected error but got nil")
	}

	switch err.(type) {
	case *InsufficientRegistersError:
	default:
		t.Fatal("expected InsufficientRegistersError")
	}
}

func TestParseModbusDataBoolInvalidInput(t *testing.T) {
	offsetZero := 0
	offsetEight := 8
	offsetSixteen := 16

	tests := []struct {
		name          string
		input         []byte
		bitOffset     *int
		insufficientR bool
	}{
		{name: "empty input", input: []byte{}, bitOffset: &offsetZero, insufficientR: true},
		{name: "three bytes", input: []byte{0, 0, 0}, bitOffset: &offsetZero, insufficientR: true},
		{name: "single byte, offset out of range", input: []byte{0xff}, bitOffset: &offsetEight},
		{name: "register, offset out of range", input: []byte{0xff, 0xff}, bitOffset: &offsetSixteen},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			d := config.MetricDef{
				DataType:  config.ModbusBool,
				BitOffset: test.bitOffset,
			}

			_, err := parseModbusData(d, test.input)
			if err == nil {
				t.Fatal("expected error but got nil")
			}

			var insufficient *InsufficientRegistersError
			if test.insufficientR && !errors.As(err, &insufficient) {
				t.Fatalf("expected InsufficientRegistersError but got %T", err)
			}
		})
	}
}

func TestParseModbusDataFloat32(t *testing.T) {
	data := make([]byte, 4)
	binary.BigEndian.PutUint32(data, math.Float32bits(32))

	def := config.MetricDef{
		DataType: config.ModbusFloat32,
	}

	floatValue, err := parseModbusData(def, data)
	if err != nil {
		t.Fatal(err)
	}

	if floatValue != 32 {
		t.Fatalf("expected 32 but got %v", floatValue)
	}
}

// TestRegisterMetricTwoMetricsSameName makes sure registerMetrics reuses a
// registered metric in case there is a second one with the same name instead of
// reregistering which would cause an exception.
func TestRegisterMetricTwoMetricsSameName(t *testing.T) {
	reg := prometheus.NewRegistry()
	a := metric{"my_metric", "", map[string]string{}, 1, config.MetricTypeCounter}
	b := metric{"my_metric", "", map[string]string{}, 1, config.MetricTypeCounter}

	err := registerMetrics(reg, "my_module", []metric{a, b})
	if err != nil {
		t.Fatalf("expected no error but got: %v", err)
	}
}

// TestRegisterMetricsRecoverNegativeCounter makes sure the function properly
// recovers from a prometheus client library panic on negative counter changes.
func TestRegisterMetricsRecoverNegativeCounter(t *testing.T) {
	reg := prometheus.NewRegistry()
	a := metric{"my_metric", "", map[string]string{"key1": "value1", "key2": "value2"}, -1, config.MetricTypeCounter}

	err := registerMetrics(reg, "my_module", []metric{a})
	if err == nil {
		t.Fatal("expected an error but got nil")
	}
}

func TestScrapeMetricSpecialValues(t *testing.T) {
	special := 65535.0
	scaled := 5.0
	factor := 0.5

	tests := []struct {
		name     string
		raw      []byte
		def      config.MetricDef
		expected metric
	}{
		{
			name: "raw value matches, factor ignored for comparison",
			raw:  []byte{0xff, 0xff},
			def: config.MetricDef{
				Name:          "watt_hours",
				Labels:        map[string]string{"phase": "1"},
				Address:       301220,
				DataType:      config.ModbusUInt16,
				MetricType:    config.MetricTypeCounter,
				Factor:        &factor,
				SpecialValues: []config.SpecialValueDef{{Value: &special, Meaning: "data pending"}},
			},
			expected: metric{
				Name:       config.SpecialValueMetricName,
				Help:       specialValueHelp,
				Labels:     map[string]string{"reading": "watt_hours", "address": "301220", "meaning": "data pending"},
				Value:      65535,
				MetricType: config.MetricTypeGauge,
			},
		},
		{
			name: "scaled value matching a special value is not caught",
			raw:  []byte{0x00, 0x0a},
			def: config.MetricDef{
				Name:          "watt_hours",
				Address:       301220,
				DataType:      config.ModbusUInt16,
				MetricType:    config.MetricTypeCounter,
				Factor:        &factor,
				SpecialValues: []config.SpecialValueDef{{Value: &scaled, Meaning: "data pending"}},
			},
			expected: metric{
				Name:       "watt_hours",
				Value:      5,
				MetricType: config.MetricTypeCounter,
			},
		},
		{
			name: "no special values configured",
			raw:  []byte{0xff, 0xff},
			def: config.MetricDef{
				Name:       "watt_hours",
				Address:    301220,
				DataType:   config.ModbusUInt16,
				MetricType: config.MetricTypeCounter,
			},
			expected: metric{
				Name:       "watt_hours",
				Value:      65535,
				MetricType: config.MetricTypeCounter,
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			f := func(address, quantity uint16) ([]byte, error) {
				return test.raw, nil
			}

			m, err := scrapeMetric(test.def, f, 1220)
			if err != nil {
				t.Fatal(err)
			}

			if m.Name != test.expected.Name || m.Help != test.expected.Help ||
				m.Value != test.expected.Value || m.MetricType != test.expected.MetricType {
				t.Fatalf("expected %+v but got %+v", test.expected, m)
			}

			if test.expected.Labels != nil && !maps.Equal(m.Labels, test.expected.Labels) {
				t.Fatalf("expected labels %v but got %v", test.expected.Labels, m.Labels)
			}
		})
	}
}

func TestRegisterMetricsSpecialValueCaught(t *testing.T) {
	reg := prometheus.NewRegistry()

	metrics := []metric{
		{
			Name:       "power",
			Labels:     map[string]string{"phase": "1"},
			Value:      230,
			MetricType: config.MetricTypeGauge,
		},
		newSpecialValueMetric(config.MetricDef{Name: "watt_hours", Address: 301220}, "data pending", 4294934512),
		newSpecialValueMetric(config.MetricDef{Name: "voltage", Address: 301222}, "reserved", 65535),
	}

	if err := registerMetrics(reg, "my_module", metrics); err != nil {
		t.Fatal(err)
	}

	metricFamilies, err := reg.Gather()
	if err != nil {
		t.Fatal(err)
	}

	for _, mf := range metricFamilies {
		if mf.GetName() != config.SpecialValueMetricName {
			continue
		}

		if len(mf.Metric) != 2 {
			t.Fatalf("expected 2 caught special values but got %v", len(mf.Metric))
		}

		return
	}

	t.Fatalf("expected metric family %v", config.SpecialValueMetricName)
}
