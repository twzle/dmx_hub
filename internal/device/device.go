package device

import (
	"context"

	"git.miem.hse.ru/hubman/dmx-executor/internal/config"
	"git.miem.hse.ru/hubman/dmx-executor/internal/models"
	"github.com/redis/go-redis/v9"
	"fmt"
	"strconv"
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
	GetUniverseFromCache(ctx context.Context) error
	SaveUniverseToCache(ctx context.Context) error
	SetScene(ctx context.Context, sceneAlias string) error
	SaveScene(ctx context.Context) error
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
		return [512]byte{}, err
	}

	universe, err = DecodeRLE(encodedUniverse)
	if err != nil {
		return [512]byte{}, err
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
	var encodedUniverse = EncodeRLE(universe, 512)

	_, err := rdb.Set(ctx, key, encodedUniverse, 0).Result()
	if err != nil {
		return err
	}
	
	return nil
}

func DecodeRLE(sequence string) ([512]byte, error) {
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

		fmt.Println(initialChannel, lastChannel)
		for j := initialChannel; j <= lastChannel; j++ {
			universe[j] = byte(channelValue)
		} 

		previousLastChannel = lastChannel
	}

	return universe, nil
}

func EncodeRLE(sequence []byte, size int) string {
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