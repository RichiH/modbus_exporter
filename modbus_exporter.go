package main

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/common/log"
	"github.com/prometheus/exporter-toolkit/web"
	webflag "github.com/prometheus/exporter-toolkit/web/kingpinflag"
	kingpin "gopkg.in/alecthomas/kingpin.v2"

	"github.com/lupoDharkael/modbus_exporter/config"
	"github.com/lupoDharkael/modbus_exporter/modbus"
)

func main() {
	modbusAddress := kingpin.Flag(
		"modbus-listen-address",
		"The address to listen on for HTTP requests exposing modbus metrics.",
	).Default(":9602").String()
	telemetryAddress := kingpin.Flag(
		"telemetry-listen-address",
		"The address to listen on for HTTP requests exposing telemetry metrics about the exporter itself.",
	).Default(":9602").String()
	configFile := kingpin.Flag(
		"config.file",
		"Sets the configuration file.",
	).Default("modbus.yml").String()
	log.AddFlags(kingpin.CommandLine)
	webConfig := webflag.AddFlags(kingpin.CommandLine)

	kingpin.HelpFlag.Short('h')
	kingpin.Parse()

	telemetryRegistry := prometheus.NewRegistry()
	telemetryRegistry.MustRegister(prometheus.NewGoCollector())
	telemetryRegistry.MustRegister(prometheus.NewProcessCollector(prometheus.ProcessCollectorOpts{}))

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
	srv := &http.Server{Addr: *modbusAddress}
	log.Fatal(web.ListenAndServe(srv, *webConfig, log))
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

	gatherer, err := e.Scrape(target, byte(subTarget), moduleName)
	if err != nil {
		httpStatus := http.StatusInternalServerError
		if strings.Contains(fmt.Sprintf("%v", err), "unable to connect with target") {
			httpStatus = http.StatusServiceUnavailable
		} else if strings.Contains(fmt.Sprintf("%v", err), "i/o timeout") {
			httpStatus = http.StatusGatewayTimeout
		}
		http.Error(
			w,
			fmt.Sprintf("failed to scrape target '%v' with module '%v': %v", target, moduleName, err),
			httpStatus,
		)
		log.Error(err)
		return
	}

	promhttp.HandlerFor(gatherer, promhttp.HandlerOpts{}).ServeHTTP(w, r)
}
