package config

import (
	"fmt"
	"gopkg.in/yaml.v3"
	"os"
)

type DMXConfig struct {
	Alias string `json:"alias" yaml:"alias"`
	Path  string `json:"path" yaml:"path"`
}

type UserConfig struct {
	DMXDevices []DMXConfig `json:"dmx_devices" yaml:"dmx_devices"`
}

func (conf *UserConfig) Validate() error {
	if len(conf.DMXDevices) == 0 {
		fmt.Println("DMX devices were not found in configuration file")
	}
	if alias, has := conf.hasDuplicateDevices(); has {
		return fmt.Errorf("found duplicate DMX device with alias {%s} in config", alias)
	}
	for idx, device := range conf.DMXDevices {
		if device.Alias == "" {
			return fmt.Errorf("device #{%d} ({%s}): "+
				"valid DMX device_name must be provided in config",
				idx, device.Alias)
		}
	}
	return nil
}

func (conf *UserConfig) hasDuplicateDevices() (string, bool) {
	x := make(map[string]struct{})

	for _, v := range conf.DMXDevices {
		if _, has := x[v.Alias]; has {
			return v.Alias, true
		}
		x[v.Alias] = struct{}{}
	}

	return "", false
}

func InitConfig(confPath string) (*UserConfig, error) {
	jsonFile, err := os.ReadFile(confPath)
	if err != nil {
		return nil, err
	}
	cfg, err := ParseConfigFromBytes(jsonFile)
	return cfg, err
}

func ParseConfigFromBytes(data []byte) (*UserConfig, error) {
	cfg := UserConfig{}

	err := yaml.Unmarshal(data, &cfg)
	if err != nil {
		return nil, err
	}

	err = cfg.Validate()
	if err != nil {
		return nil, err
	}

	return &cfg, nil
}
