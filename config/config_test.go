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
				DataType:   ModbusBool,
				MetricType: MetricTypeCounter,
			},
			nil,
		},
		{
			"bool",
			MetricDef{
				DataType:   ModbusInt16,
				BitOffset:  &one,
				MetricType: MetricTypeCounter,
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

func TestModuleValidate(t *testing.T) {
	m := Module{}

	m.Protocol = "invalid"
	m.AnalogOutput = []MetricDef{
		{
			DataType: ModbusInt16,
		},
	}

	err := m.validate()
	if err == nil {
		t.Fatal("expected validation to fail with invalid modbus protocol")
	}
}
