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


func NewArtNetDevice(ctx context.Context, signals chan core.Signal, conf device.ArtNetConfig, logger *zap.Logger) (device.Device, error) {
	newArtNet := &artnetDevice{
		BaseDevice: *device.NewBaseDevice(ctx, conf.Alias, conf.NonBlackoutChannels, conf.Scenes, conf.ReconnectInterval, signals, logger),
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


func (d *artnetDevice) checkHealth() {
	_, ok := d.dev.OutputAddress[artnet.Address{Net: d.net, SubUni: d.subUni}]
	if ok {
		return
	}

	d.Connected.CompareAndSwap(true, false)
}


func (d *artnetDevice) connect() {
	_, ok := d.dev.OutputAddress[artnet.Address{Net: d.net, SubUni: d.subUni}]
	if !ok {
		d.Logger.Warn("Unable to connect ArtNet device", zap.Any("net", d.net), zap.Any("subuni", d.subUni))
		return
	}

	d.Logger.Info("Connected ArtNet device",  zap.Any("net", d.net), zap.Any("subuni", d.subUni))
	d.Connected.CompareAndSwap(false, true)

	d.WriteUniverseToDevice()
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
	err := d.BaseDevice.WriteUniverseToDevice()
	if err != nil {
		return err
	}

	d.dev.SendDMXToAddress(d.Universe, artnet.Address{Net: d.net, SubUni: d.subUni})
	return nil
}

func (d *artnetDevice) WriteValueToChannel(command models.SetChannel) error {
	err := d.BaseDevice.WriteValueToChannel(command)
	if err != nil {
		return err
	}

	if command.Channel < 1 || command.Channel >= 512 {
		return fmt.Errorf("channel number should be beetwen 1 and 511, but got: %v", command.Channel)
	}
	d.dev.SendDMXToAddress(d.Universe, artnet.Address{Net: d.net, SubUni: d.subUni})

	return nil
}

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