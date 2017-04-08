# Modbus exporter
Exporter which retrieves stats from a modbus system and exports them via HTTP for Prometheus consumption.

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

```yaml
Arduino1:
port: "/dev/ttyUSB0"
id: 1
timeout: 1000
baudrate: 19200
databits: 8
stopbits: 1
parity: N
discreteInputs:
- 55:60
inputRegisters:
- 2:5 = Temperature$...
- 53, 101, 102, 154 = Fog$, Humidity_$, Fog
Arduino2:
port: "localhost:9090"
id: 2
timeout: 1000
coils:
- 2 = Light_$
- 34:56 =sensor$,sensorA$,sensorA$
holdingRegisters:
- 62:65 = motor_F, motor_F, motor_G, motor_G
- 66 = otherSensor
```

The format of the file is yaml so you can apply its rules when writing the file (comments with # and so on)
Your are allowed to use the “enumerator” with the symbol $ at the end of the names. When the assignation occurs the name receives an index at the end. The first unique number with enumerator will be named with a 0. As the parser finds more identical names with the enumerator, it adds a number incremented by 1 with respect to the previous. sensorA$, sensorA$ would be translated to sensorA1, sensorA2.

## General information
- Default values:
    + listen address :9010
    + Baud rate: 19200
    + Data bits:  8
    + Stop bits: 1
    + Parity: E - Even
- The use of no parity requires 2 stop bits.

## TODO
- Implement loging with Logrus
- Improve help format
- General clean up
- Finish modbus logic
