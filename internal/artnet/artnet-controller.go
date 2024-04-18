package artnet

import (
	"net"
	"sync"

	"github.com/jsimonetti/go-artnet"
)

var (
	controller *artnet.Controller
	once       sync.Once
)

func NewArtNetController() *artnet.Controller {
	log := artnet.NewDefaultLogger()
	dev := artnet.NewController("ArtNet controller", net.ParseIP("127.0.0.1"), log)
	dev.Start()
	return dev
}

func GetArtNetController() *artnet.Controller {
	once.Do(func() {
		controller = NewArtNetController()
	})

	return controller
}