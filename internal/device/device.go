package device

import (
	"context"
	"git.miem.hse.ru/hubman/dmx-executor/internal/models"

)

type Device interface {
	GetAlias() string
	SetValueToChannel(ctx context.Context, command models.SetChannel) error
	Blackout(ctx context.Context) error
}