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

// Package config contains all the configuration related components
package config

import (
	"io/ioutil"

	yaml "gopkg.in/yaml.v2"
)

// LoadConfig unmarshals the targets configuration file.
func LoadConfig(pathToTargets string) (Config, error) {
	ls := Config{}
	yamlFile, err := ioutil.ReadFile(pathToTargets)
	if err != nil {
		return Config{}, err

	}

	err = yaml.Unmarshal(yamlFile, &ls)
	if err != nil {
		return Config{}, err
	}

	if err := ls.validate(); err != nil {
		return Config{}, err
	}

	return ls, nil
}
