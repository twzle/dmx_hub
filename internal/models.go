package internal

import (
	"fmt"
)

type DMXConfig struct {
	Alias string `json:"alias"`
	Path  string `json:"path"`
}

type RefreshConfig struct {
	DMXDevices []DMXConfig `json:"devices"`
}

func (o *RefreshConfig) Validate() error {
	for _, dmx := range o.DMXDevices {
		if dmx.Alias == "" {
			return fmt.Errorf("empty alias for dmx %v", dmx)
		}
		if dmx.Path == "" {
			return fmt.Errorf("empty path for dmx with alias %v", dmx.Path)
		}
	}
	return nil
}
