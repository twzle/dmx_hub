package device

import (
	"encoding/json"
	"fmt"
	"net"
)

type ChannelMapConfig struct {
	SceneChannelID    uint16 `json:"scene_channel_id" yaml:"scene_channel_id"`
	UniverseChannelID uint16 `json:"universe_channel_id" yaml:"universe_channel_id"`
}

type SceneConfig struct {
	Alias      string             `json:"scene_alias" yaml:"scene_alias"`
	ChannelMap []ChannelMapConfig `json:"channel_map" yaml:"channel_map"`
}

type ArtNetConfig struct {
	Alias  string        `json:"alias" yaml:"alias"`
	IP     net.IP        `json:"ip" yaml:"ip"`
	Scenes []SceneConfig `json:"scenes" yaml:"scenes"`
}

type DMXConfig struct {
	Alias  string        `json:"alias" yaml:"alias"`
	Path   string        `json:"path" yaml:"path"`
	Scenes []SceneConfig `json:"scenes" yaml:"scenes"`
}

type UserConfig struct {
	DMXDevices    []DMXConfig    `json:"dmx_devices" yaml:"dmx_devices"`
	ArtNetDevices []ArtNetConfig `json:"artnet_devices" yaml:"artnet_devices"`
}

func (conf *UserConfig) Validate() error {
	if len(conf.DMXDevices) == 0 && len(conf.ArtNetDevices) == 0 {
		fmt.Println("DMX/ArtNet devices were not found in configuration file")
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
	for idx, device := range conf.ArtNetDevices {
		if device.Alias == "" {
			return fmt.Errorf("device #{%d} ({%s}): "+
				"valid ArtNet device_name must be provided in config",
				idx, device.Alias)
		}
		if device.IP == nil {
			return fmt.Errorf("device #{%d} ({%s}): "+
				"valid ArtNet IP address must be provided in config",
				idx, device.IP)
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

	for _, v := range conf.ArtNetDevices {
		if _, has := x[v.Alias]; has {
			return v.Alias, true
		}
		x[v.Alias] = struct{}{}
	}

	return "", false
}

func ParseConfigFromBytes(data []byte) (*UserConfig, error) {
	cfg := UserConfig{}

	err := json.Unmarshal(data, &cfg)
	if err != nil {
		return nil, err
	}

	return &cfg, nil
}

func ReadScenesFromDeviceConfig(sceneListConfig []SceneConfig) map[string]Scene {
	scenes := make(map[string]Scene)

	for _, sceneConfig := range sceneListConfig {
		scene := Scene{
			Alias:      "",
			ChannelMap: make(map[int]Channel)}
		for _, channelMap := range sceneConfig.ChannelMap {
			channel := Channel{
				UniverseChannelID: int(channelMap.UniverseChannelID),
				Value:             0}
			scene.ChannelMap[int(channelMap.SceneChannelID)] = channel
		}
		scene.Alias = sceneConfig.Alias
		scenes[scene.Alias] = scene
	}

	return scenes
}
