# Modbus Configuration File

This exporter needs a configuration file because the modbus protocol requires a lot of parameters in order to start a communication.
The format of the file is yaml so you can apply its rules when writing the file.

## Slave Parameters
### General
- timeout: sets the timeout in milliseconds.
- id: assigns the slave's id.

### TCP/IP
- port: you have to define a valid IP address. e.g. `"localhost:8080"` or `"192.168.0.192:9090"`.
- keepAlive: if the slave handles a keep alive connection. Boolean `true` or `false`. Default value = `false`

### Serial
- port: you have to define a valid port. e.g. `"/dev/ttyUSB0"`.
- baudrate: numeric value of baudrate.  Default value = `19200`.
- databits: numeric value of databits.  Default value = `8`.
- stopbits: numeric value of stopbits.  Default value = `1`.
- parity: type of parity, the valid values are `'O'`(odd), `'N'`(none) and `'E'`(even). Default value = `'E'`.

## Register Definition
Every register must be defined by its type, you can group them as `analogOut`, `analogIn`, `digitalIn` and `digitalOut`.
There are 2 ways to define the registers to be processed.
- Enumeration: coma separated registers of the slaves, it can be a single name. e.g. `53, 101, 102, 154` or `34`
- Range: a register value and a higher one separated by a colon, it defines the whole range of register (it is inclusive). e.g. `20:25`, that would be translated to `20,21,22,23,24,25`.

General example:
```yml
digitalIn:
- 55:60
analogOut:
- 2:5
- 53, 101, 102, 154
```

### Naming Registers
Registers are named automatically based on its type and register number.
For example:
```yml
digitalIn:
- 55
```
here the defined register would be named `DIn_55`, as you can see it shows the type first (`DIn`, `DOut`, `AIn`, `AOut`) and then just adds an underscore and the register number.

If you want to define a specific name you have to assign it with the `=` operator.

The name definition is done by coma separated values and they are assigned one by one to the registers, you can define less names than registers but not the opposite. That is valid for enumerated and range registers.

e.g.
```yml
- 53, 101, 102, 154 = tempSensor, Humidity, lightSensor
```
### Ellipsis
The ellipsis is very useful, it repeats the naming sequence through the whole available registers in the statement.

You can use it to give an unique name to a range of registers:
```yml
- 50:100 = uniqueName...
```

Or to generate alternate sequences:
```yml
- 53, 101, 102, 154 = typeA, typeB...
```
in this one we'd have `53=typeA`, `101=typeB`, `102=typeA` and `154=typeB`.

### Enumeration
The symbol `$` at the end of a name assigns an index it. The first unique number with enumerator will be named with a 0. As the parser finds more identical names with the enumerator, it adds that number incremented by 1 with respect to the previous.
`sensorA$, sensorA$` would be translated to `sensorA1, sensorA2`.

The enumeration can be combined with the ellipsis in a very practical way:
```yml
 - 2:12 = Temperature$...
```

### Examples
Check the examples in this directory.
