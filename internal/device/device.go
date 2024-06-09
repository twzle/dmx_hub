package device

import (
	"context"


	"git.miem.hse.ru/hubman/dmx-executor/internal/models"
)

// Represenation of channel entity
type Channel struct {
	UniverseChannelID int
	Value             int
}

// Represenation of scene entity
type Scene struct {
	Alias      string
	ChannelMap map[int]Channel
}

// Represenation of abstract device entity
type Device interface {
	GetAlias() string
	SetScene(ctx context.Context, command models.SetScene) error
	SaveScene(ctx context.Context) error
	SetChannel(ctx context.Context, command models.SetChannel) error
	IncrementChannel(ctx context.Context, command models.IncrementChannel) error
	WriteValueToChannel(command models.SetChannel) error
	WriteUniverseToDevice() error
	Blackout(ctx context.Context) error
	Close()
}