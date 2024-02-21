package internal

import (
	"context"
	"fmt"
	"git.miem.hse.ru/hubman/dmx-executor/config"
	"github.com/akualab/dmx"
)

type DMXDevice interface {
	GetAlias() string
	SetValueToChannel(ctx context.Context, command SetChannel) error
	Blackout(ctx context.Context) error
}

func NewDMXDevice(ctx context.Context, conf config.DMXConfig) (DMXDevice, error) {
	dev, err := dmx.NewDMXConnection(conf.Path)
	if err != nil {
		return nil, fmt.Errorf("error with creating new dmx dmxDevice: %v", err)
	}

	newDMX := &dmxDevice{alias: conf.Alias, dev: dev}

	if err != nil {
		return nil, fmt.Errorf("error getting dmxDevice with alias %v profile token: %v", conf.Alias, err)
	}

	return newDMX, nil
}

type dmxDevice struct {
	alias string
	dev   *dmx.DMX
}

func (d *dmxDevice) GetAlias() string {
	return d.alias
}

func (d *dmxDevice) SetValueToChannel(ctx context.Context, command SetChannel) error {
	if command.Channel < 1 || command.Channel >= 512 {
		return fmt.Errorf("channel number should be beetwen 1 and 511, but got: %v", command.Channel)
	}
	sig := byte(command.Value)
	err := d.dev.SetChannel(command.Channel, sig)
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
	d.dev.ClearAll()
	err := d.dev.Render()
	if err != nil {
		return fmt.Errorf("sending frame to dev error: %v", err)
	}

	return nil
}
