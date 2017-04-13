package modbus

import "github.com/prometheus/client_golang/prometheus"

var (
	modbusDigitalIn = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "modbus_digital_total",
			Help: "Modbus digital input registers.",
		},
		[]string{"slave", "name"},
	)

	modbusAnalogIn = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "modbus_analog_total",
			Help: "Modbus analog input registers.",
		},
		[]string{"slave", "name"},
	)
	modbusDigitalOut = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "modbus_digital_total",
			Help: "Modbus digital output registers.",
		},
		[]string{"slave", "name"},
	)

	modbusAnalogOut = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "modbus_analog_total",
			Help: "Modbus analog output registers.",
		},
		[]string{"slave", "name"},
	)
)

func init() {
	prometheus.MustRegister(modbusDigitalIn)
	prometheus.MustRegister(modbusDigitalOut)
	prometheus.MustRegister(modbusAnalogIn)
	prometheus.MustRegister(modbusAnalogOut)
}
