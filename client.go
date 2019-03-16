package main

import (
    "fmt"
    "os"
    "time"
    "sync"
    "./ambilight"
    //"image"
    "github.com/cretz/go-scrap"
)

func main() {
    // Create the Ambilight object
    var amb = ambilight.Init(
        "192.168.1.101",
        4197,
        84,
    )
    fmt.Println("Initializing client...")
    // Get primary display
    d, err := scrap.PrimaryDisplay()
	if err != nil {
        panic(err)
		return
	}
	// Create capturer for it
	c, err := scrap.NewCapturer(d)
    if err != nil {
        panic(err)
		return
	}
    // Try to reconnect if connection is closed
    for {
        // Connect to remote server
        conn := amb.Connect()
        // Screen capture and send data once we have an image, loop until
        // There is an error during transmission
        for {
            // Get the color data averages for each led
            // Grab a frame capture once one is ready (max ~ 60 per second)
            img := AcquireImage(c)
            // Get width and height of the display
            width, height := GetDisplayResolution(c)
            // Get the LED data from the borders of the captured image
            data := CaptureBounds(img, width, height, amb.Count)
            // Send the color data to the server
            err := amb.Send(conn, data)
            if err != nil {
                fmt.Println("Transmission failed.")
                // Close the connection
                err := amb.Disconnect(conn)
                if err != nil {
                    fmt.Println("Connection could not be closed.")
                    os.Exit(3)
                }
                fmt.Println("Connection closed.")
                // Error occured, stop and try to re-establish connection
                //close(stop)
                break
            }
        }
        // Try to reconnect every second (let's not flood the server shall we)
        time.Sleep(1 * time.Second)
        fmt.Println("Re-trying to connect...")
    }
}

func AcquireImage(c *scrap.Capturer) (*scrap.FrameImage) {
    // Initialize a new waitgroup
    var wg sync.WaitGroup
    wg.Add(1)
    // Initialize image object
    var img *scrap.FrameImage
    var err error
    // Get an image once it is available, trying once every ~1/60th of a second
    go func() {
        // Release waitgroup once done
        defer wg.Done()
        // Start a new ticker
        ticker := time.NewTicker(16 * time.Millisecond)
        // Stop the ticker once the routine is complete
        defer ticker.Stop()
        // Repeat
        for range ticker.C {
            // Try to capture
            img, _, err = c.FrameImage();
            if img != nil || err != nil {
                // Image is available
                if img != nil {
                    // Detach the image so it's safe to use after this method
                    img.Detach()
                    // Break the loop
                    break
                }
            }
        }
    }()
    // Wait until an image is ready
    wg.Wait()
    // Dispatch the image
    return img
}

func GetDisplayResolution(c *scrap.Capturer) (width int, height int) {
    // Get width and height from capturer
    width = c.Width()
    height = c.Height()
    // Return them
    return width, height
}

func CaptureBounds(img *scrap.FrameImage, width int, height int, count int) []uint8 {
	// Initialize new waitgroup
    var wg sync.WaitGroup
    wg.Add(4)
    // Two horizontal two vertical, 3 colors (3 bytes) for each pixel
    data := make([]uint8, width * 3 * 2 + height * 3 * 2)
    // Create a wait group and add the four routines
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
        for x := 0; x < width; x++ {
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
    		r, g, b, _ = img.At(0, y).RGBA()
            data[offset + y * 3] = uint8(r)
            data[offset + y * 3 + 1] = uint8(g)
            data[offset + y * 3 + 2] = uint8(b)
    	}
    }()
    // Wait until all routines are complete
    wg.Wait()
    // Lets get the approximate segment size in bytes for each of the pixels
    segment_size := int((width * 3 * 2 + height * 3 * 2) / count)
    // We want the actual pixels to be divisible by 3 (3 bytes = rgb for 1 pixel)
    pixels_per_segment := segment_size / 3
    // Initialize the leds color data
    led_data := make([]uint8, count * 3)
    // Loop for LED count
    for i := 0; i < count; i++ {
        // Initialize the color values to zero
        var r, g, b int = 0, 0, 0
        // Loop all pixels in the current segment
        for j := 0; j < pixels_per_segment; j ++ {
            // Calculate the offset (based on current segment)
            offset := segment_size * i
            // Add the casted color integer to the last value
            r += int(data[offset + j * 3]);
            g += int(data[offset + j * 3 + 1]);
            b += int(data[offset + j * 3 + 2]);
        }
        // Get the average by dividing the accumulated color value with the
        // count of the pixels in the segment
        r = r / pixels_per_segment
        g = g / pixels_per_segment
        b = b / pixels_per_segment
        // Modify the correct bytes on the LED data
        led_data[i * 3] = uint8(r)
        led_data[i * 3 + 1] = uint8(g)
        led_data[i * 3 + 2] = uint8(b)
    }
    // Return the LED data
    return led_data
}
