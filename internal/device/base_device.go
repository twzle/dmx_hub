package device

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"git.miem.hse.ru/hubman/dmx-executor/internal/models"
	"git.miem.hse.ru/hubman/hubman-lib/core"
	"go.uber.org/zap"
)

type BaseDevice struct {
	Alias               string
	Universe            [512]byte
	NonBlackoutChannels map[int]struct{}
	Scenes              map[string]Scene
	CurrentScene        *Scene
	Signals             chan core.Signal
	Logger              *zap.Logger
	Connected           atomic.Bool
	ReconnectInterval   time.Duration
	StopReconnect       chan struct{}
	Mutex               sync.Mutex
}

func NewBaseDevice(ctx context.Context, alias string, nonBlackoutChannels []int, scenes []SceneConfig, reconnectInteval int, signals chan core.Signal, logger *zap.Logger) *BaseDevice {
	device := BaseDevice{
		Alias:               alias,
		Universe:            [512]byte{},
		NonBlackoutChannels: make(map[int]struct{}),
		Scenes:              make(map[string]Scene),
		CurrentScene:        nil,
		Signals:             signals,
		Logger:              logger.With(zap.String("device", alias)),
		Connected:           atomic.Bool{},
		ReconnectInterval:   time.Duration(time.Millisecond * time.Duration(reconnectInteval)),
		StopReconnect:       make(chan struct{}),
		Mutex:               sync.Mutex{},
	}

	device.NonBlackoutChannels = ReadNonBlackoutChannelsFromDeviceConfig(nonBlackoutChannels)
	device.Scenes = ReadScenesFromDeviceConfig(scenes)
	device.GetUniverseFromCache(ctx)
	device.GetScenesFromCache(ctx)
	return &device
}

func (b *BaseDevice) GetUniverseFromCache(ctx context.Context) {
	err := b.ReadUnvierse(ctx)
	if err != nil {
		b.Logger.Warn("get universe from cache failed", zap.Error(err))
	}
}

func (b *BaseDevice) SaveUniverseToCache(ctx context.Context) {
	err := b.WriteUniverse(ctx)
	if err != nil {
		b.Logger.Warn("save universe to cache failed", zap.Error(err))
	}
}

func (b *BaseDevice) GetScenesFromCache(ctx context.Context) {
	b.ReadScenes(ctx)
}

func (b *BaseDevice) SaveScenesToCache(ctx context.Context) {
	b.WriteScenes(ctx)
}

func (b *BaseDevice) GetAlias() string {
	return b.Alias
}

func (b *BaseDevice) SetScene(ctx context.Context, command models.SetScene) error {
	if !b.Connected.Load() {
		return fmt.Errorf("no connection to device")
	}

	scene, ok := b.Scenes[command.SceneAlias]
	if !ok {
		return fmt.Errorf("invalid scene alias '%s'", command.SceneAlias)
	}
	b.CurrentScene = &scene
	return nil
}

func (b *BaseDevice) SaveScene(ctx context.Context) error {
	if b.CurrentScene == nil {
		return fmt.Errorf("no scene is selected")
	}

	for sceneChannelID, channel := range b.CurrentScene.ChannelMap {
		channel.Value = int(b.Universe[channel.UniverseChannelID])
		b.CurrentScene.ChannelMap[sceneChannelID] = channel
	}

	b.SaveScenesToCache(ctx)
	b.CreateSceneSavedSignal()
	return nil
}

func (b *BaseDevice) SetChannel(ctx context.Context, command *models.SetChannel) error {
	if !b.Connected.Load() {
		return fmt.Errorf("no connection to device")
	}

	if b.CurrentScene == nil {
		return fmt.Errorf("no scene is selected")
	}

	channel, ok := b.CurrentScene.ChannelMap[command.Channel]
	if !ok {
		return fmt.Errorf("channel '%d' doesn't belong to current scene '%s'", command.Channel, b.CurrentScene.Alias)
	}
	command.Channel = channel.UniverseChannelID
	b.Universe[command.Channel] = byte(command.Value)
	b.SaveUniverseToCache(ctx)
	return nil
}

func (b *BaseDevice) IncrementChannel(ctx context.Context, command *models.IncrementChannel) error {
	if !b.Connected.Load() {
		return fmt.Errorf("no connection to device")
	}

	if b.CurrentScene == nil {
		return fmt.Errorf("no scene is selected")
	}

	channel, ok := b.CurrentScene.ChannelMap[command.Channel]
	if !ok {
		return fmt.Errorf("channel '%d' doesn't belong to current scene '%s'", command.Channel, b.CurrentScene.Alias)
	}

	command.Channel = channel.UniverseChannelID
	command.Value = int(b.Universe[command.Channel]) + command.Value
	if (int(b.Universe[command.Channel])+command.Value) < 0 || (int(b.Universe[command.Channel])+command.Value) > 255 {
		return fmt.Errorf("incremented channel value '%d' out of range [0, 255]", command.Value)
	}

	b.Universe[command.Channel] = byte(command.Value)
	b.SaveUniverseToCache(ctx)
	return nil
}

func (b *BaseDevice) WriteValueToChannel(command models.SetChannel) error {
	if !b.Connected.Load() {
		return fmt.Errorf("no connection to device")
	}
	return nil
}

func (b *BaseDevice) WriteUniverseToDevice() error {
	if !b.Connected.Load() {
		return fmt.Errorf("no connection to device")
	}
	return nil
}

func (b *BaseDevice) Blackout(ctx context.Context) error {
	if !b.Connected.Load() {
		return fmt.Errorf("no connection to device")
	}

	for i := 0; i < 512; i++ {
		_, ok := b.NonBlackoutChannels[i]
		if !ok {
			b.Universe[i] = 0
		}
	}
	b.SaveUniverseToCache(ctx)
	return nil
}

func (b *BaseDevice) CreateSceneChangedSignal() {
	signal := models.SceneChanged{
		DeviceAlias: b.Alias,
		SceneAlias:  b.CurrentScene.Alias}
	b.Signals <- signal
}

func (b *BaseDevice) CreateSceneSavedSignal() {
	signal := models.SceneSaved{
		DeviceAlias: b.Alias,
		SceneAlias:  b.CurrentScene.Alias}
	b.Signals <- signal
}

func (b *BaseDevice) Close() {
	b.StopReconnect <- struct{}{}
	close(b.StopReconnect)
}
