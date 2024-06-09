package internal

import (
	"context"
	"fmt"

	"git.miem.hse.ru/hubman/hubman-lib/core"

	"git.miem.hse.ru/hubman/dmx-executor/internal/artnet"
	"git.miem.hse.ru/hubman/dmx-executor/internal/device"
	"git.miem.hse.ru/hubman/dmx-executor/internal/dmx"
	"git.miem.hse.ru/hubman/dmx-executor/internal/models"
	"go.uber.org/zap"
)

// Function initializes device manager entity
func NewManager(logger *zap.Logger, checkManager core.CheckRegistry) *manager {
	return &manager{
		devices: make(map[string]device.Device),
		signals: make(chan core.Signal),
		logger:  logger,
		checkManager: checkManager,
	}
}

// Representation of device manager entity
type manager struct {
	devices      map[string]device.Device
	signals      chan core.Signal
	logger       *zap.Logger
	checkManager core.CheckRegistry
}

// Function returns signal channel value of device manager
func (m *manager) GetSignals() chan core.Signal {
	return m.signals
}

// Function returns current device list of device manager
func (m *manager) GetDevices() map[string]device.Device {
	return m.devices
}

// Function updates current device list of device manager
func (m *manager) UpdateDevices(ctx context.Context, userConfig device.UserConfig) {
	dmxDeviceConfig := userConfig.DMXDevices
	artnetDeviceConfig := userConfig.ArtNetDevices

	m.checkManager.Clear()

	for alias := range m.devices {
		err := m.removeDevice(ctx, alias)
		if err != nil {
			m.logger.Error("error while removing device", zap.Error(err), zap.Any("alias", alias))
		}
	}

	for _, conf := range artnetDeviceConfig {
		err := m.addArtNet(ctx, conf)
		if err != nil {
			m.logger.Error("error while adding new Artnet device", zap.Error(err), zap.Any("conf", conf))
		}
	}

	for _, conf := range dmxDeviceConfig {
		err := m.addDMX(ctx, conf)
		if err != nil {
			m.logger.Error("error while adding new DMX device", zap.Error(err), zap.Any("conf", conf))
		}
	}
}

// Function processing set channel command
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

// Function processing increment channel command
func (m *manager) ProcessIncrementChannel(ctx context.Context, command models.IncrementChannel) error {
	dev, err := m.checkDevice(command.DeviceAlias)
	if err != nil {
		return err
	}

	err = dev.IncrementChannel(ctx, command)
	if err != nil {
		return fmt.Errorf("device with alias %v setting value error: %v", dev.GetAlias(), err)
	}
	return nil
}

// Function processing blackout command
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

// Function processing set scene command
func (m *manager) ProcessSetScene(ctx context.Context, command models.SetScene) error {
	dev, err := m.checkDevice(command.DeviceAlias)
	if err != nil {
		return err
	}

	err = dev.SetScene(ctx, command)
	if err != nil {
		return fmt.Errorf("device with alias %v blackout error: %v", dev.GetAlias(), err)
	}
	return nil
}

// Function processing save scene command
func (m *manager) ProcessSaveScene(ctx context.Context, command models.SaveScene) error {
	dev, err := m.checkDevice(command.DeviceAlias)
	if err != nil {
		return err
	}

	err = dev.SaveScene(ctx)
	if err != nil {
		return fmt.Errorf("device with alias %v blackout error: %v", dev.GetAlias(), err)
	}
	return nil
}

// Function checks devices list containing device with specified alias
func (m *manager) checkDevice(deviceAlias string) (device.Device, error) {
	dev, devExist := m.devices[deviceAlias]
	if !devExist {
		return nil, fmt.Errorf("dmx-device with alias %v not found", deviceAlias)
	}
	return dev, nil

}

// Function adds DMX device to device list
func (m *manager) addDMX(ctx context.Context, conf device.DMXConfig) error {
	newDMX, err := dmx.NewDMXDevice(ctx, m.signals, conf, m.logger, m.checkManager)
	if err != nil {
		return fmt.Errorf("error with add device: %v", err)
	}
	m.devices[newDMX.GetAlias()] = newDMX
	return nil
}

// Function adds Artnet device to device list
func (m *manager) addArtNet(ctx context.Context, conf device.ArtNetConfig) error {
	newArtNet, err := artnet.NewArtNetDevice(ctx, m.signals, conf, m.logger, m.checkManager)
	if err != nil {
		return fmt.Errorf("error with add device: %v", err)
	}
	m.devices[newArtNet.GetAlias()] = newArtNet
	return nil
}

// Function removes any device from device list
func (m *manager) removeDevice(_ context.Context, alias string) error {
	dev := m.devices[alias]
	dev.Close()
	delete(m.devices, alias)
	return nil
}
