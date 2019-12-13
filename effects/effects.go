package effects

import (
	"github.com/sht/ambilight/ws281x"
	"time"
)

// Ambilight simply sets each led's color based on the received data
func Ambilight(ws *ws281x.Engine, data []byte) {
	ws.Add()
	defer ws.Done()
	// start := time.Now()
	//
	// elapsed := time.Since(start)
	// log.Printf("Binomial took %s", elapsed)

	select {
	default:
		// Initialize variables
		var r, g, b uint8
		for i := 0; i < ws.LedsCount; i++ {
			// Parse color data for current LED
			r = uint8(data[i*3])
			g = uint8(data[i*3+1])
			b = uint8(data[i*3+2])
			// Set the current LED's color
			// Not need to check for error
			_ = ws.SetLedColor(i, r, g, b)
		}
		// Render the leds, ignoring errors
		_ = ws.Render()
	case <-ws.Cancel():
		// Stop executing if signaled from main process
		return
	}
}

// Rainbow loops a smooth gradient color swipe across the led strip
func Rainbow(ws *ws281x.Engine, payload []byte) {
	// fmt.Println("-- init rainbow")
	ws.Add()
	defer ws.Done()
	var r, g, b int
	for {
		// Not need to check for error
		// Loop color brightness 3 times
		for i := 0; i < 256*3; i++ {
			select {
			default:
				// For each of the leds
				for j := 0; j < ws.LedsCount; j++ {
					if (i+j)%768 < 256 {
						// transition from red to green
						r = (255 - (i+j)%256)
						g = (i + j) % 256
						b = 0
					} else if (i+j)%768 < 512 {
						// transition from green to blue
						r = 0
						g = (255 - (i+j)%256)
						b = (i + j) % 256
					} else {
						// transition from blue to red
						r = (i + j) % 256
						g = 0
						b = (255 - (i+j)%256)
					}
					_ = ws.SetLedColor(j, uint8(r), uint8(g), uint8(b))
				}
				// Render the leds, ignoring errors
				_ = ws.Render()
				time.Sleep(16 * time.Millisecond)
			case <-ws.Cancel():
				// Stop executing if signaled from main process
				// fmt.Println("-- stopping rainbow!")
				return
			}
		}
	}
}
