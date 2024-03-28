package models

type SetChannel struct {
	Channel     int    `hubman:"channel"` // up to 512
	Value       int    `hubman:"value"`
	DeviceAlias string `hubman:"device_alias"`
}

func (s SetChannel) Code() string {
	return "SetChannel"
}

func (s SetChannel) Description() string {
	return "Sending value to chosen channel of single DMX/Artnet device by alias"
}

type IncrementChannel struct {
	Channel     int    `hubman:"channel"` // up to 512
	Value       int    `hubman:"value"`
	DeviceAlias string `hubman:"device_alias"`
}

func (i IncrementChannel) Code() string {
	return "IncrementChannel"
}

func (i IncrementChannel) Description() string {
	return "Incrementing value of chosen channel by specified value of single DMX/Artnet device by alias"
}

type Blackout struct {
	DeviceAlias string `hubman:"device_alias"`
}

func (b Blackout) Code() string {
	return "Blackout"
}

func (b Blackout) Description() string {
	return "Clear all channels of given dmx-alias"
}

type SetScene struct {
	DeviceAlias string `hubman:"device_alias"`
	SceneAlias  string `hubman:"scene_alias"`
}

func (s SetScene) Code() string {
	return "SetScene"
}

func (s SetScene) Description() string {
	return "Sets scene by alias for single DMX/Artnet device"
}

type SaveScene struct {
	DeviceAlias string `hubman:"device_alias"`
}

func (s SaveScene) Code() string {
	return "SaveScene"
}

func (s SaveScene) Description() string {
	return "Saves current dmx scene for single DMX/Artnet device"
}
