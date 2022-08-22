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
