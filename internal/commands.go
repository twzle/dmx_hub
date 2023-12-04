package internal

type SetChannel struct {
	Channel  int    `hubman:"channel"` // up to 512
	Value    byte   `hubman:"value"`
	DMXAlias string `hubman:"dmx_alias"`
}

func (s SetChannel) Code() string {
	return "SetChannel"
}

func (s SetChannel) Description() string {
	return "Sending value to chosen channel of given dmx-alias"
}

type Blackout struct {
	DMXAlias string `hubman:"dmx_alias"`
}

func (b Blackout) Code() string {
	return "Blackout"
}

func (b Blackout) Description() string {
	return "Clear all channels of given dmx-alias"
}
