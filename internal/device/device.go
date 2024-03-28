package device

import (
	"context"
	"log"

	"fmt"
	"strconv"

	"git.miem.hse.ru/hubman/dmx-executor/internal/config"
	"git.miem.hse.ru/hubman/dmx-executor/internal/models"
	"github.com/redis/go-redis/v9"
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
	GetUniverseFromCache(ctx context.Context)
	SaveUniverseToCache(ctx context.Context)
	GetScenesFromCache(ctx context.Context)
	SaveScenesToCache(ctx context.Context)
	SetScene(ctx context.Context, command models.SetScene) error
	SaveScene(ctx context.Context) error
	SetChannel(ctx context.Context, command models.SetChannel) error
	IncrementChannel(ctx context.Context, command models.IncrementChannel) error
	WriteValueToChannel(command models.SetChannel) error
	WriteUniverseToDevice() error
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

func ReadUnvierse(ctx context.Context, deviceAlias string) ([512]byte, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	})

	var universe [512]byte
	key := fmt.Sprintf("%s_universe", deviceAlias)

	encodedUniverse, err := rdb.Get(ctx, key).Result()
	if err != nil {
		return [512]byte{}, fmt.Errorf("reading cached universe with key '%s' failed with error: %s", key, err)
	}

	universe, err = DecodeUniverse(encodedUniverse)
	if err != nil {
		return [512]byte{}, fmt.Errorf("universe with key '%s' decoding failed with error: %s", key, err)
	}

	return universe, nil
}

func WriteUniverse(ctx context.Context, deviceAlias string, universe []byte) error {
	rdb := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	})

	key := fmt.Sprintf("%s_universe", deviceAlias)
	var encodedUniverse = EncodeUniverse(universe, 512)

	_, err := rdb.Set(ctx, key, encodedUniverse, 0).Result()
	if err != nil {
		return fmt.Errorf("writing universe with key '%s' to cache failed with error: %s", key, err)
	}
	
	return nil
}

func ReadScenes(ctx context.Context, deviceAlias string, deviceScenes map[string]Scene) map[string]Scene {
	rdb := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	})
	
	encodedScenesMap := make(map[string]string)

	for sceneAlias, _ := range deviceScenes {
		key := fmt.Sprintf("%s_scene_%s", deviceAlias, sceneAlias)
		encodedScene, err := rdb.Get(ctx, key).Result()
		if err != nil {
			log.Printf("scene with key '%s' was not found in cache: %s", key, err)
			continue
		}
		encodedScenesMap[sceneAlias] = encodedScene
	}

	for sceneAlias, encodedScene := range encodedScenesMap {
		decodedScene, _ := DecodeScene(encodedScene)
		decodedScene.Alias = sceneAlias
		deviceScenes[sceneAlias] = decodedScene
	}

	return deviceScenes
}

func WriteScenes(ctx context.Context, deviceAlias string, deviceScenes map[string]Scene) error {
	rdb := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	})

	for sceneAlias, scene := range deviceScenes {

		key := fmt.Sprintf("%s_scene_%s", deviceAlias, sceneAlias)
		var encodedScene = EncodeScene(scene)

		_, err := rdb.Set(ctx, key, encodedScene, 0).Result()
		if err != nil {
			return fmt.Errorf("writing universe with key '%s' to cache failed with error: %s", key, err)
		}
	}
	
	return nil
}

func DecodeUniverse(sequence string) ([512]byte, error) {
	var universe [512]byte
	size := len(sequence)
	if size % 9 != 0 {
		return [512]byte{}, fmt.Errorf("got invalid RLE sequence size")
	}

	previousLastChannel := -1
	for i := 0; i < size; i+=9 {
		subsequence := sequence[i:i+9]
		initialChannel, err := strconv.Atoi(subsequence[0:3])
		if err != nil {
			return [512]byte{}, err
		}
		lastChannel, err := strconv.Atoi(subsequence[3:6])
		if err != nil {
			return [512]byte{}, err
		}
		channelValue, err := strconv.Atoi(subsequence[6:9])
		if err != nil {
			return [512]byte{}, err
		}
		if initialChannel > lastChannel {
			return [512]byte{}, fmt.Errorf("got invalid RLE sequence (initialChannel > lastChannel)")
		}

		if initialChannel < 0 || initialChannel > 511 {
			return [512]byte{}, fmt.Errorf("got invalid RLE sequence (initialChannel out of range [0:511])")
		}

		if lastChannel < 0 || lastChannel > 511 {
			return [512]byte{}, fmt.Errorf("got invalid RLE sequence (lastChannel out of range [0:511])")
		}

		if channelValue < 0 || channelValue > 511 {
			return [512]byte{}, fmt.Errorf("got invalid RLE sequence (channelValue out of range [0:511])")
		}

		if previousLastChannel >= initialChannel {
			return [512]byte{}, fmt.Errorf("got invalid RLE sequence (previousLastChannel > currentInitialChannel)")
		}

		for j := initialChannel; j <= lastChannel; j++ {
			universe[j] = byte(channelValue)
		} 

		previousLastChannel = lastChannel
	}

	return universe, nil
}

func EncodeUniverse(sequence []byte, size int) string {
	var result string
	var currentChannelValue int = int(sequence[0])
	var sequenceStart int = 0
	for idx, channel := range sequence {
		if currentChannelValue != int(channel){
			result += fmt.Sprintf("%03d", sequenceStart) + fmt.Sprintf("%03d", idx - 1) + fmt.Sprintf("%03d", currentChannelValue)
			currentChannelValue = int(channel)
			sequenceStart = idx
		}
		if idx == 511 {
			result += fmt.Sprintf("%03d", sequenceStart) + fmt.Sprintf("%03d", idx) + fmt.Sprintf("%03d", currentChannelValue)
		}
	}

	return result
}

func EncodeScene(scene Scene) string {
	var result string

	for sceneChannelID, channel := range scene.ChannelMap {
		result += fmt.Sprintf("%03d", sceneChannelID) + fmt.Sprintf("%03d", channel.UniverseChannelID) + fmt.Sprintf("%03d", channel.Value)
	}

	return result
}

func DecodeScene(sequence string) (Scene, error) {
	var result Scene
	result.ChannelMap = make(map[int]Channel)

	size := len(sequence)
	if size % 9 != 0 {
		return Scene{}, fmt.Errorf("got invalid RLE sequence size")
	}

	for i := 0; i < size; i+=9 {
		sceneChannelID, err := strconv.Atoi(sequence[i:i+3])
		if err != nil {
			return Scene{}, nil
		}

		universeChannelID, err := strconv.Atoi(sequence[i+3:i+6])
		if err != nil {
			return Scene{}, nil
		}

		channelValue, err := strconv.Atoi(sequence[i+6:i+9])
		if err != nil {
			return Scene{}, nil
		}

		if sceneChannelID < 0 || sceneChannelID > 511 {
			return Scene{}, fmt.Errorf("scene channel out of range [0:511]")
		}

		if universeChannelID < 0 || universeChannelID > 511 {
			return Scene{}, fmt.Errorf("universe channel out of range [0:511]")
		}

		if channelValue < 0 || channelValue > 511 {
			return Scene{}, fmt.Errorf("channel value out of range [0:511]")
		}

		result.ChannelMap[sceneChannelID] = Channel{UniverseChannelID: universeChannelID, Value: channelValue}
	}
	return result, nil
}