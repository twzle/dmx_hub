package device

import (
	"context"

	"git.miem.hse.ru/hubman/dmx-executor/internal/config"
	"git.miem.hse.ru/hubman/dmx-executor/internal/models"
)

type Channel struct {
	UniverseChannelID int
	Value             int
}

type Scene struct {
	Alias      string
	ChannelMap map[int]Channel
}

type Device interface {
	GetAlias() string
	SetUniverse(ctx context.Context, universe [512]byte)
	SetScene(ctx context.Context, sceneAlias string) error
	SaveScene(ctx context.Context)
	SetChannel(ctx context.Context, command models.SetChannel) error
	WriteValueToChannel(ctx context.Context, command models.SetChannel) error
	Blackout(ctx context.Context) error
}

func ReadScenesFromDeviceConfig(sceneListConfig []config.Scene) map[string]Scene {
	scenes := make(map[string]Scene)

	for _, sceneConfig := range sceneListConfig {
		scene := Scene{Alias: "", ChannelMap: make(map[int]Channel)}
		for _, channelMap := range sceneConfig.ChannelMap {
			channel := Channel{
				UniverseChannelID: int(channelMap.UniverseChannelID), 
				Value: 0}
			scene.ChannelMap[int(channelMap.SceneChannelID)] = channel
		}
		scene.Alias = sceneConfig.Alias
		scenes[scene.Alias] = scene
	}

	return scenes
}

func GetSceneById(scenes map[string]Scene, sceneId int) (*Scene){
	var i int
	for _, scene := range scenes {
		if i == sceneId {
			return &scene
		}
		i++
	}
	return nil
}
