package internal

import (
	"context"
	"fmt"
	"go.uber.org/zap"
)

func NewManager(logger *zap.Logger) *dmxManager {
	return &dmxManager{
		devices: make(map[string]DMXDevice),
		logger:  logger,
	}
}

type dmxManager struct {
	devices map[string]DMXDevice
	logger  *zap.Logger
}

func (m *dmxManager) UpdateDMXDevices(ctx context.Context, dmxConfigs []DMXConfig) {
	dmxSet := make(map[string]bool, len(dmxConfigs))
	for _, dev := range dmxConfigs {
		dmxSet[dev.Alias] = true
	}

	// Deleting the devices that have not been received
	for alias := range m.devices {
		if !dmxSet[alias] {
			err := m.deleteDMX(ctx, alias)
			if err != nil {
				m.logger.Error("error with update devices", zap.Error(err), zap.Any("alias", alias))
			}
		}
	}

	// Connect to new devices
	for _, conf := range dmxConfigs {
		if _, exist := m.devices[conf.Alias]; !exist {
			err := m.addDMX(ctx, conf)
			if err != nil {
				m.logger.Error("error with update devices", zap.Error(err), zap.Any("conf", conf))
			}
		}
	}
}

func (m *dmxManager) ProcessSetChannel(ctx context.Context, command SetChannel) error {
	dev, err := m.getDMXForAction(command.DMXAlias)
	if err != nil {
		return err
	}

	err = dev.SetValueToChannel(ctx, command)
	if err != nil {
		return fmt.Errorf("device with alias %v setting value error: %v", dev.GetAlias(), err)
	}
	return nil
}

func (m *dmxManager) ProcessBlackout(ctx context.Context, command Blackout) error {
	dev, err := m.getDMXForAction(command.DMXAlias)
	if err != nil {
		return err
	}

	err = dev.Blackout(ctx)
	if err != nil {
		return fmt.Errorf("device with alias %v blackout error: %v", dev.GetAlias(), err)
	}
	return nil
}

func (m *dmxManager) getDMXForAction(DMXAlias string) (DMXDevice, error) {
	dev, devExist := m.devices[DMXAlias]
	if !devExist {
		return nil, fmt.Errorf("dmx-device with alias %v not found", DMXAlias)
	}
	return dev, nil

}

func (m *dmxManager) addDMX(ctx context.Context, conf DMXConfig) error {
	newDMX, err := NewDMXDevice(ctx, conf)
	if err != nil {
		return fmt.Errorf("error with add device: %v", err)
	}
	m.devices[newDMX.GetAlias()] = newDMX
	return nil
}

func (m *dmxManager) deleteDMX(_ context.Context, alias string) error {
	delete(m.devices, alias)
	return nil
}
