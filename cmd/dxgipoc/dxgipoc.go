package main

import (
	"context"
	"os"
	"os/signal"
	"time"

	"ledctl3/internal/client/controller/ambilight"
	"ledctl3/internal/client/controller/ambilight/capturer/dxgi"
)

func main() {
	defer time.Sleep(1 * time.Second)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	//go func() {
	//	capturer, err := dxgi.New()
	//	if err != nil {
	//		panic(err)
	//	}
	//
	//	displays, err := capturer.All()
	//	if err != nil {
	//		panic(err)
	//	}
	//
	//	frames := make(chan []byte)
	//
	//	id := displays[0].Id()
	//
	//	go displays[0].SyncCapture(ctx, frames)
	//
	//	for range frames {
	//		fmt.Print(id)
	//	}
	//}()

	capturer, err := dxgi.New()
	if err != nil {
		panic(err)
	}

	displays, err := capturer.All()
	if err != nil {
		panic(err)
	}

	for _, d := range displays {
		frames := d.Capture(ctx, 1)

		go func(d ambilight.Display) {
			for range frames {
				if d.Id() == 0 {
					os.Exit(0)
				}
			}
		}(d)
	}

	exit := make(chan os.Signal, 1)
	signal.Notify(exit, os.Interrupt)
	<-exit
}
