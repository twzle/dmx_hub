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

// Function initializes and returns Artnet controller entity
func NewArtNetController() *artnet.Controller {
	log := artnet.NewDefaultLogger()
	dev := artnet.NewController("ArtNet controller", net.ParseIP("127.0.0.1"), log)
	dev.Start()
	return dev
}

// Singletone initialization of artnet controller
func GetArtNetController() *artnet.Controller {
	once.Do(func() {
		controller = NewArtNetController()
	})

	return controller
}