# Modbus exporter
Exporter which retrieves stats from a modbus system and exports them via HTTP for Prometheus consumption.

**DO NOT USE THIS EXPORTER (NOT YET), IT REQUIRES MORE WORK IN ORDER TO BE RELIABLE.**

## Getting Started

To run it:

```bash
./modbus_exporter [flags]
```

The configuration will be taken from a configuration file, the exporter will search a fille called `slaves.yml` in the same directory by default.

Setting a different file:
```bash
./modbus_exporter -config.file="path/to/file" [flags]
```

Help on flags:

```bash
./modbus_exporter --help`
```

## Configuration File

Check the `examples/` file to read the information about the configuration file and some examples.

## General information
- Default values:
    + listen address :9010
    + Baud rate: 19200
    + Data bits:  8
    + Stop bits: 1
    + Parity: E - Even
- The use of no parity requires 2 stop bits.

## TODO
- General clean up
- Finish modbus logic
