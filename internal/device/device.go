package device

import (
	"context"
	"git.miem.hse.ru/hubman/dmx-executor/internal/models"
	"git.miem.hse.ru/hubman/dmx-executor/internal/config"

)

type Device interface {
	GetAlias() string
	SetValueToChannel(ctx context.Context, command models.SetChannel) error
	SetUniverse(universe []config.ChannelRange)
	Blackout(ctx context.Context) error
}