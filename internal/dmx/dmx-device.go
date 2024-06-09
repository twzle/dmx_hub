package dmx

import (
	"context"
	"fmt"
	"time"

	"git.miem.hse.ru/hubman/hubman-lib/core"
	"go.uber.org/zap"

	"git.miem.hse.ru/hubman/dmx-executor/internal/device"
	"git.miem.hse.ru/hubman/dmx-executor/internal/models"
)

// Function initializes and returns DMX device entity
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
	dev  *DMX
}

// Function reconnects single DMX device
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

// Function connects to the DMX device through OS
func (d *dmxDevice) connect() {
	dev, err := NewDMXConnection(d.path)
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

// Function checks availability of single DMX device
func (d *dmxDevice) checkHealth() {
	d.Mutex.Lock()
	defer d.Mutex.Unlock()

	err := d.dev.Render()
	if err != nil {
		d.dev.Close()
		d.Connected.Store(false)
	}
}

// Function writes universe to single DMX device
func (d *dmxDevice) WriteUniverseToDevice() error {
	err := d.BaseDevice.WriteUniverseToDevice()
	if err != nil {
		return err
	}

	d.Mutex.Lock()
	defer d.Mutex.Unlock()

	for i := 0; i < 512; i++ {
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

// Function sets scene specified in command and updates universe of single DMX device
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

// Function sets and writes value to channel of universe specified in command for single DMX device
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

// Function writes incremented value to channel of universe specified in command for single DMX device
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

// Function writes value to channel of universe specified in command for single DMX device
func (d *dmxDevice) WriteValueToChannel(command models.SetChannel) error {
	err := d.BaseDevice.WriteValueToChannel(command)
	if err != nil {
		return err
	}

	err = d.dev.SetChannel(command.Channel, byte(command.Value))
	if err != nil {
		return err
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

// Function handles blackout for whole DMX universe of single DMX device
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

// Function frees resources of DMX device entity
func (d *dmxDevice) Close(){
	d.BaseDevice.Close()
	if d.Connected.Load() {
		d.dev.Close()
	}
}