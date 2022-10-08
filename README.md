# Modbus exporter

Prometheus exporter which retrieves stats from a modbus tcp system and exports
them via HTTP for Prometheus consumption.

![Scrape sequence](/scrape-sequence.svg "Scrape sequence")

<details>
 <summary>Reproduce diagram</summary>

 Go to: https://bramp.github.io/js-sequence-diagrams/

 ```
Note right of Prometheus: promehteus.yml \n --- \n target: Modbus-TCP-10.0.0.5 \n sub_target: Modbus-Unit-10 \n module: VendorXY
Prometheus->Exporter: http://xxx.de/metrics?target=10.0.0.5&sub_target=10&module=vendorxy
Note right of Exporter: modbus.yml \n --- \n module: VendorXY \n - temperature_a: 400001 \n - temperature_b: 400002

Exporter->Modbus_TCP_10.0.0.5: tcp://10.0.0.5?unit=10&register=400001
Modbus_TCP_10.0.0.5->Modbus_RTU_10: rtu://_?register=400001
Modbus_RTU_10-->Modbus_TCP_10.0.0.5: value=20
Modbus_TCP_10.0.0.5-->Exporter: value=20

Exporter->Modbus_TCP_10.0.0.5: tcp://10.0.0.5?unit=10&register=400002
Modbus_TCP_10.0.0.5->Modbus_RTU_10: rtu://_?register=400002
Modbus_RTU_10-->Modbus_TCP_10.0.0.5: value=19
Modbus_TCP_10.0.0.5-->Exporter: value=19

Exporter-->Prometheus:temperature_a{module="VendorXY",sub_target="10"} 20 \ntemperature_b{module="VendorXY",sub_target="10"} 19

 ```

</details>



## Building

```bash
make build
```


## Getting Started

The modbus exporter needs to be passed *target* (including port), *module* and *sub_module* as parameters
by Prometheus, this can be done with relabelling (see
[prometheus.yml](prometheus.yml)).

Once Prometheus is properly configured, run the exporter via:

```bash
./modbus_exporter [flags]
```

Supported flags:

[embedmd]:# (help.txt)
```txt
Usage of ./modbus_exporter:
  -config.file string
    	Sets the configuration file. (default "modbus.yml")
  -modbus-listen-address string
    	The address to listen on for HTTP requests exposing modbus metrics. (default ":9602")
  -telemetry-listen-address string
    	The address to listen on for HTTP requests exposing telemetry metrics about the exporter itself. (default ":9602")
```
Visit http://localhost:9602/modbus?target=1.2.3.4:502&module=fake&sub_target=1 where 1.2.3.4:502 is the IP and port number of the modbus IP device to get metrics from,
while module and sub_target parameters specify which module and subtarget to use from the config file.
If your device doesn't use sub-targets you can usually just set it to 1.

Visit http://localhost:9602/metrics to get the metrics of the exporter itself.

## Configuration File

Check out [`modbus.yml`](/modbus.yml) for more details on the configuration file
format.


## TODO

- Rework logging.

- Revisit bit parsing.

- Print name, version, ... on exporter startup.


# Misc info

## ModBus RTU

Support for serial ModBus (RTU) was dropped in git commit d06573828793094fd2bdf3e7c5d072e7a4fd381b.
Please send a PR if you need it again.
For now, we suggest using a ModBus PLC/bridge/master to bridge from RTU into TCP.

## Software provenance

This is forked from https://github.com/lupoDharkael/modbus_exporter which was not maintained any more and did not follow Prometheus best practices.
Initially, development happened in https://github.com/mxinden/modbus_exporter which has now been retired in favour of https://github.com/RichiH/modbus_exporter .
