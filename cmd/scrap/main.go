package main

import (
	"fmt"
	"image/png"
	"os"

	goscrap "github.com/rdnt/go-scrap"
)

func main() {
	sd, err := goscrap.GetDisplay(1)
	if err != nil {
		panic(err)
	}

	cap, err := goscrap.NewCapturer(sd)
	if err != nil {
		panic(err)
	}

	fmt.Println(cap.Width(), cap.Height())

	img, wouldBlock, err := cap.FrameImage()
	if err != nil {
		panic(err)
	}

	if wouldBlock {
		fmt.Println("WOULD BLOCK")
		return
	}

	img.Detach()

	f, err := os.OpenFile("frame.png", os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		fmt.Println(err)
	}
	defer f.Close()

	img.Stride = img.Height * 4

	// Encode to `PNG` with `DefaultCompression` level
	// then save to file
	err = png.Encode(f, img)
	if err != nil {
		panic(err)
	}

	fmt.Println("done")
}
