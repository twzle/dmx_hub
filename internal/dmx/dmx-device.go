package dmx

import (
	"context"
	"fmt"
	"git.miem.hse.ru/hubman/hubman-lib/core"

	"git.miem.hse.ru/hubman/dmx-executor/internal/config"
	"git.miem.hse.ru/hubman/dmx-executor/internal/device"
	"git.miem.hse.ru/hubman/dmx-executor/internal/models"
	DMX "github.com/akualab/dmx"
)

func NewDMXDevice(ctx context.Context, signals chan core.Signal, conf config.DMXConfig) (device.Device, error) {
	dev, err := DMX.NewDMXConnection(conf.Path)
	if err != nil {
		return nil, fmt.Errorf("error with creating new dmx dmxDevice: %v", err)
	}

	newDMX := &dmxDevice{alias: conf.Alias, dev: dev, signals: signals}
	newDMX.SetUniverse(ctx, conf.Universe)
	newDMX.scenes = device.ReadScenesFromDeviceConfig(conf.Scenes)
	newDMX.signals = signals
	newDMX.currentScene = device.GetSceneById(newDMX.scenes, 0)

	if err != nil {
		return nil, fmt.Errorf("error getting dmxDevice with alias %v profile token: %v", conf.Alias, err)
	}

	return newDMX, nil
}

type dmxDevice struct {
	alias        string
	dev          *DMX.DMX
	universe     [512]byte
	scenes       map[string]device.Scene
	currentScene *device.Scene
	signals      chan core.Signal
}

func (d *dmxDevice) GetAlias() string {
	return d.alias
}

func (d *dmxDevice) SetUniverse(ctx context.Context, universe []config.ChannelRange) {
	dmxUniverse := make(map[uint16]uint16, len(universe))
	for _, channelRange := range universe {
		dmxUniverse[channelRange.InitialIndex] = channelRange.Value
	}

	var channelValue uint16
	for idx, channel := range d.universe {
		value, ok := dmxUniverse[uint16(idx)]
		if ok {
			channelValue = value
		}

		channel = byte(channelValue)
		d.universe[idx] = channel
	}
}

func (d *dmxDevice) SetScene(ctx context.Context, sceneAlias string) error {
	scene, ok := d.scenes[sceneAlias]
	if !ok {
		return fmt.Errorf("invalid scene alias '%s' for device '%s'", sceneAlias, d.alias)
	}
	d.currentScene = &scene

	for _, channel := range d.currentScene.ChannelMap {
		d.universe[channel.UniverseChannelID] = byte(channel.Value)
		d.WriteValueToChannel(ctx, models.SetChannel{Channel: channel.UniverseChannelID, Value: channel.Value, DeviceAlias: d.alias})
	}

	signal := models.SceneChanged{DeviceAlias: d.alias, SceneAlias: d.currentScene.Alias}
	d.signals <- signal
	return nil
}

func (d *dmxDevice) SaveScene(ctx context.Context) {
	for sceneChannelID, channel := range d.currentScene.ChannelMap {
		channel.Value = int(d.universe[channel.UniverseChannelID])
		d.currentScene.ChannelMap[sceneChannelID] = channel
	}
	
	signal := models.SceneSaved{DeviceAlias: d.alias, SceneAlias: d.currentScene.Alias}
	d.signals <- signal
}

func (d *dmxDevice) SetChannel(ctx context.Context, command models.SetChannel) error {
	if d.currentScene == nil {
		return fmt.Errorf("no scene is selected")
	}
	channel, ok := d.currentScene.ChannelMap[command.Channel]
	if !ok {
		return fmt.Errorf("channel '%d' doesn't belong to current scene '%s'", command.Channel, d.currentScene.Alias)
	}
	command.Channel = channel.UniverseChannelID
	d.universe[command.Channel] = byte(command.Value)
	err := d.WriteValueToChannel(ctx, command)
	if err != nil {
		return err
	}
	return nil
}

func (d *dmxDevice) WriteValueToChannel(ctx context.Context, command models.SetChannel) error {
	if command.Channel < 1 || command.Channel >= 512 {
		return fmt.Errorf("channel number should be beetwen 1 and 511, but got: %v", command.Channel)
	}
	err := d.dev.SetChannel(command.Channel, byte(command.Value))
	if err != nil {
		return fmt.Errorf("setting value to channel error: %v", err)
	}
	err = d.dev.Render()
	if err != nil {
		return fmt.Errorf("sending frame to dev error: %v", err)
	}
	return nil
}

func (d *dmxDevice) Blackout(ctx context.Context) error {
	d.universe = [512]byte{}
	d.dev.ClearAll()
	err := d.dev.Render()
	if err != nil {
		return fmt.Errorf("sending frame to dev error: %v", err)
	}

	return nil
}
