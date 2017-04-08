package modbus

import "github.com/prometheus/client_golang/prometheus"

var (
	modbusDigital = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "modbus_digital_total",
			Help: "",
		},
		[]string{"slave", "type"},
	)

	modbusAnalog = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "modbus_analog_total",
			Help: "",
		},
		[]string{"slave", "type"},
	)
)

func init() {
	prometheus.MustRegister(modbusDigital)
	prometheus.MustRegister(modbusAnalog)
}
