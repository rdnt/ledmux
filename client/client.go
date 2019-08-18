package main

import (
	"../ambilight"
	"fmt"
	"github.com/SHT/go-scrap"
	"log"
	"os"
	"sync"
	"time"
)

func main() {
	var amb = ambilight.Init()
	// Get primary display
	i := 0
	var displays []*scrap.Display
	for {
		d, err := scrap.GetDisplay(i)
		if err != nil && i == 0 {
			// Fatal error while getting the primary display.
			log.Fatal("No displays detected. Exiting.")
		} else if err != nil {
			// There was an error while loading further displays.
			// Break the loop and continue with the displays we have.
			break
		}
		displays = append(displays, d)
		i++
	}
	// Create a capturer for each display
	var capturers []*scrap.Capturer
	for i = range displays {
		c, err := scrap.NewCapturer(displays[i])
		if err != nil {
			log.Fatal(err)
		}
		capturers = append(capturers, c)
	}
	fmt.Println("Attempting to connect to the Ambilight server...")
	// Try to reconnect if connection is closed
	for {
		// Connect to the ambilight server
		conn := amb.Connect()
		// Screen capture and send data once we have an image, loop until
		// There is an error during transmission
		fmt.Println("connected")
		for {
			// Get the color data averages for each led
			// Grab a frame capture once one is ready (max ~ 60 frames per second)
			var wg sync.WaitGroup
			wg.Add(len(capturers))
			var data [][]byte

			for i, c := range capturers {
				go func() {
					// Release waitgroup once done
					defer wg.Done()
					//work
					img := AcquireImage(c, amb.Framerate)
					// Get width and height of the display
					width, height := GetDisplayResolution(c)
					// Get the LED data from the borders of the captured image
					bounds := CaptureBounds(img, width, height)
					data[i] = bounds
				}()
			}

			// Wait until all bounds are captured
			wg.Wait()
			for _, d := range data {
				fmt.Println("wtf")
				fmt.Printf("%X\n\n", d)
			}
			os.Exit(0)

			// Send the color data to the server
			err := amb.Send(conn, []byte{})
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

// AcquireImage captures an image from the GPU's backbuffer and returns it
func AcquireImage(c *scrap.Capturer, framerate int) *scrap.FrameImage {
	// Initialize a new waitgroup
	var wg sync.WaitGroup
	wg.Add(1)
	// Initialize image object
	var img *scrap.FrameImage
	// Get an image once it is available, trying once every ~1/60th of a second
	go func() {
		// Release waitgroup once done
		defer wg.Done()
		// Start a new ticker
		ticker := time.NewTicker(time.Duration(1000/framerate) * time.Millisecond)
		// Stop the ticker once the routine is complete
		defer ticker.Stop()
		// Repeat
		for range ticker.C {
			// Try to capture
			img, _, err := c.FrameImage()
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

// GetDisplayResolution returns the width and height of the target display
func GetDisplayResolution(c *scrap.Capturer) (width int, height int) {
	// Get width and height from capturer
	width = c.Width()
	height = c.Height()
	// Return them
	return width, height
}

// CaptureBounds decodes the pixel data from the specified image, stores the
// border pixels in four arrays, averages the borders based on the specified
// length of the strip, sets the operation mode to 'A' (Ambilight) and returns
// the color data as a bytes array
func CaptureBounds(img *scrap.FrameImage, width int, height int) []byte {
	// Initialize new waitgroup
	var wg sync.WaitGroup
	wg.Add(4)
	// Two horizontal two vertical, 3 colors (3 bytes) for each pixel
	data := make([]byte, width*3*2+height*3*2)
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
			// Convert the RGB values to byte and modify the correct bytes
			data[x*3] = byte(r)
			data[x*3+1] = byte(g)
			data[x*3+2] = byte(b)
		}
	}()
	// Right
	go func() {
		defer wg.Done()
		// Offset is 3 times the width of the display,
		// since we need 3 bytes per pixel (RGB values)
		offset := width * 3
		for y := 0; y < height; y++ {
			r, g, b, _ = img.At(width-1, y).RGBA()
			data[offset+y*3] = byte(r)
			data[offset+y*3+1] = byte(g)
			data[offset+y*3+2] = byte(b)
		}
	}()
	// Bottom
	go func() {
		defer wg.Done()
		offset := width*3 + height*3
		for x := 0; x < width; x++ {
			r, g, b, _ = img.At(width-x-1, height-1).RGBA()
			data[offset+x*3] = byte(r)
			data[offset+x*3+1] = byte(g)
			data[offset+x*3+2] = byte(b)
		}
	}()
	// Left
	go func() {
		defer wg.Done()
		offset := width*3*2 + height*3
		for y := 0; y < height; y++ {
			r, g, b, _ = img.At(0, height-y-1).RGBA()
			data[offset+y*3] = byte(r)
			data[offset+y*3+1] = byte(g)
			data[offset+y*3+2] = byte(b)
		}
	}()
	// Wait until all routines are complete
	wg.Wait()
	// Return the bounding pixels
	return data
}
