package models

type SceneChanged struct {
	DeviceAlias string `hubman:"device_alias"`
	SceneAlias  string `hubman:"scene_alias"`
}

func (s SceneChanged) Code() string {
	return "SceneChanged"
}

func (s SceneChanged) Description() string {
	return "SceneChanged - signal represents event of successful scene change on a single DMX-compatible device"
}
