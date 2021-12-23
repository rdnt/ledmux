package main

import (
	"context"
	"fmt"
	"ledctl3/internal/client/controller/ambilight/capturer/dxgi"
	"ledctl3/internal/client/interfaces"
	"os"
	"os/signal"
	"time"
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
		frames := d.Capture(ctx)

		go func(d interfaces.Display) {
			for range frames {
				fmt.Print(d.Id())
			}
		}(d)
	}

	exit := make(chan os.Signal, 1)
	signal.Notify(exit, os.Interrupt)
	<-exit
}
