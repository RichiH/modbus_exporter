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
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/common/log"

	"github.com/lupoDharkael/modbus_exporter/modbus"

	"github.com/lupoDharkael/modbus_exporter/config"
)

var (
	modbusAddress = flag.String("modbus-listen-address", ":9010",
		"The address to listen on for HTTP requests exposing modbus metrics.")
	telemetryAddress = flag.String("telemetry-listen-address", ":9011",
		"The address to listen on for HTTP requests exposing telemetry metrics about the exporter itself.")
	configFile = flag.String("config.file", "modbus.yml",
		"Sets the configuration file.")
	scrapeInterval = flag.Duration("scrape-interval", 8,
		"Sets scrape interval in seconds.")
)

func main() {
	flag.Parse()
	config.ScrapeInterval = time.Second * (*scrapeInterval)

	telemetryRegistry := prometheus.NewRegistry()
	telemetryRegistry.MustRegister(prometheus.NewGoCollector())
	telemetryRegistry.MustRegister(prometheus.NewProcessCollector(prometheus.ProcessCollectorOpts{}))

	log.Infoln("Loading configuration file", *configFile)
	config, err := config.LoadConfig(*configFile)
	if err != nil {
		log.Fatalln(err)
	}

	log.Infoln("telemetry metrics at: " + *telemetryAddress)
	go func() {
		log.Fatal(
			http.ListenAndServe(*telemetryAddress, promhttp.HandlerFor(telemetryRegistry, promhttp.HandlerOpts{})),
		)
	}()

	exporter := modbus.NewExporter(config)

	handler := func(w http.ResponseWriter, r *http.Request) {
		moduleName := r.URL.Query().Get("module")

		target := r.URL.Query().Get("target")

		log.Infof("got scrape request for module '%v' and target '%v'", moduleName, target)

		gatherer, err := exporter.Scrape(target, moduleName)
		if err != nil {
			log.Fatal(err)
		}

		promhttp.HandlerFor(gatherer, promhttp.HandlerOpts{}).ServeHTTP(w, r)
	}

	log.Infoln("Modbus metrics at: " + *modbusAddress)
	log.Fatal(
		http.ListenAndServe(*modbusAddress, http.HandlerFunc(handler)),
	)
}
