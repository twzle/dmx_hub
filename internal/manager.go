package internal

import (
	"context"
	"fmt"
	"git.miem.hse.ru/hubman/hubman-lib/core"

	"git.miem.hse.ru/hubman/dmx-executor/internal/artnet"
	"git.miem.hse.ru/hubman/dmx-executor/internal/config"
	"git.miem.hse.ru/hubman/dmx-executor/internal/device"
	"git.miem.hse.ru/hubman/dmx-executor/internal/dmx"
	"git.miem.hse.ru/hubman/dmx-executor/internal/models"
	"go.uber.org/zap"
)

func NewManager(logger *zap.Logger) *manager {
	return &manager{
		devices: make(map[string]device.Device),
		signals: make(chan core.Signal),
		logger:  logger,
	}
}

type manager struct {
	devices map[string]device.Device
	signals chan core.Signal
	logger  *zap.Logger
}

func (m *manager) GetSignals() chan core.Signal{
	return m.signals
}

func (m *manager) UpdateDevices(ctx context.Context, userConfig config.UserConfig) {
	dmxConfigs := userConfig.DMXDevices
	artnetDeviceConfig := userConfig.ArtNetDevices
	dmxSet := make(map[string]bool, len(dmxConfigs))
	artnetDeviceSet := make(map[string]bool, len(artnetDeviceConfig))

	m.UpdateDMXDevices(ctx, dmxSet, dmxConfigs)
	m.UpdateArtNetDevices(ctx, artnetDeviceSet, artnetDeviceConfig)
}

func (m *manager) UpdateArtNetDevices(ctx context.Context, artnetDeviceSet map[string]bool, artnetDeviceConfig []config.ArtNetConfig) {
	for _, dev := range artnetDeviceConfig {
		artnetDeviceSet[dev.Alias] = true
	}

	// Deleting the devices that have not been received
	for alias := range m.devices {
		if !artnetDeviceSet[alias] {
			err := m.removeDevice(ctx, alias)
			if err != nil {
				m.logger.Error("error with update devices", zap.Error(err), zap.Any("alias", alias))
			}
		}
	}

	for _, conf := range artnetDeviceConfig {
		if _, exist := m.devices[conf.Alias]; !exist {
			err := m.addArtNet(ctx, conf)
			if err != nil {
				m.logger.Error("error with update devices", zap.Error(err), zap.Any("conf", conf))
			}
		}
	}
}

func (m *manager) UpdateDMXDevices(ctx context.Context, dmxDeviceSet map[string]bool, dmxDeviceConfig []config.DMXConfig) {
	for _, dev := range dmxDeviceConfig {
		dmxDeviceSet[dev.Alias] = true
	}

	// Deleting the devices that have not been received
	for alias := range m.devices {
		if !dmxDeviceSet[alias] {
			err := m.removeDevice(ctx, alias)
			if err != nil {
				m.logger.Error("error with update devices", zap.Error(err), zap.Any("alias", alias))
			}
		}
	}

	for _, conf := range dmxDeviceConfig {
		if _, exist := m.devices[conf.Alias]; !exist {
			err := m.addDMX(ctx, conf)
			if err != nil {
				m.logger.Error("error with update devices", zap.Error(err), zap.Any("conf", conf))
			}
		}
	}
}

func (m *manager) ProcessSetChannel(ctx context.Context, command models.SetChannel) error {
	dev, err := m.checkDevice(command.DeviceAlias)
	if err != nil {
		return err
	}

	err = dev.SetChannel(ctx, command)
	if err != nil {
		return fmt.Errorf("device with alias %v setting value error: %v", dev.GetAlias(), err)
	}
	return nil
}

func (m *manager) ProcessBlackout(ctx context.Context, command models.Blackout) error {
	dev, err := m.checkDevice(command.DeviceAlias)
	if err != nil {
		return err
	}

	err = dev.Blackout(ctx)
	if err != nil {
		return fmt.Errorf("device with alias %v blackout error: %v", dev.GetAlias(), err)
	}
	return nil
}

func (m *manager) checkDevice(deviceAlias string) (device.Device, error) {
	dev, devExist := m.devices[deviceAlias]
	if !devExist {
		return nil, fmt.Errorf("dmx-device with alias %v not found", deviceAlias)
	}
	return dev, nil

}

func (m *manager) addDMX(ctx context.Context, conf config.DMXConfig) error {
	newDMX, err := dmx.NewDMXDevice(ctx, m.signals, conf)
	if err != nil {
		return fmt.Errorf("error with add device: %v", err)
	}
	m.devices[newDMX.GetAlias()] = newDMX
	return nil
}

func (m *manager) addArtNet(ctx context.Context, conf config.ArtNetConfig) error {
	newArtNet, err := artnet.NewArtNetDevice(ctx, m.signals, conf)
	if err != nil {
		return fmt.Errorf("error with add device: %v", err)
	}
	m.devices[newArtNet.GetAlias()] = newArtNet
	return nil
}

func (m *manager) removeDevice(_ context.Context, alias string) error {
	delete(m.devices, alias)
	return nil
}
