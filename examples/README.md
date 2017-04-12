# Modbus Configuration File

This exporter needs a configuration file because the modbus protocol requires a lot of parameters in order to start a communication.
The format of the file is yaml so you can apply its rules when writing the file.

## Slave Parameters

### TCP/IP
TO DO

### Serial
TO DO

## Register Definition
TO DO

### Naming Registers
TO DO

### Range
TO DO

### Ellipsis
TO DO

### Enumeration
The symbol `$` at the end of a name assigns an index it. The first unique number with enumerator will be named with a 0. As the parser finds more identical names with the enumerator, it adds that number incremented by 1 with respect to the previous. 
`sensorA$, sensorA$` would be translated to `sensorA1, sensorA2`.

### Naming Examples
TO DO
