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

package main

import (
	"flag"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/common/log"

	"github.com/lupoDharkael/modbus_exporter/config"
	"github.com/lupoDharkael/modbus_exporter/modbus"
)

// ModbusRequestStatusType possible status of the modbus request
type ModbusRequestStatusType string

const (
	// ModbusRequestStatusOK successful
	ModbusRequestStatusOK ModbusRequestStatusType = "OK"
	// ModbusRequestStatusErrorSock error opening socket connection
	ModbusRequestStatusErrorSock ModbusRequestStatusType = "ERROR_SOCKET"
	// ModbusRequestStatusErrorTimeout connection established but no response from modbus device
	ModbusRequestStatusErrorTimeout ModbusRequestStatusType = "ERROR_TIMEOUT"
	// ModbusRequestStatusErrorParsingValue error parsing value received
	ModbusRequestStatusErrorParsingValue ModbusRequestStatusType = "ERROR_PARSING_VALUE"
)

var (
	modbusDurationCounterVec *prometheus.CounterVec
	modbusRequestsCounterVec *prometheus.CounterVec
	mutex sync.Mutex
	mutexWaiters uint64 = 0;
	modbusSerialMutexDurationCounterVec *prometheus.CounterVec
	modbusSerialMutexWaitersGaugeVec *prometheus.GaugeVec
	modbusSerialRetriesCounterVec *prometheus.CounterVec
)

func main() {
	modbusAddress := flag.String("modbus-listen-address", ":9602",
		"The address to listen on for HTTP requests exposing modbus metrics.")
	telemetryAddress := flag.String("telemetry-listen-address", ":9602",
		"The address to listen on for HTTP requests exposing telemetry metrics about the exporter itself.")
	configFile := flag.String("config.file", "modbus.yml",
		"Sets the configuration file.")

	flag.Parse()

	telemetryRegistry := prometheus.NewRegistry()
	telemetryRegistry.MustRegister(prometheus.NewGoCollector())
	telemetryRegistry.MustRegister(prometheus.NewProcessCollector(prometheus.ProcessCollectorOpts{}))

	modbusDurationCounterVec = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "modbus_request_duration_seconds_total",
		Help: "Total duration of modbus successful requests by target in seconds",
	}, []string{"target"})
	telemetryRegistry.MustRegister(modbusDurationCounterVec)

	modbusSerialMutexDurationCounterVec = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "modbus_request_serial_mutex_duration_seconds_total",
		Help: "Total duration of waiting for mutex lock for serial bus by target in seconds",
	}, []string{"target"})
	telemetryRegistry.MustRegister(modbusSerialMutexDurationCounterVec)

	modbusSerialMutexWaitersGaugeVec = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "modbus_request_serial_mutex_waiters",
		Help: "Total number of threads currently waiting for mutex lock by serial bus",
	}, []string{"target"})
	telemetryRegistry.MustRegister(modbusSerialMutexWaitersGaugeVec)

	modbusSerialRetriesCounterVec = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "modbus_request_serial_retries_total",
		Help: "Total number of serial retries following errors by serial bus and modbus_target",
	}, []string{"target", "modbus_target"})
	telemetryRegistry.MustRegister(modbusSerialRetriesCounterVec)

	modbusRequestsCounterVec = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "modbus_requests_total",
		Help: "Number of modbus request by status and target",
	}, []string{"target", "status"})
	telemetryRegistry.MustRegister(modbusRequestsCounterVec)

	log.Infoln("Loading configuration file", *configFile)
	config, err := config.LoadConfig(*configFile)
	if err != nil {
		log.Fatalln(err)
	}
	router := http.NewServeMux()
	router.Handle("/metrics", promhttp.HandlerFor(telemetryRegistry, promhttp.HandlerOpts{}))
	log.Infoln("telemetry metrics at: " + *telemetryAddress)
	exporter := modbus.NewExporter(config)
	router.Handle("/modbus",
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			scrapeHandler(exporter, w, r)
		}),
	)

	log.Infoln("Modbus metrics at: " + *modbusAddress)
	log.Fatal(http.ListenAndServe(*modbusAddress, router))
}

