package internal

import (
	"context"
	"fmt"
	"log"
	"net"

	"git.miem.hse.ru/hubman/dmx-executor/config"
	"github.com/jsimonetti/go-artnet"
)

type ArtNetDevice interface {
	GetAlias() string
	SetValueToChannel(ctx context.Context, channel uint16, value uint16) error
	Blackout(ctx context.Context) error
}

func NewArtNetDevice(ctx context.Context, conf config.ArtNetConfig) (ArtNetDevice, error) {
	ip := net.ParseIP(conf.IP)
	if ip == nil {
		return nil, fmt.Errorf("invalid ip address for artnet device: %v", conf.IP)
	}

	log := artnet.NewDefaultLogger()
	dev := artnet.NewController(conf.Alias, ip, log)
	dev.Start()

	newArtNet := &artnetDevice{alias: conf.Alias, dev: dev}
	newArtNet.ReadUniverse(conf.Universe)

	return newArtNet, nil
}

type artnetDevice struct {
	alias    string
	universe [512]byte
	dev      *artnet.Controller
}

func (d *artnetDevice) GetAlias() string {
	return d.alias
}

func (d *artnetDevice) ReadUniverse(universe []config.ChannelRange) string {
	artNetUniverse := make(map[uint16]uint16, len(universe))
	for _, channelRange := range universe {
		artNetUniverse[channelRange.InitialIndex] = channelRange.Value
	}

	var channelValue uint16
	for idx, channel := range d.universe {
		value, ok := artNetUniverse[uint16(idx)] 
		if ok {
			channelValue = value
			log.Println(value)
		}

		channel = byte(channelValue)
		d.universe[idx] = channel
	}

	for _, channel := range d.universe {
		fmt.Printf("%d ", channel)
	}

	return d.alias
}

func (d *artnetDevice) SetValueToChannel(ctx context.Context, channel uint16, value uint16) error {
	if channel < 1 || channel >= 512 {
		return fmt.Errorf("channel number should be beetwen 1 and 511, but got: %v", channel)
	}
	d.universe[channel] = byte(value)
	d.dev.SendDMXToAddress(d.universe, artnet.Address{Net: 0, SubUni: 0})

	for _, channel := range d.universe {
		fmt.Printf("%d ", channel)
	}

	return nil
}

func (d *artnetDevice) Blackout(ctx context.Context) error {
	d.universe = [512]byte{}
	d.dev.SendDMXToAddress(d.universe, artnet.Address{Net: 0, SubUni: 0})

	return nil
}
