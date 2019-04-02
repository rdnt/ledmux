package main

import (
    "net"
    "fmt"
    "os"
    "time"
    "sync"
    "strconv"
    "./ambilight"
    "github.com/cretz/go-scrap"
)

func main() {
    // Get all arguments except for program
    args := os.Args[1:]
    // Make sure we get exactly 3 arguments
    if len(args) != 4 {
        fmt.Println("Usage: ./client [ip] [port] [led_count] [framerate]")
        return
    }
    // Validate destination IP address
    ip := net.ParseIP(args[0])
    if ip.To4() == nil {
        fmt.Println(args[0], ": Not a valid IPv4 address.")
        return
    }
    // Validate destination port is in allowed range (1024 - 65535)
    port, err := strconv.Atoi(args[1])
    if err != nil || port < 1024 || port > 65535 {
        fmt.Println(args[1], ": Port out of range. (1024 - 65535)")
        return
    }
    // Validate leds count (should be the same with controller)
    led_count, err := strconv.Atoi(args[2])
    if err != nil || led_count < 1 || led_count > 65535 {
        fmt.Println(args[2], ": Invalid LED count. (1 - 65535)")
        return
    }
    // Validate capturing frames per second
    framerate, err := strconv.Atoi(args[3])
    if err != nil || framerate < 1 || framerate > 144 {
        fmt.Println(args[3], ": Invalid framerate. (1 - 144)")
        return
    }
    // Create the Ambilight object
    var amb = ambilight.Init(
        ip.String(),
        port,
        led_count,
    )
    fmt.Println("Trying to connect to Ambilight server...")
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
            // Grab a frame capture once one is ready (max ~ 60 frames per second)
            img := AcquireImage(c, int(framerate))
            // Get width and height of the display
            width, height := GetDisplayResolution(c)
            // Get the LED data from the borders of the captured image
            data := CaptureBounds(img, width, height, int(amb.Count))
            // Send the color data to the server
            err := amb.Send(conn, data)
            if err != nil {
                // Close the connection
                err := amb.Disconnect(conn)
                if err != nil {
                    fmt.Println("Connection could not be closed.")
                    fmt.Println("Exiting.")
                    os.Exit(3)
                }
                // Error occured, stop and try to re-establish connection
                fmt.Println("Connection closed.")
                fmt.Println("Retrying...")
                break
            }
        }
        // Try to reconnect every second (let's not flood the server shall we)
        time.Sleep(1 * time.Second)
    }
}

func AcquireImage(c *scrap.Capturer, framerate int) (*scrap.FrameImage) {
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
        ticker := time.NewTicker(time.Duration(1000 / framerate) * time.Millisecond)
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
    		r, g, b, _ = img.At(width - x - 1, height - 1).RGBA()
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
    		r, g, b, _ = img.At(0, height - y - 1).RGBA()
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
    // Initialize the leds color data + 1 byte for the mode
    led_data := make([]uint8, count * 3 + 1)
    // Loop for LED count
    for i := 0; i < count; i++ {
        // Initialize the color values to zero
        var r, g, b int = 0, 0, 0
        // Loop all pixels in the current segment
        for j := 0; j < pixels_per_segment; j ++ {
            // Calculate the offset (based on current segment)
            offset := pixels_per_segment * 3 * i
            // Add the casted color integer to the last value
            r += int(data[offset + j * 3]);
            g += int(data[offset + j * 3 + 1]);
            b += int(data[offset + j * 3 + 2]);
            // r = int(data[offset + j * 3]);
            // g = int(data[offset + j * 3 + 1]);
            // b = int(data[offset + j * 3 + 2]);
        }
        // Get the average by dividing the accumulated color value with the
        // count of the pixels in the segment
        r = r / pixels_per_segment
        g = g / pixels_per_segment
        b = b / pixels_per_segment

        // Modify the correct bytes on the LED data
        // Leaving the first byte untouched
        led_data[i * 3 + 1] = uint8(r)
        led_data[i * 3 + 2] = uint8(g)
        led_data[i * 3 + 3] = uint8(b)
    }
    mode := 'R'
    led_data[0] = uint8(mode)
    // Return the LED data
    return led_data
}