func scrapeHandler(e *modbus.Exporter, w http.ResponseWriter, r *http.Request) {
	moduleName := r.URL.Query().Get("module")
	if moduleName == "" {
		http.Error(w, "'module' parameter must be specified", http.StatusBadRequest)
		return
	}

	module := e.GetConfig().GetModule(moduleName)
	if module == nil {
		http.Error(w, fmt.Sprintf("module '%v' not defined in configuration file", moduleName), http.StatusBadRequest)
		return
	}

	target := r.URL.Query().Get("target")
	if target == "" {
		http.Error(w, "'target' parameter must be specified", http.StatusBadRequest)
		return
	}

	sT := r.URL.Query().Get("sub_target")
	if sT == "" {
		http.Error(w, "'sub_target' parameter must be specified", http.StatusBadRequest)
		return
	}

	subTarget, err := strconv.Atoi(sT)
	if err != nil {
		http.Error(w, fmt.Sprintf("'sub_target' parameter must be a valid integer: %v", err), http.StatusBadRequest)
		return
	}

	log.Infof("got scrape request for module '%v' target '%v' and sub_target '%v'", moduleName, target, subTarget)

	start := time.Now()
	if module.Protocol == config.ModbusProtocolSerial {
		if atomic.LoadUint64(&mutexWaiters) > 0 {
			log.Infof("Serial port '%v' is busy, waiting for module '%v' to scrape sub_target '%v' (%d other threads are also waiting).", target, moduleName, subTarget, atomic.LoadUint64(&mutexWaiters))
		}
		atomic.AddUint64(&mutexWaiters, 1)
		modbusSerialMutexWaitersGaugeVec.WithLabelValues(target).Set(float64(mutexWaiters))
		mutex.Lock()
		atomic.AddUint64(&mutexWaiters, ^uint64(0))
		modbusSerialMutexWaitersGaugeVec.WithLabelValues(target).Set(float64(mutexWaiters))
		modbusSerialMutexDurationCounterVec.WithLabelValues(target).Add(time.Since(start).Seconds())
	}
	gatherer, err := e.Scrape(target, byte(subTarget), moduleName)
	if module.Protocol == config.ModbusProtocolSerial {
		// retry up to two times when a serial scrape fails
		if err != nil {
			modbusSerialRetriesCounterVec.WithLabelValues(target, string(subTarget)).Inc()
			gatherer, err = e.Scrape(target, byte(subTarget), moduleName)
		}
		if err != nil {
			modbusSerialRetriesCounterVec.WithLabelValues(target, string(subTarget)).Inc()
			gatherer, err = e.Scrape(target, byte(subTarget), moduleName)
		}
		mutex.Unlock()
	}
	duration := time.Since(start).Seconds()
	if err != nil {
		if strings.Contains(fmt.Sprintf("%v", err), "unable to connect with target") {
			modbusRequestsCounterVec.WithLabelValues(target, string(ModbusRequestStatusErrorSock)).Inc()
		} else if strings.Contains(fmt.Sprintf("%v", err), "i/o timeout") {
			modbusRequestsCounterVec.WithLabelValues(target, string(ModbusRequestStatusErrorTimeout)).Inc()
		} else {
			modbusRequestsCounterVec.WithLabelValues(target, string(ModbusRequestStatusErrorParsingValue)).Inc()
		}
		http.Error(
			w,
			fmt.Sprintf("failed to scrape target '%v' sub_target '%d' with module '%v': %v", target, subTarget, moduleName, err),
			http.StatusInternalServerError,
		)
		log.Error(err)
		return
	}
	modbusDurationCounterVec.WithLabelValues(target).Add(duration)
	modbusRequestsCounterVec.WithLabelValues(target, string(ModbusRequestStatusOK)).Inc()

	promhttp.HandlerFor(gatherer, promhttp.HandlerOpts{}).ServeHTTP(w, r)
}
