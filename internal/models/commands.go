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
	return "Sending value to chosen channel of given dmx-alias"
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
