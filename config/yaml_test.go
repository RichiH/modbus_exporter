package config

import (
	"fmt"
	"testing"
)

func TestMetricDefValidate(t *testing.T) {
	one := 1
	for _, test := range []struct {
		name        string
		metricDef   MetricDef
		expectedErr error
	}{
		{
			"bool",
			MetricDef{
				DataType: ModbusBool,
			},
			nil,
		},
		{
			"bool",
			MetricDef{
				DataType:  ModbusInt16,
				BitOffset: &one,
			},
			fmt.Errorf("bitPosition can only be used with boolean data type"),
		},
	} {
		err := test.metricDef.validate()

		if err != test.expectedErr {
			if err == nil || test.expectedErr == nil {
				t.Fatalf("expected err to be %v but got %v", test.expectedErr, err)
			}

			if err.Error() != test.expectedErr.Error() {
				t.Fatalf("expected err to be %v but got %v", test.expectedErr, err)
			}
		}
	}
}

func TestMetricDefParse(t *testing.T) {
	offsetZero := 0
	offsetOne := 1

	tests := []struct {
		name          string
		input         func() [2]byte
		metricDef     func() *MetricDef
		expectedValue float64
	}{
		{
			name: "bool, no bit",
			input: func() [2]byte {
				input := [2]byte{}
				input[0] = uint8(0)
				input[1] = uint8(0)

				return input
			},
			metricDef: func() *MetricDef {
				return &MetricDef{
					DataType:  ModbusBool,
					BitOffset: &offsetZero,
				}
			},
			expectedValue: 0,
		},
		{
			name: "bool, first bit",
			input: func() [2]byte {
				input := [2]byte{}
				input[0] = uint8(0)
				input[1] = uint8(1)

				return input
			},
			metricDef: func() *MetricDef {
				return &MetricDef{
					DataType:  ModbusBool,
					BitOffset: &offsetZero,
				}
			},
			expectedValue: 1,
		},
		{
			name: "bool, second bit",
			input: func() [2]byte {
				input := [2]byte{}
				input[0] = uint8(0)
				input[1] = uint8(2)

				return input
			},
			metricDef: func() *MetricDef {
				return &MetricDef{
					DataType:  ModbusBool,
					BitOffset: &offsetOne,
				}
			},
			expectedValue: 1,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			f, err := test.metricDef().Parse(test.input())
			if err != nil {
				t.Fatal(err)
			}

			if f != test.expectedValue {
				t.Fatalf("expected metric value to be %v but got %v", test.expectedValue, f)
			}
		})
	}
}
