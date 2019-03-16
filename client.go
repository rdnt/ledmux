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
    fmt.Printf("%+v", c)

    // Try to reconnect if connection is closed
    for {
        // Connect to remote server
        conn := amb.Connect()
        // Send data indefinitely


        // Capture every approximately 1/30th of a second (30fps)


        stop := make(chan struct{})


        go func() {

            ticker := time.NewTicker(16 * time.Millisecond)

            defer ticker.Stop()

            for range ticker.C {
                // Get the color data averages for each led
                // Grab a frame capture once one is ready (max ~ 60 per second)
                img, err := AcquireImage(c)
                if err != nil {
                    panic(err)
                }
                data := captureBounds(img, amb.Count)
                // Send the color data to the server
                err = amb.Send(conn, data)
                if err != nil {
                    fmt.Println("Transmission failed.")
                    // Close the connection
                    err := amb.Disconnect(conn)
                    if err != nil {
                        fmt.Println("Connection could not be closed.")
                        os.Exit(3)
                    }
                    fmt.Println("Connection closed.")
                    // Break and try to re-establish connection
                    //return
                    close(stop)

                }

                // stop
            }
        }()

        <-stop

        // Try to reconnect every second (let's not flood the server shall we)
        time.Sleep(1 * time.Second)
        fmt.Println("Re-trying to connect...")
    }
}

func AcquireImage(c *scrap.Capturer) (*scrap.FrameImage, error) {

    var wg sync.WaitGroup

    wg.Add(1)

    var img *scrap.FrameImage
    var err error

    go func() {

        defer wg.Done()

        ticker := time.NewTicker(16 * time.Millisecond)

        defer ticker.Stop()

        for range ticker.C {
            img, _, err = c.FrameImage();
            // Detach the image so it's safe to use after this method
            if img != nil || err != nil {
                if img != nil {
                    img.Detach()
                    break
                }

            }
        }
    }()

    wg.Wait()

    return img, err

}

func GetDisplayResolution(c *scrap.Capturer) (width int, height int) {
    width = c.Width()
    height = c.Height()
    return width, height
}

func captureBounds(img *scrap.FrameImage, count int) []uint8 {
	// Get main display's bounds
    var wg sync.WaitGroup
	//bounds := screenshot.GetDisplayBounds(0)



	// Get an image, trying until one available
	//for {

        // img, _, err := c.FrameImage();
        // if img != nil {
        //     img.Detach()
        // }
        // fmt.Printf("3")
        // fmt.Printf(">>> %+v <<<\n", img)
        // if err != nil {
        //     fmt.Printf("4")
        //     return []uint8{}, err
        // }
		// Sleep 17ms (~1/60th of a second)
		//time.Sleep(17 * time.Millisecond)
	//}



    // Capture a screenshot
	//img, err := screenshot.CaptureScreen(bounds)

	// Get width and height in pixels
	width := 1920
	height := 1080
    // Two horizontal two vertical, 3 colors (3 bytes) for each pixel
    data := make([]uint8, width * 3 * 2 + height * 3 * 2)
    // Create a wait group and add the four routines
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
