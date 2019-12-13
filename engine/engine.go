package engine

import (
	"fmt"
	"github.com/cretz/go-scrap"
	ws281x "github.com/sht/ambilight/ws281x"
	"log"
	"net"
	"sync"
	"time"
)

// Engine asd
type Engine struct {
	IP   string `json:"ip,omitempty"`
	Port int    `json:"port"`

	LedsCount  int `json:"leds_count"`
	GPIOPin    int `json:"gpio_pin,omitempty"`
	Brightness int `json:"brightness,omitempty"`
	Framerate  int `json:"framerate,omitempty"`

	Listener   *net.TCPListener   `json:"-"`
	Clients    map[string]*Client `json:"-"`
	Connection net.Conn           `json:"-"`

	Ws *ws281x.Engine `json:"-"`

	Displays []*Display `json:"displays,omitempty"`

	Action    string `json:"-"`
	Connected bool   `json:"-"`
	Running   bool   `json:"-"`
}

// Display holds parameters like leds count and pixel offsets in the config
type Display struct {
	From Vector2 `json:"from"`
	To   Vector2 `json:"to"`
	Leds int     `json:"leds"`

	Capturer     *scrap.Capturer `json:"-"`
	Width        int             `json:"-"`
	Height       int             `json:"-"`
	BoundsOffset int             `json:"-"`
	BoundsSize   int             `json:"-"`
}

// Pixel represents the color data of a specific pixel
type Pixel struct {
	R uint8
	G uint8
	B uint8
}

// Vector2 is a X,Y pair of coordinates
type Vector2 struct {
	X int `json:"x"`
	Y int `json:"y"`
}

// Init returns an ambilight object with the default values and the specified
// IP port and leds count
func Init(mode string) (*Engine, error) {
	if mode == "server" {
		ws, err := ws281x.Init(18, 75, 255)
		if err != nil {
			return nil, err
		}
		return &Engine{
			Ws:         ws,
			IP:         "localhost",
			Port:       4197,
			LedsCount:  100,
			GPIOPin:    18,
			Framerate:  60,
			Brightness: 255,
			Action:     "",
			Clients:    make(map[string]*Client),
		}, nil
	} else if mode == "client" {
		return &Engine{
			IP:         "192.168.1.22",
			Port:       4197,
			LedsCount:  75,
			GPIOPin:    18,
			Framerate:  60,
			Brightness: 255,
			Action:     "",
			Displays: []*Display{
				&Display{
					From: Vector2{
						X: 300,
						Y: 1079,
					},
					To: Vector2{
						X: 1320,
						Y: 1079,
					},
					Leds: 75,
				},
			},
		}, nil
	}
	return nil, fmt.Errorf("invalid startup mode")
}

//
// // ReloadEngine asd
func (e *Engine) Reload(
	ledsCount uint16, gpioPin uint8, brightness uint8) error {
	err := e.Ws.Clear()
	if err != nil {
		return err
	}
	e.Ws.Fini()
	engine, err := ws281x.Init(int(gpioPin), int(ledsCount), int(brightness))
	if err != nil {
		return err
	}
	e.Ws = engine
	return nil
}

func (e *Engine) GetDisplays() ([]*Display, error) {
	var displays []*Display
	i := 0
	for {
		if i >= len(e.Displays) {
			break
		}
		d, err := scrap.GetDisplay(i)
		if err != nil && i == 0 {
			// Fatal error while getting the primary display.
			return nil, err
		} else if err != nil {
			// There was an error while loading further displays.
			// Possibly because no other display is present.
			// Break the loop and continue with the displays we have.
			// TODO @sht fix GetDisplays go-scrap function to return all
			// the available displays
			break
		}
		// Get display resolution
		width := d.Width()
		height := d.Height()
		// Create a capturer
		c, err := scrap.NewCapturer(d)
		if err != nil {
			return nil, err
		}
		// Validate coordinates of segment and filter the pixels based on the
		// segment's offset and size
		from := e.Displays[i].From
		to := e.Displays[i].To
		leds := e.Displays[i].Leds
		v1 := ValidateCoordinates(width, height, from.X, from.Y)
		v2 := ValidateCoordinates(width, height, to.X, to.Y)
		if !v1 || !v2 {
			log.Fatalf("Invalid coordinates for display %d.\n", i+1)
		}
		fromOffset := CalculateOffset(width, height, from.X, from.Y)
		toOffset := CalculateOffset(width, height, to.X, to.Y)
		size := GetPixSliceSize(width, height, fromOffset, toOffset)
		// Append this display on the displays array
		displays = append(displays, &Display{
			Capturer:     c,
			Width:        width,
			Height:       height,
			BoundsOffset: fromOffset,
			BoundsSize:   size,
			Leds:         leds,
		})
		i++
	}
	return displays, nil
}

// ValidateCoordinates asd
func ValidateCoordinates(width, height, x, y int) bool {
	if x == 0 || x == width-1 {
		if y >= 0 && y < height {
			return true
		}
	} else if y == 0 || y == height-1 {
		if x >= 0 && x < width {
			return true
		}
	}
	return false
}

// CalculateOffset returns the offset in pixels of the given edge coordinates,
// from the start of the monitor bounds (x:0, y:0), calculating clockwise
func CalculateOffset(width, height, x, y int) int {
	var offset int
	if x == 0 {
		offset = 2*width + height + (height - y)
	} else if x == width-1 {
		offset = width + y
	} else {
		if y == 0 {
			offset = x
		} else if y == height-1 {
			offset = width + height + (width - x)
		} else {
			return 0
		}
	}
	// offset = offset % (d.Width*2 + d.Height*2)
	return offset
}

