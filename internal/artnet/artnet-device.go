package artnet

import (
	"context"
	"fmt"

	"git.miem.hse.ru/hubman/dmx-executor/internal/config"
	"git.miem.hse.ru/hubman/dmx-executor/internal/device"
	"git.miem.hse.ru/hubman/dmx-executor/internal/models"
	"git.miem.hse.ru/hubman/hubman-lib/core"
	"github.com/jsimonetti/go-artnet"
)

func NewArtNetDevice(ctx context.Context, signals chan core.Signal, conf config.ArtNetConfig) (device.Device, error) {
	log := artnet.NewDefaultLogger()
	dev := artnet.NewController(conf.Alias, conf.IP, log)
	dev.Start()

	newArtNet := &artnetDevice{alias: conf.Alias, dev: dev}
	newArtNet.SetUniverse(conf.Universe)
	newArtNet.scenes = device.ReadScenesFromDeviceConfig(conf.Scenes)
	newArtNet.currentScene = device.GetSceneById(newArtNet.scenes, 0)

	return newArtNet, nil
}

type artnetDevice struct {
	alias        string
	dev          *artnet.Controller
	universe     [512]byte
	scenes       map[string]device.Scene
	currentScene *device.Scene
	signals      chan core.Signal
}

func (d *artnetDevice) GetAlias() string {
	return d.alias
}

func (d *artnetDevice) SetUniverse(universe []config.ChannelRange) {
	artNetUniverse := make(map[uint16]uint16, len(universe))
	for _, channelRange := range universe {
		artNetUniverse[channelRange.InitialIndex] = channelRange.Value
	}

	var channelValue uint16
	for idx, channel := range d.universe {
		value, ok := artNetUniverse[uint16(idx)]
		if ok {
			channelValue = value
		}

		channel = byte(channelValue)
		d.universe[idx] = channel
	}
}


func (d *artnetDevice) SetScene(sceneAlias string) error {
	scene, ok := d.scenes[sceneAlias]
	if !ok {
		return fmt.Errorf("invalid scene alias '%s' for device '%s'", sceneAlias, d.alias)
	}
	d.currentScene = &scene

	signal := models.SceneChanged{DeviceAlias: d.alias, SceneAlias: d.currentScene.Alias}
	d.signals <- signal
	return nil
}

func (d *artnetDevice) SetChannel(ctx context.Context, command models.SetChannel) error {
	if d.currentScene == nil {
		return fmt.Errorf("no scene is selected")
	}
	_, ok := d.currentScene.ChannelMap[command.Channel]
	if !ok {
		return fmt.Errorf("channel '%d' doesn't belong to current scene '%s'", command.Channel, d.currentScene.Alias)
	}
	d.universe[command.Channel] = byte(command.Value)
	err := d.WriteValueToChannel(ctx, command)
	if err != nil {
		return err
	}
	return nil
}


func (d *artnetDevice) WriteValueToChannel(ctx context.Context, command models.SetChannel) error {
	if command.Channel < 1 || command.Channel >= 512 {
		return fmt.Errorf("channel number should be beetwen 1 and 511, but got: %v", command.Channel)
	}
	d.dev.SendDMXToAddress(d.universe, artnet.Address{Net: 0, SubUni: 0})

	return nil
}

func (d *artnetDevice) Blackout(ctx context.Context) error {
	d.universe = [512]byte{}
	d.dev.SendDMXToAddress(d.universe, artnet.Address{Net: 0, SubUni: 0})

	return nil
}
