// Copyright 2019 Richard Hartmann
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package config

import (
	"os"
	"path/filepath"

	yaml "gopkg.in/yaml.v2"
)

// LoadConfig unmarshals the targets configuration file(s).
func LoadConfig(pathToTargets []string) (Config, error) {
	fullConfig := Config{}
	for _, p := range pathToTargets {
		files, err := filepath.Glob(p)
		if err != nil {
			return Config{}, err
		}
		for _, f := range files {
			ls := Config{}
			yamlFile, err := os.ReadFile(f)
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

			fullConfig.Modules = append(fullConfig.Modules, ls.Modules...)
		}
	}

	return fullConfig, nil
}
