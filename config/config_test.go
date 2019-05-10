package config

import (
	"fmt"
	"strconv"
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

func TestCheckPort(t *testing.T) {
	tests := []struct {
		input         string
		protocol      ModbusProtocol
		expectedError error
	}{
		{
			"localhost:8080",
			ModbusProtocolTCPIP,
			nil,
		},
		{
			"192.168.0.23:8080",
			ModbusProtocolTCPIP,
			nil,
		},
		{
			"192.168.0.3333.043",
			"",
			&ModbusProtocolValidationError{},
		},
		{":7070", "", &ModbusProtocolValidationError{}},
		{"300.34.23.2:6767", "", &ModbusProtocolValidationError{}},
		{"/dev/ttyS4sw34", "", &ModbusProtocolValidationError{}},
		{"/dev", "", &ModbusProtocolValidationError{}},
		{"/dev/ttyUSB0", ModbusProtocolSerial, nil},
		{"/dev/ttyS0", ModbusProtocolSerial, nil},
	}
	for i, loopTest := range tests {
		test := loopTest

		t.Run(strconv.Itoa(i), func(t *testing.T) {
			protocol, err := CheckPortTarget(test.input)

			if test.expectedError == nil {
				if test.protocol != protocol {
					t.Fatalf("expected protocol %v but got %v", test.protocol, protocol)
				}

				if err != nil {
					t.Fatalf("expected no error but got %v", err)
				}

				return
			}
		})
	}
}

func TestValidate(t *testing.T) {
	var (
		targetsBad = [...]Module{
			{Parity: "abc"},
			{Parity: "N"},
			{Stopbits: 4},
			{Baudrate: -1},
			{Databits: 50},
			{Baudrate: -1},
		}
		regDefTest = []MetricDef{
			{
				Name:     "test",
				Address:  34,
				DataType: "int16",
			},
		}
		targetsGood = [...]Module{
			{DigitalOutput: regDefTest, Protocol: ModbusProtocolTCPIP},
		}
	)

	for _, s := range targetsGood {
		if err := s.validate(); err != nil {
			t.Errorf("validation of %v expected to pass but received the error:\n"+
				"%s", s.PrettyString(), err)
		}
	}
	for _, s := range targetsBad {
		if err := s.validate(); err == nil {
			t.Errorf("validation of %v expected to fail but it didn't.",
				s.PrettyString())
		}
	}
}

func BenchmarkPrettyPrint(b *testing.B) {
	s := Module{Parity: "O", Stopbits: 1, Databits: 7}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		s.PrettyString()
	}
}
