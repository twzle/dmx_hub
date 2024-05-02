package dmx

import (
	"context"
	"fmt"
	"time"

	"git.miem.hse.ru/hubman/hubman-lib/core"
	"go.uber.org/zap"

	"git.miem.hse.ru/hubman/dmx-executor/internal/device"
	"git.miem.hse.ru/hubman/dmx-executor/internal/models"
	DMX "github.com/akualab/dmx"
)

func NewDMXDevice(ctx context.Context, signals chan core.Signal, conf device.DMXConfig, logger *zap.Logger, checkManager core.CheckRegistry) (device.Device, error) {
	newDMX := &dmxDevice{
		BaseDevice: *device.NewBaseDevice(ctx, conf.Alias, conf.NonBlackoutChannels, conf.Scenes, conf.ReconnectInterval, signals, logger, checkManager),
		path:       conf.Path,
		dev:        nil,
	}

	go newDMX.reconnect()
	return newDMX, nil
}

type dmxDevice struct {
	device.BaseDevice
	path string
	dev  *DMX.DMX
}

func (d *dmxDevice) reconnect() {
	ticker := time.NewTicker(d.ReconnectInterval)
	for {
		select {
		case <-d.StopReconnect:
			ticker.Stop()
			return
		case <-ticker.C:
			if !d.Connected.Load() {
				d.connect()
			} else {
				d.checkHealth()
			}
		}
	}
}

func (d *dmxDevice) connect() {
	dev, err := DMX.NewDMXConnection(d.path)
	if err != nil {
		connCheck := core.NewCheck(
			fmt.Sprintf(device.DeviceDisconnectedCheckLabelFormat, d.Alias),
			"",
		)
		d.CheckManager.RegisterFail(connCheck)
		d.Logger.Warn("Unable to connect DMX device", zap.Any("path", d.path), zap.Error(err))
		return
	}

	d.Connected.CompareAndSwap(false, true)

	d.Mutex.Lock()
	d.dev = dev
	d.Mutex.Unlock()
	d.WriteUniverseToDevice()

	connCheck := core.NewCheck(
		fmt.Sprintf(device.DeviceDisconnectedCheckLabelFormat, d.Alias),
		"",
	)
	d.CheckManager.RegisterSuccess(connCheck)
	d.Logger.Info("Connected DMX device", zap.Any("path", d.path))
}

func (d *dmxDevice) checkHealth() {
	d.Mutex.Lock()
	defer d.Mutex.Unlock()

	err := d.dev.Render()
	if err != nil {
		d.dev.Close()
		d.Connected.Store(false)
	}
}

func (d *dmxDevice) WriteUniverseToDevice() error {
	err := d.BaseDevice.WriteUniverseToDevice()
	if err != nil {
		return err
	}

	d.Mutex.Lock()
	defer d.Mutex.Unlock()

	for i := 1; i < 511; i++ {
		err := d.dev.SetChannel(i, d.Universe[i])
		if err != nil {
			return fmt.Errorf("setting value to channel error: %v", err)
		}
	}

	err = d.dev.Render()
	if err != nil {
		d.dev.Close()
		d.Connected.CompareAndSwap(true, false)
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
	err := d.BaseDevice.WriteValueToChannel(command)
	if err != nil {
		return err
	}

	if command.Channel < 1 || command.Channel >= 512 {
		return fmt.Errorf("channel number should be beetwen 1 and 511, but got: %v", command.Channel)
	}
	err = d.dev.SetChannel(command.Channel, byte(command.Value))
	if err != nil {
		return fmt.Errorf("setting value to channel error: %v", err)
	}

	d.Mutex.Lock()
	defer d.Mutex.Unlock()

	err = d.dev.Render()
	if err != nil {
		d.dev.Close()
		d.Connected.CompareAndSwap(true, false)
		return fmt.Errorf("sending frame to device error: %v", err)
	}
	return nil
}

func (d *dmxDevice) Blackout(ctx context.Context) error {
	err := d.BaseDevice.Blackout(ctx)
	if err != nil {
		return err
	}

	err = d.WriteUniverseToDevice()
	if err != nil {
		return err
	}

	return nil
}

func (d *dmxDevice) Close(){
	d.BaseDevice.Close()
	if d.Connected.Load() {
		d.dev.Close()
	}
}