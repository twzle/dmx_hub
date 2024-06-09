package models

// Represenation of scene changed signal
type SceneChanged struct {
	DeviceAlias string `hubman:"device_alias"`
	SceneAlias  string `hubman:"scene_alias"`
}

// Function returns string code of signal
func (s SceneChanged) Code() string {
	return "SceneChanged"
}

// Function returns string description of signal
func (s SceneChanged) Description() string {
	return "SceneChanged - signal represents event of successful scene change on a single DMX-compatible device"
}

// Represenation of scene saved signal
type SceneSaved struct {
	DeviceAlias string `hubman:"device_alias"`
	SceneAlias  string `hubman:"scene_alias"`
}

// Function returns string code of signal
func (s SceneSaved) Code() string {
	return "SceneSaved"
}
// Function returns string description of signal
func (s SceneSaved) Description() string {
	return "SceneSaved - signal represents event of successful scene save on a single DMX-compatible device"
}