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

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/common/log"

	"github.com/lupoDharkael/modbus_exporter/modbus"
	"github.com/lupoDharkael/modbus_exporter/parser"

	"github.com/lupoDharkael/modbus_exporter/config"
)

var (
	listenAddress = flag.String("listen-address", ":9010",
		"The address to listen on for HTTP requests.")
	configFile = flag.String("config.file", "modbus.yml",
		"Sets the configuration file.")
	scrapeInterval = flag.Duration("scrape-interval", 8,
		"Sets scrape interval in seconds.")
)

func main() {
	flag.Parse()
	config.ScrapeInterval = time.Second * (*scrapeInterval)

	slavesFile, err := config.LoadSlaves(*configFile)
	if err != nil {
		log.Fatalln(err)
	}
	log.Infoln("Loading configuration file", *configFile)
	parsedSlaves, err := parser.ParseSlaves(slavesFile)
	if err != nil {
		log.Fatalln(err)
	}
	err = modbus.RegisterData(parsedSlaves, slavesFile)
	if err != nil {
		log.Fatalln(err)
	}
	// Expose the registered metrics via HTTP.
	http.Handle("/metrics", promhttp.Handler())

	log.Infoln("Listening on ", *listenAddress)
	log.Fatalln(http.ListenAndServe(*listenAddress, nil))

}
