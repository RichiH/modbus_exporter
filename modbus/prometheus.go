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

package modbus

import "github.com/prometheus/client_golang/prometheus"

var (
	modbusDigitalIn = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "modbus_digital_input_total",
			Help: "Modbus digital input registers.",
		},
		[]string{"slave", "name"},
	)

	modbusAnalogIn = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "modbus_analog_input_total",
			Help: "Modbus analog input registers.",
		},
		[]string{"slave", "name"},
	)
	modbusDigitalOut = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "modbus_digital_output_total",
			Help: "Modbus digital output registers.",
		},
		[]string{"slave", "name"},
	)

	modbusAnalogOut = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "modbus_analog_output_total",
			Help: "Modbus analog output registers.",
		},
		[]string{"slave", "name"},
	)
)

// RegisterMetrics registers modbus specific metrics at the given registerer.
func RegisterMetrics(r prometheus.Registerer) {
	r.MustRegister(modbusDigitalIn)
	r.MustRegister(modbusDigitalOut)
	r.MustRegister(modbusAnalogIn)
	r.MustRegister(modbusAnalogOut)
}