// GetPixSliceSize returns the size the filtered pixels slice will have from
// the given offset coordinates
func GetPixSliceSize(width, height, from, to int) int {
	return (width*2 + height*2) - from + to
}

// AveragePixels returns the led color data after averaging the pixels slice,
// based on the leds count
func AveragePixels(pix []*Pixel, count int) []uint8 {
	pixelsPerLed := len(pix) / count
	// pixelsPerSegment := segmentSize / 3
	data := make([]uint8, count*3) // + 1 for the mode char
	// var total float64 = 0
	// for _, value := range x {
	// 	total += value
	// }
	// fmt.Println()

	// for i := len(pix)/2 - 1; i >= 0; i-- {
	// 	opp := len(pix) - 1 - i
	// 	pix[i], pix[opp] = pix[opp], pix[i]
	// }

	for i := 0; i < count; i++ {
		// Initialize the color values to zero
		var r, g, b int = 0, 0, 0
		// Loop all pixels in the current segment
		offset := pixelsPerLed * i
		if i == count-1 {
			// Grab the remaining n pixels
			// They will be at most len(pix) % count
			pixelsPerLed = len(pix) - (pixelsPerLed * (count - 1))
		}
		for j := 0; j < pixelsPerLed; j++ {
			// Calculate the offset (based on current segment)
			// Add the casted color integer to the last value
			r += int(pix[offset+j].R)
			g += int(pix[offset+j].G)
			b += int(pix[offset+j].B)
			// r = int(data[offset + j * 3]);
			// g = int(data[offset + j * 3 + 1]);
			// b = int(data[offset + j * 3 + 2]);
			// fmt.Println(offset + j)
		}
		// Get the average by dividing the accumulated color value with the
		// count of the pixels in the segment
		r = r / pixelsPerLed
		g = g / pixelsPerLed
		b = b / pixelsPerLed

		// Modify the correct bytes on the LED data
		// Leaving the first byte untouched
		data[i*3] = uint8(r)
		data[i*3+1] = uint8(g)
		data[i*3+2] = uint8(b)
	}

	return data

}

// FilterPixels returns the new pixels slice based on the given offset and size
func FilterPixels(d *Display, pix []*Pixel, offset, size int) []*Pixel {
	newBounds := make([]*Pixel, size) // 3 times the size (R G B bytes)
	for i := 0; i < size; i++ {
		// index = (i+from) % 6000 (adapted for 3 times the size for R G B)
		// fmt.Println((i + offset*3) % (len(pix)))
		newBounds[i] = pix[(i+offset)%len(pix)]
		// newBounds[i] = pix[(i+offset*3)%(len(pix))]
	}
	return newBounds
}

// AcquireImage captures an image from the GPU's backbuffer and returns it
func AcquireImage(c *scrap.Capturer, framerate int) *scrap.FrameImage {
	// Initialize a new waitgroup
	var wg sync.WaitGroup
	wg.Add(1)
	// Initialize image object
	var image *scrap.FrameImage
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
			if err != nil {
				log.Fatalf("Failed to capture framebuffer: %s\n.", err)
			}
			if img != nil || err != nil {
				// Image is available
				if img != nil {
					// Detach the image so it's safe to use after this method
					img.Detach()
					image = img
					// Break the loop
					break
				}
			}
		}
	}()
	// Wait until an image is ready
	wg.Wait()
	// Dispatch the image
	return image
}

// CapturePixels decodes the pixel data from the specified image, stores the
// border pixels in four arrays, averages the borders based on the specified
// length of the strip, sets the operation mode to 'A' (Ambilight) and returns
// the color data as a bytes array
func CapturePixels(img *scrap.FrameImage, width int, height int) []*Pixel {
	// Initialize new waitgroup
	var wg sync.WaitGroup
	wg.Add(4)
	// Two horizontal two vertical, 3 colors (3 bytes) for each pixel
	// data := make([]uint8, width*3*2+height*3*2)
	data := make([]*Pixel, width*2+height*2)
	// Create a wait group and add the four routines
	// Initialize RGB values
	// Capture all the top edge pixel data
	go func() {
		// Once complete set as done
		defer wg.Done()
		// Offset is 0 for the top edge, we are going clockwise
		// Loop all the pixels
		for x := 0; x < width; x++ {
			// Parse RGB data
			r, g, b, _ := img.At(x, 0).RGBA()
			// Convert the RGB values to byte and modify the correct bytes
			data[x] = &Pixel{
				R: uint8(r),
				G: uint8(g),
				B: uint8(b),
			}
		}
	}()
	// Right
	go func() {
		defer wg.Done()
		// Offset is 3 times the width of the display,
		// since we need 3 bytes per pixel (RGB values)
		offset := width
		for y := 0; y < height; y++ {
			r, g, b, _ := img.At(width-1, y).RGBA()
			data[offset+y] = &Pixel{
				R: uint8(r),
				G: uint8(g),
				B: uint8(b),
			}
		}
	}()
	// Bottom
	go func() {
		defer wg.Done()
		offset := width + height
		for x := 0; x < width; x++ {
			r, g, b, _ := img.At(width-x-1, height-1).RGBA()
			data[offset+x] = &Pixel{
				R: uint8(r),
				G: uint8(g),
				B: uint8(b),
			}
		}
	}()
	// Left
	go func() {
		defer wg.Done()
		offset := width*2 + height
		for y := 0; y < height; y++ {
			r, g, b, _ := img.At(0, height-y-1).RGBA()
			data[offset+y] = &Pixel{
				R: uint8(r),
				G: uint8(g),
				B: uint8(b),
			}
		}
	}()
	// Wait until all routines are complete
	wg.Wait()
	// Return the bounding pixels
	return data
}
