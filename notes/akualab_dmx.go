package main

import (
	"github.com/akualab/dmx"
	"log"
	"math/rand"
	"time"
)

func main() {

	r := rand.New(rand.NewSource(99))
	log.Printf("start DMX")
	dmxDevice, e := dmx.NewDMXConnection("/dev/ttyUSB0")
	if e != nil {
		log.Fatal(e)
	}
	time.Sleep(1 * time.Second)

	// Initial values.
	dmxDevice.SetChannel(1, 100)
	dmxDevice.SetChannel(2, 70)
	dmxDevice.SetChannel(3, 130)
	dmxDevice.SetChannel(4, 180)
	dmxDevice.Render()

	// sending random signals to first 4 channels
	for i := 0; i < 10; i++ {

		// Wait.
		time.Sleep(100 * time.Millisecond)

		dmxDevice.SetChannel(1, 20)
		dmxDevice.SetChannel(2, byte(r.Intn(256)))
		dmxDevice.SetChannel(3, byte(r.Intn(256)))
		dmxDevice.SetChannel(4, byte(r.Intn(256)))
		dmxDevice.Render()

	}

	// sending 0 to all channels
	for i := 1; i <= 512; i++ {
		dmxDevice.SetChannel(i, 0)
	}
	dmxDevice.Render()
	time.Sleep(100 * time.Millisecond)
}
