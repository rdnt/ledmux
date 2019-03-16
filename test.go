package main

import (
    "fmt"
    "time"
    "sync"
    "github.com/kbinani/screenshot"
)

func main() {
    fmt.Println("Initializing test...")
    // Try to reconnect if connection is closed
    for {
        data := captureBounds(84)
        if data[0] == 0 {

        }
        time.Sleep(30 * time.Millisecond)
    }
}

func captureBounds(count int) []uint8 {
	// Get main display's bounds
	bounds := screenshot.GetDisplayBounds(0)
    // Capture a screenshot
	img, err := screenshot.CaptureRect(bounds)
	if err != nil {
		panic(err)
	}
	// Get width and height in pixels
	width := bounds.Dx()
	height := bounds.Dy()
    // Two horizontal two vertical, 3 colors (3 bytes) for each pixel
    data := make([]uint8, width * 3 * 2 + height * 3 * 2)
    // Create a wait group and add the four routines
    var wg sync.WaitGroup
    wg.Add(4)
    // Initialize RGB values
	var r, g, b uint32
    // Capture all the top edge pixel data
    go func() {
        // Once complete set as done
        defer wg.Done()
        // Offset is 0 for the top edge, we are going clockwise
        // Loop all the pixels
        for x := 0; x < width; x++ {
            // Parse RGB data
    		r, g, b, _ = img.At(x, 0).RGBA()
            // Convert the RGB values to uint8 and modify the correct bytes
            data[x * 3] = uint8(r)
            data[x * 3 + 1] = uint8(g)
            data[x * 3 + 2] = uint8(b)
    	}
    }()
    // Right
    go func() {
        defer wg.Done()
        // Offset is 3 times the width of the display,
        // since we need 3 bytes per pixel (RGB values)
        offset := width * 3
        for y := 0; y < height; y++ {
    		r, g, b, _ = img.At(width - 1, y).RGBA()
            data[offset + y * 3] = uint8(r)
            data[offset + y * 3 + 1] = uint8(g)
            data[offset + y * 3 + 2] = uint8(b)
    	}
    }()
    // Bottom
    go func() {
        defer wg.Done()
        offset := width * 3 + height * 3
        for x := width - 1; x >= 0; x-- {
    		r, g, b, _ = img.At(x, height - 1).RGBA()
            data[offset + x * 3] = uint8(r)
            data[offset + x * 3 + 1] = uint8(g)
            data[offset + x * 3 + 2] = uint8(b)
    	}
    }()
    // Left
    go func() {
        defer wg.Done()
        offset := width * 3 * 2 + height * 3
        for y := 0; y < height; y++ {
    		r, g, b, _ = img.At(0, height - y).RGBA()
            data[offset + y * 3] = uint8(r)
            data[offset + y * 3 + 1] = uint8(g)
            data[offset + y * 3 + 2] = uint8(b)
    	}
    }()
    // Wait until all routines are complete
    wg.Wait()

    // Return the LED data
    return data
}
