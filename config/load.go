// Copyright 2017 Alejandro Sirgo Rica
//
// This file is part of GryphOn.
//
//     GryphOn is free software: you can redistribute it and/or modify
//     it under the terms of the GNU General Public License as published by
//     the Free Software Foundation, either version 3 of the License, or
//     (at your option) any later version.
//
//     GryphOn is distributed in the hope that it will be useful,
//     but WITHOUT ANY WARRANTY; without even the implied warranty of
//     MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
//     GNU General Public License for more details.
//
//     You should have received a copy of the GNU General Public License
//     along with GryphOn.  If not, see <http://www.gnu.org/licenses/>.

// Package config contains all the configuration related components
package config

import (
	"io/ioutil"

	yaml "gopkg.in/yaml.v2"
)

// LoadSlaves unmarshals the slaves configuration file.
func LoadSlaves(pathToSlaves string) (ListSlaves, error) {
	ls := make(ListSlaves)
	yamlFile, err := ioutil.ReadFile(pathToSlaves)
	if err == nil {
		err = yaml.Unmarshal(yamlFile, &ls)
	}
	return ls, err
}
