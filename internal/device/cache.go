package device

import (
	"context"
	"fmt"
	"strconv"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)
func (b *BaseDevice) ReadUnvierse(ctx context.Context) error {
	rdb := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	})

	key := fmt.Sprintf("%s_universe", b.Alias)

	encodedUniverse, err := rdb.Get(ctx, key).Result()
	if err != nil {
		return fmt.Errorf("reading cached universe with key '%s' failed with error: %s", key, err)
	}

	err = b.DecodeUniverse(encodedUniverse)
	if err != nil {
		b.Universe = [512]byte{}
		return err
	}

	return nil
}

func (b *BaseDevice) WriteUniverse(ctx context.Context) error {
	rdb := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	})

	key := fmt.Sprintf("%s_universe", b.Alias)
	var encodedUniverse = b.EncodeUniverse()

	_, err := rdb.Set(ctx, key, encodedUniverse, 0).Result()
	if err != nil {
		return fmt.Errorf("writing universe with key '%s' to cache failed with error: %s", key, err)
	}

	return nil
}

func (b *BaseDevice) ReadScenes(ctx context.Context){
	rdb := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	})
	
	encodedScenesMap := make(map[string]string)

	for sceneAlias, _ := range b.Scenes {
		key := fmt.Sprintf("%s_scene_%s", b.Alias, sceneAlias)
		encodedScene, err := rdb.Get(ctx, key).Result()
		if err != nil {
			b.Logger.Warn(fmt.Sprintf("reading scene '%s' from cache failed", sceneAlias), zap.Error(err), zap.Any("device", b.Alias))
			continue
		}
		encodedScenesMap[sceneAlias] = encodedScene
	}

	for sceneAlias, encodedScene := range encodedScenesMap {
		err := b.DecodeScene(encodedScene, b.Scenes[sceneAlias])
		if err != nil {
			b.Scenes[sceneAlias] = Scene{}
			b.Logger.Warn(fmt.Sprintf("decoding cached scene '%s' failed", sceneAlias), zap.Error(err), zap.Any("device", b.Alias))
		}
	}
}

func (b *BaseDevice) WriteScenes(ctx context.Context) {
	rdb := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	})

	for sceneAlias, scene := range b.Scenes {

		key := fmt.Sprintf("%s_scene_%s", b.Alias, sceneAlias)
		var encodedScene = b.EncodeScene(scene)

		_, err := rdb.Set(ctx, key, encodedScene, 0).Result()
		if err != nil {
			b.Logger.Warn(fmt.Sprintf("writing scene '%s' to cache failed", sceneAlias), zap.Error(err), zap.Any("device", b.Alias))
		}
	}
}

func (b *BaseDevice) DecodeUniverse(sequence string) error {
	size := len(sequence)
	if size % 9 != 0 {
		return fmt.Errorf("got invalid RLE sequence size")
	}

	previousLastChannel := -1
	for i := 0; i < size; i+=9 {
		subsequence := sequence[i:i+9]
		initialChannel, err := strconv.Atoi(subsequence[0:3])
		if err != nil {
			return err
		}
		lastChannel, err := strconv.Atoi(subsequence[3:6])
		if err != nil {
			return err
		}
		channelValue, err := strconv.Atoi(subsequence[6:9])
		if err != nil {
			return err
		}
		if initialChannel > lastChannel {
			return fmt.Errorf("got invalid RLE sequence (initialChannel > lastChannel)")
		}

		if initialChannel < 0 || initialChannel > 511 {
			return fmt.Errorf("got invalid RLE sequence (initialChannel out of range [0:511])")
		}

		if lastChannel < 0 || lastChannel > 511 {
			return fmt.Errorf("got invalid RLE sequence (lastChannel out of range [0:511])")
		}

		if channelValue < 0 || channelValue > 511 {
			return fmt.Errorf("got invalid RLE sequence (channelValue out of range [0:511])")
		}

		if previousLastChannel >= initialChannel {
			return fmt.Errorf("got invalid RLE sequence (previousLastChannel > currentInitialChannel)")
		}

		for j := initialChannel; j <= lastChannel; j++ {
			b.Universe[j] = byte(channelValue)
		} 

		previousLastChannel = lastChannel
	}

	return nil
}

func (b *BaseDevice) EncodeUniverse() string {
	var result string
	var currentChannelValue int = int(b.Universe[0])
	var sequenceStart int = 0
	for idx, channel := range b.Universe {
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

func (b *BaseDevice) EncodeScene(scene Scene) string {
	var result string

	for sceneChannelID, channel := range scene.ChannelMap {
		result += fmt.Sprintf("%03d", sceneChannelID) + fmt.Sprintf("%03d", channel.UniverseChannelID) + fmt.Sprintf("%03d", channel.Value)
	}

	return result
}

func (b *BaseDevice) DecodeScene(sequence string, scene Scene) error {
	size := len(sequence)
	if size % 9 != 0 {
		return fmt.Errorf("got invalid RLE sequence size")
	}

	for i := 0; i < size; i+=9 {
		sceneChannelID, err := strconv.Atoi(sequence[i:i+3])
		if err != nil {
			return err
		}

		UniverseChannelID, err := strconv.Atoi(sequence[i+3:i+6])
		if err != nil {
			return err
		}

		channelValue, err := strconv.Atoi(sequence[i+6:i+9])
		if err != nil {
			return err
		}

		if sceneChannelID < 0 || sceneChannelID > 511 {
			return fmt.Errorf("scene channel out of range [0:511]")
		}

		if UniverseChannelID < 0 || UniverseChannelID > 511 {
			return fmt.Errorf("universe channel out of range [0:511]")
		}

		if channelValue < 0 || channelValue > 511 {
			return fmt.Errorf("channel value out of range [0:511]")
		}

		scene.ChannelMap[sceneChannelID] = Channel{UniverseChannelID: UniverseChannelID, Value: channelValue}
	}

	return nil
}