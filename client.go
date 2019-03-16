package main

import (
    "fmt"
    "os"
    "time"
    "sync"
    "./ambilight"
    "github.com/kbinani/screenshot"
)

// IP and port to connect to
const IP = "127.0.0.1"
const port = 4197
const count = 75

func main() {

    // Create the Ambilight object
    var amb = ambilight.Init(
        IP,
        port,
        count,
    )

    // Try to reconnect if connection is closed
    for {
        fmt.Println("Connecting...")
        // Connect to remote server
        conn := amb.Connect()
        // Send data indefinitely
        for {
            fmt.Println("Sending data...")
            //err := amb.Send(conn, data)
            data := captureBounds(count)
            err := amb.Send(conn, data)
            if err != nil {
                fmt.Println("Failed to send the data.")
                err := amb.Disconnect(conn)
                if err != nil {
                    fmt.Println("Connection could not be closed.")
                    os.Exit(3)
                }
                fmt.Println("Connection closed.")
                break
            }
            time.Sleep(1 * time.Second)



        }
        // Try to reconnect every second
        time.Sleep(1 * time.Second)
    }
}

func captureBounds(count int) []uint8 {
	// Get display 0's bounds
	bounds := screenshot.GetDisplayBounds(0)
    // Capture the screenshot
	img, err := screenshot.CaptureRect(bounds)
	if err != nil {
		panic(err)
	}

	var r, g, b uint32

	// Get width and height in pixels
	width := bounds.Dx()
	height := bounds.Dy()



    // Two horizontal two vertical, 3 colors (3 bytes) for each pixel
    data := make([]uint8, width * 3 * 2 + height * 3 * 2)






    var wg sync.WaitGroup

    // you can also add these one at
    // a time if you need to

    wg.Add(4)
    // Top
    go func() {

        defer wg.Done()

        for x := 0; x < width; x++ {
    		r, g, b, _ = img.At(x, 0).RGBA()

            // offset = 0

            data[x * 3] = uint8(r)
            data[x * 3 + 1] = uint8(g)
            data[x * 3 + 2] = uint8(b)
    	}

    }()
    go func() {
        defer wg.Done()

        offset := width * 3

        for y := 0; y < height; y++ {

    		r, g, b, _ = img.At(width - 1, y).RGBA()

            data[offset + y * 3] = uint8(r)
            data[offset + y * 3 + 1] = uint8(g)
            data[offset + y * 3 + 2] = uint8(b)

    	}
    }()
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
    go func() {
        defer wg.Done()

        offset := width * 3 * 2 + height * 3
        fmt.Println(offset)

        for y := height - 1; y >= 0; y-- {

    		r, g, b, _ = img.At(0, y).RGBA()

            data[offset + y * 3] = uint8(r)
            data[offset + y * 3 + 1] = uint8(g)
            data[offset + y * 3 + 2] = uint8(b)

    	}
    }()

    wg.Wait()

    // Lets get the approximate segment size in bytes for each of the pixels
    segment_size := int((width * 3 * 2 + height * 3 * 2) / count)
    // We want the actual pixels to be divisible by 3 (3 bytes = rgb for 1 pixel)
    segment_size = segment_size / 3 // 80
    fmt.Println(segment_size) // 240

    pixel_data := make([]uint8, count * 3)

    for i := 0; i < count; i++ {

        var r, g, b int = 0, 0, 0

        for j := 0; j < segment_size; j ++ {
            offset := segment_size * i

            r += int(data[offset + j * 3]);
            g += int(data[offset + j * 3 + 1]);
            b += int(data[offset + j * 3 + 2]);




        }


        r = r / segment_size
        g = g / segment_size
        b = b / segment_size

        pixel_data[i * 3] = uint8(r)
        pixel_data[i * 3 + 1] = uint8(g)
        pixel_data[i * 3 + 2] = uint8(b)
        // fmt.Printf("#")
        // fmt.Printf("%X", r)
        // fmt.Printf("%X", g)
        // fmt.Printf("%X\n", b)


    }


    return pixel_data
}
