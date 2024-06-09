package artnet

import (
	"context"
	"fmt"
	"time"

	"git.miem.hse.ru/hubman/dmx-executor/internal/device"
	"git.miem.hse.ru/hubman/dmx-executor/internal/models"
	"git.miem.hse.ru/hubman/hubman-lib/core"
	"github.com/jsimonetti/go-artnet"

	"go.uber.org/zap"
)

// Function initializes and returns Artnet device entity
func NewArtNetDevice(ctx context.Context, signals chan core.Signal, conf device.ArtNetConfig, logger *zap.Logger, checkManager core.CheckRegistry) (device.Device, error) {
	newArtNet := &artnetDevice{
		BaseDevice: *device.NewBaseDevice(ctx, conf.Alias, conf.NonBlackoutChannels, conf.Scenes, conf.ReconnectInterval, signals, logger, checkManager),
		net:        uint8(conf.Net),
		subUni:     uint8(conf.SubUni),
		dev:        GetArtNetController()}

	go newArtNet.reconnect()
	return newArtNet, nil
}

type artnetDevice struct {
	device.BaseDevice
	net    uint8
	subUni uint8
	dev    *artnet.Controller
}

// Function reconnects single Artnet device
func (d *artnetDevice) reconnect() {
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

// Function checks availability of single Artnet device
func (d *artnetDevice) checkHealth() {
	_, ok := d.dev.OutputAddress[artnet.Address{Net: d.net, SubUni: d.subUni}]
	if ok {
		return
	}

	d.Connected.CompareAndSwap(true, false)
}

// Function connects to the Artnet device through network
func (d *artnetDevice) connect() {
	_, ok := d.dev.OutputAddress[artnet.Address{Net: d.net, SubUni: d.subUni}]
	if !ok {
		connCheck := core.NewCheck(
			fmt.Sprintf(device.DeviceDisconnectedCheckLabelFormat, d.Alias),
			"",
		)
		d.CheckManager.RegisterFail(connCheck)
		d.Logger.Warn("Unable to connect ArtNet device", zap.Any("net", d.net), zap.Any("subuni", d.subUni))
		return
	}

	d.Connected.CompareAndSwap(false, true)
	d.WriteUniverseToDevice()
	
	connCheck := core.NewCheck(
		fmt.Sprintf(device.DeviceDisconnectedCheckLabelFormat, d.Alias),
		"",
	)
	d.CheckManager.RegisterSuccess(connCheck)
	d.Logger.Info("Connected ArtNet device",  zap.Any("net", d.net), zap.Any("subuni", d.subUni))
}

// Function sets scene specified in command and updates universe of single Artnet device
func (d *artnetDevice) SetScene(ctx context.Context, command models.SetScene) error {
	err := d.BaseDevice.SetScene(ctx, command)
	if err != nil {
		return err
	}

	for _, channel := range d.CurrentScene.ChannelMap {
		d.Universe[channel.UniverseChannelID] = byte(channel.Value)
	}

	d.SaveUniverseToCache(ctx)
	d.WriteUniverseToDevice()
	d.CreateSceneChangedSignal()
	return nil
}

// Function sets and writes value to channel of universe specified in command for single Artnet device
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

// Function writes incremented value to channel of universe specified in command for single Artnet device
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

// Function writes universe to single Artnet device
func (d *artnetDevice) WriteUniverseToDevice() error {
	err := d.BaseDevice.WriteUniverseToDevice()
	if err != nil {
		return err
	}

	d.dev.SendDMXToAddress(d.Universe, artnet.Address{Net: d.net, SubUni: d.subUni})
	return nil
}

// Function writes value to channel of universe specified in command for single Artnet device
func (d *artnetDevice) WriteValueToChannel(command models.SetChannel) error {
	err := d.BaseDevice.WriteValueToChannel(command)
	if err != nil {
		return err
	}

	if command.Channel < 0 || command.Channel >= 511 {
		return fmt.Errorf("channel number should be beetwen 0 and 511, but got: %v", command.Channel)
	}
	d.dev.SendDMXToAddress(d.Universe, artnet.Address{Net: d.net, SubUni: d.subUni})

	return nil
}

// Function handles blackout for whole DMX universe of single Artnet device
func (d *artnetDevice) Blackout(ctx context.Context) error {
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