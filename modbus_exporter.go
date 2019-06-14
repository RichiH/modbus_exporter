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

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/common/log"

	"github.com/lupoDharkael/modbus_exporter/config"
	"github.com/lupoDharkael/modbus_exporter/modbus"
)

var (
	modbusAddress = flag.String("modbus-listen-address", ":9010",
		"The address to listen on for HTTP requests exposing modbus metrics.")
	telemetryAddress = flag.String("telemetry-listen-address", ":9011",
		"The address to listen on for HTTP requests exposing telemetry metrics about the exporter itself.")
	configFile = flag.String("config.file", "modbus.yml",
		"Sets the configuration file.")
)

func main() {
	flag.Parse()

	telemetryRegistry := prometheus.NewRegistry()
	telemetryRegistry.MustRegister(prometheus.NewGoCollector())
	telemetryRegistry.MustRegister(prometheus.NewProcessCollector(prometheus.ProcessCollectorOpts{}))

	log.Infoln("Loading configuration file", *configFile)
	config, err := config.LoadConfig(*configFile)
	if err != nil {
		log.Fatalln(err)
	}

	go func() {
		telemetryEndpoint(telemetryRegistry, *telemetryAddress)
	}()

	exporter := modbus.NewExporter(config)

	router := http.NewServeMux()
	router.Handle("/metrics",
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			scrapeHandler(exporter, w, r)
		}),
	)

	log.Infoln("Modbus metrics at: " + *modbusAddress)
	log.Fatal(http.ListenAndServe(*modbusAddress, router))
}

func telemetryEndpoint(registry prometheus.Gatherer, address string) {
	router := http.NewServeMux()
	router.Handle("/metrics", promhttp.HandlerFor(registry, promhttp.HandlerOpts{}))

	log.Infoln("telemetry metrics at: " + address)
	log.Fatal(http.ListenAndServe(address, router))
}

func scrapeHandler(e *modbus.Exporter, w http.ResponseWriter, r *http.Request) {
	moduleName := r.URL.Query().Get("module")
	if moduleName == "" {
		http.Error(w, "'module' parameter must be specified", http.StatusBadRequest)
		return
	}

	if !e.GetConfig().HasModule(moduleName) {
		http.Error(w, fmt.Sprintf("module '%v' not defined in configuration file", moduleName), http.StatusBadRequest)
		return
	}

	target := r.URL.Query().Get("target")
	if target == "" {
		http.Error(w, "'target' parameter must be specified", http.StatusBadRequest)
		return
	}

	log.Infof("got scrape request for module '%v' and target '%v'", moduleName, target)

	gatherer, err := e.Scrape(target, moduleName)
	if err != nil {
		http.Error(
			w,
			fmt.Sprintf("failed to scrape target '%v' with module '%v': %v", target, moduleName, err),
			http.StatusInternalServerError,
		)
		log.Error(err)
		return
	}

	promhttp.HandlerFor(gatherer, promhttp.HandlerOpts{}).ServeHTTP(w, r)
}
