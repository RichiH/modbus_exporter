package modbus

type metric struct {
	Name   string
	Help   string
	Labels map[string]string
	Value  float64
}
