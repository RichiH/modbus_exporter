# Modbus exporter
Exporter which retrieves stats from a modbus tcp system and exports them via HTTP for Prometheus consumption.


## Building

```bash
go build
```


## Getting Started

To run it:

```bash
./modbus_exporter [flags]
```

The configuration will be taken from a configuration file, the exporter will search for a file called `modbus.yml` in the same directory by default.

Setting a different file and a different listen address:
```bash
./modbus_exporter -config.file="path/to/file" -modbus-listen-address=":8080"
```

Help on flags:

```bash
./modbus_exporter --help
```


## Configuration File

Check the `examples/` folder to read the information about the configuration file and some examples.


## TODO

- Rework logging.

- Revisit bit parsing.

- Introduce metric type 'gauge', 'counter' in config yaml.

- Make metric name and labels configurable per metric definition.
