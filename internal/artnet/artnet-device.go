package artnet

import (
	"context"
	"fmt"

	"git.miem.hse.ru/hubman/dmx-executor/internal/device"
	"git.miem.hse.ru/hubman/dmx-executor/internal/models"
	"git.miem.hse.ru/hubman/hubman-lib/core"
	"github.com/jsimonetti/go-artnet"

	"go.uber.org/zap"
)

func NewArtNetDevice(ctx context.Context, signals chan core.Signal, conf device.ArtNetConfig, logger *zap.Logger) (device.Device, error) {
	log := artnet.NewDefaultLogger()
	dev := artnet.NewController(conf.Alias, conf.IP, log)
	dev.Start()

	newArtNet := &artnetDevice{
		BaseDevice: device.NewBaseDevice(ctx, conf.Alias, conf.Scenes, signals, logger),
		dev:        dev}
		
	newArtNet.WriteUniverseToDevice()
	return newArtNet, nil
}

type artnetDevice struct {
	device.BaseDevice
	dev *artnet.Controller
}

func (d *artnetDevice) SetScene(ctx context.Context, command models.SetScene) error {
	err := d.BaseDevice.SetScene(ctx, command)
	if err != nil {
		return err
	}

	for _, channel := range d.CurrentScene.ChannelMap {
		d.Universe[channel.UniverseChannelID] = byte(channel.Value)
	}

	d.SaveUniverseToCache(ctx)
	d.WriteValueToChannel(models.SetChannel{})
	d.CreateSceneChangedSignal()
	return nil
}

func (d *artnetDevice) SetChannel(ctx context.Context, command models.SetChannel) error {
	err := d.BaseDevice.SetChannel(ctx, &command)
	if err != nil {
		return err
	}
	err = d.WriteValueToChannel(command)
	if err != nil {
		return err
	}
	return nil
}

func (d *artnetDevice) IncrementChannel(ctx context.Context, command models.IncrementChannel) error {
	err := d.BaseDevice.IncrementChannel(ctx, &command)
	if err != nil {
		return err
	}

	cmd := models.SetChannel{
		Channel:     command.Channel,
		Value:       int(d.Universe[command.Channel]),
		DeviceAlias: d.Alias}

	err = d.WriteValueToChannel(cmd)
	if err != nil {
		return err
	}
	return nil
}

func (d *artnetDevice) WriteUniverseToDevice() error {
	d.dev.SendDMXToAddress(d.BaseDevice.Universe, artnet.Address{Net: 0, SubUni: 0})
	return nil
}

func (d *artnetDevice) WriteValueToChannel(command models.SetChannel) error {
	if command.Channel < 1 || command.Channel >= 512 {
		return fmt.Errorf("channel number should be beetwen 1 and 511, but got: %v", command.Channel)
	}
	d.dev.SendDMXToAddress(d.Universe, artnet.Address{Net: 0, SubUni: 0})

	return nil
}

func (d *artnetDevice) Blackout(ctx context.Context) error{
	d.BaseDevice.Blackout(ctx)
	d.dev.SendDMXToAddress(d.Universe, artnet.Address{Net: 0, SubUni: 0})
	return nil
}
