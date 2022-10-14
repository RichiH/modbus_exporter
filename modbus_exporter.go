package main

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/common/promlog"
	"github.com/prometheus/common/promlog/flag"
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

	promlogConfig := &promlog.Config{}
	flag.AddFlags(kingpin.CommandLine, promlogConfig)
	webConfig := webflag.AddFlags(kingpin.CommandLine)
	kingpin.HelpFlag.Short('h')
	kingpin.Parse()
	logger := promlog.New(promlogConfig)

	telemetryRegistry := prometheus.NewRegistry()
	telemetryRegistry.MustRegister(prometheus.NewGoCollector())
	telemetryRegistry.MustRegister(prometheus.NewProcessCollector(prometheus.ProcessCollectorOpts{}))

	level.Info(logger).Log("Loading configuration file", *configFile)
	config, err := config.LoadConfig(*configFile)
	if err != nil {
		level.Error(logger).Log("err", err)
	}
	router := http.NewServeMux()
	router.Handle("/metrics", promhttp.HandlerFor(telemetryRegistry, promhttp.HandlerOpts{}))
        level.Info(logger).Log("msg", "telemetry metrics at: " + *telemetryAddress)
	exporter := modbus.NewExporter(config)
	router.Handle("/modbus",
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			scrapeHandler(exporter, w, r, logger)
		}),
	)

	level.Info(logger).Log("msg", "Modbus metrics at: " + *modbusAddress)
	srv := &http.Server{Addr: *modbusAddress, Handler: router}
	level.Error(logger).Log("err", web.ListenAndServe(srv, *webConfig, logger))
}

func scrapeHandler(e *modbus.Exporter, w http.ResponseWriter, r *http.Request, logger log.Logger) {
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

	level.Info(logger).Log("msg", "got scrape request", "module", moduleName, "target", target, "sub_target", subTarget)

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
		level.Error(logger).Log("err", err)
		return
	}

	promhttp.HandlerFor(gatherer, promhttp.HandlerOpts{}).ServeHTTP(w, r)
}
