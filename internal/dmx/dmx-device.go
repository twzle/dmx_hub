package dmx

import (
	"context"
	"fmt"

	"git.miem.hse.ru/hubman/hubman-lib/core"
	"go.uber.org/zap"

	"git.miem.hse.ru/hubman/dmx-executor/internal/device"
	"git.miem.hse.ru/hubman/dmx-executor/internal/models"
	DMX "github.com/akualab/dmx"
)

func NewDMXDevice(ctx context.Context, signals chan core.Signal, conf device.DMXConfig, logger *zap.Logger) (device.Device, error) {
	dev, err := DMX.NewDMXConnection(conf.Path)
	if err != nil {
		return nil, fmt.Errorf("error with creating new dmx dmxDevice: %v", err)
	}

	newDMX := &dmxDevice{
		BaseDevice: device.NewBaseDevice(ctx, conf.Alias, conf.Scenes, signals, logger),
		dev:        dev}

	newDMX.WriteUniverseToDevice()
	return newDMX, nil
}

type dmxDevice struct {
	device.BaseDevice
	dev     *DMX.DMX
}

func (d *dmxDevice) WriteUniverseToDevice() error {
	for i := 0; i < 510; i++ {
		err := d.dev.SetChannel(i+1, d.Universe[i])
		if err != nil {
			return fmt.Errorf("setting value to channel error: %v", err)
		}
	}

	err := d.dev.Render()
	if err != nil {
		return fmt.Errorf("sending frame to device error: %v", err)
	}
	return nil
}

func (d *dmxDevice) SetScene(ctx context.Context, command models.SetScene) error {
	err := d.BaseDevice.SetScene(ctx, command)
	if err != nil {
		return err
	}

	for _, channel := range d.CurrentScene.ChannelMap {
		d.Universe[channel.UniverseChannelID] = byte(channel.Value)
		d.WriteValueToChannel(
			models.SetChannel{
				Channel:     channel.UniverseChannelID,
				Value:       channel.Value,
				DeviceAlias: d.Alias})
	}
	d.SaveUniverseToCache(ctx)
	d.CreateSceneChangedSignal()
	return nil
}

func (d *dmxDevice) SetChannel(ctx context.Context, command models.SetChannel) error {
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

func (d *dmxDevice) IncrementChannel(ctx context.Context, command models.IncrementChannel) error {
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

func (d *dmxDevice) WriteValueToChannel(command models.SetChannel) error {
	if command.Channel < 1 || command.Channel >= 512 {
		return fmt.Errorf("channel number should be beetwen 1 and 511, but got: %v", command.Channel)
	}
	err := d.dev.SetChannel(command.Channel, byte(command.Value))
	if err != nil {
		return fmt.Errorf("setting value to channel error: %v", err)
	}
	err = d.dev.Render()
	if err != nil {
		return fmt.Errorf("sending frame to device error: %v", err)
	}
	return nil
}

func (d *dmxDevice) Blackout(ctx context.Context) error {
	d.BaseDevice.Blackout(ctx)
	d.dev.ClearAll()
	err := d.dev.Render()
	if err != nil {
		return fmt.Errorf("sending frame to device error: %v", err)
	}
	return nil
}
