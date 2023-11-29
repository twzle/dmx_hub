package main

import (
	"github.com/akualab/dmx"
	"log"
	"time"
)

func main() {

	//r := rand.New(rand.NewSource(99))
	log.Printf("start DMX")
	dmxDevice, e := dmx.NewDMXConnection("/dev/ttyUSB0")
	if e != nil {
		log.Fatal(e)
	}

	// Send RGB
	// 1: brightness/flash, 2: red, 3: blue, 4: green
	//dmxDevice.ChannelMap(1, 2, 4, 3)
	//dmxDevice.SendRGB(130, 100, 100, 100)
	time.Sleep(1 * time.Second)

	// Initial color.
	dmxDevice.SetChannel(1, 100)
	dmxDevice.SetChannel(2, 70)
	dmxDevice.SetChannel(3, 130)
	dmxDevice.SetChannel(4, 180)
	dmxDevice.Render()

	//for {
	//
	//	// Wait.
	//	time.Sleep(100 * time.Millisecond)
	//
	//	dmxDevice.SetChannel(1, 20)                // Intensity
	//	dmxDevice.SetChannel(2, byte(r.Intn(256))) // R
	//	dmxDevice.SetChannel(3, byte(r.Intn(256))) // G
	//	dmxDevice.SetChannel(4, byte(r.Intn(256))) // B
	//	dmxDevice.Render()
	//
	//}
}
