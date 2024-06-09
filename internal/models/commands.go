package models

// Represenation of set channel command
type SetChannel struct {
	Channel     int    `hubman:"channel"` // up to 512
	Value       int    `hubman:"value"`
	DeviceAlias string `hubman:"device_alias"`
}

// Function returns string code of command
func (s SetChannel) Code() string {
	return "SetChannel"
}

// Function returns string description of command
func (s SetChannel) Description() string {
	return "Sending value to chosen channel of single DMX/Artnet device by alias"
}

// Represenation of increment channel command
type IncrementChannel struct {
	Channel     int    `hubman:"channel"` // up to 512
	Value       int    `hubman:"value"`
	DeviceAlias string `hubman:"device_alias"`
}

// Function returns string code of command
func (i IncrementChannel) Code() string {
	return "IncrementChannel"
}

// Function returns string description of command
func (i IncrementChannel) Description() string {
	return "Incrementing value of chosen channel by specified value of single DMX/Artnet device by alias"
}

// Represenation of blackout command
type Blackout struct {
	DeviceAlias string `hubman:"device_alias"`
}

// Function returns string code of command
func (b Blackout) Code() string {
	return "Blackout"
}

// Function returns string description of command
func (b Blackout) Description() string {
	return "Clear all channels of given dmx-alias"
}

// Represenation of set scene command
type SetScene struct {
	DeviceAlias string `hubman:"device_alias"`
	SceneAlias  string `hubman:"scene_alias"`
}

// Function returns string code of command
func (s SetScene) Code() string {
	return "SetScene"
}

// Function returns string description of command
func (s SetScene) Description() string {
	return "Sets scene by alias for single DMX/Artnet device"
}

// Represenation of save scene command
type SaveScene struct {
	DeviceAlias string `hubman:"device_alias"`
}

// Function returns string code of command
func (s SaveScene) Code() string {
	return "SaveScene"
}

// Function returns string description of command
func (s SaveScene) Description() string {
	return "Saves current dmx scene for single DMX/Artnet device"
}
