module github.com/RichiH/modbus_exporter

go 1.19

require (
	github.com/alecthomas/kingpin/v2 v2.3.2
	github.com/go-kit/log v0.2.1
	github.com/goburrow/modbus v0.1.0
	github.com/goburrow/serial v0.1.0 // indirect
	github.com/hashicorp/go-multierror v0.0.0-20161216184304-ed905158d874
	github.com/prometheus/client_golang v1.14.0
	github.com/prometheus/common v0.41.0
	github.com/prometheus/exporter-toolkit v0.9.1
	github.com/tbrandon/mbserver v0.0.0-20170611213546-993e1772cc62
	gopkg.in/yaml.v2 v2.4.0
)

require github.com/hashicorp/errwrap v0.0.0-20141028054710-7554cd9344ce // indirect
