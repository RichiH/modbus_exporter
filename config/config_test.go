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
	m.Metrics = []MetricDef{
		{
			DataType: ModbusInt16,
		},
	}

	err := m.validate()
	if err == nil {
		t.Fatal("expected validation to fail with invalid modbus protocol")
	}
}
