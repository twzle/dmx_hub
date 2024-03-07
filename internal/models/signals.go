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


type SceneSaved struct {
	DeviceAlias string `hubman:"device_alias"`
	SceneAlias  string `hubman:"scene_alias"`
}

func (s SceneSaved) Code() string {
	return "SceneSaved"
}

func (s SceneSaved) Description() string {
	return "SceneSaved - signal represents event of successful scene save on a single DMX-compatible device"
}