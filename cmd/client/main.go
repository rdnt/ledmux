package main

import (
	"fmt"
	"os"
	"os/signal"

	tray "github.com/getlantern/systray"

	"ledctl3/internal/client"
	"ledctl3/internal/client/config"
)

func main() {
	//leds := 83
	//cap := "dxgi"

	//displays, err := dxgi.New()
	//if err != nil {
	//	panic(err)
	//}

	//amb, err := video.New(
	//	video.WithLedsCount(leds),
	//	video.WithDisplayRepository(displays),
	//)
	//if err != nil {
	//	panic(err)
	//}

	//viz, err := audio.New(
	//	audio.WithLedsCount(leds),
	//	audio.WithAudioSource(nil),
	//)
	//if err != nil {
	//	panic(err)
	//}

	cfg, err := config.Load()
	if err != nil {
		panic(err)
	}

	c, err := client.New(
		client.WithConfig(cfg),
	)
	if err != nil {
		panic(err)
	}

	err = c.Start()
	if err != nil {
		panic(err)
	}
	defer c.Stop()

	//err = s.SetMode(serverrpi.Ambilight)
	//if err != nil {
	//	panic(err)
	//}

	exit := make(chan os.Signal, 1)
	signal.Notify(exit, os.Interrupt)

	go tray.Run(func() {
		// Setup menu items
		title := tray.AddMenuItem("ledctl", "")
		title.Disable()

		tray.AddSeparator()

		videoMode := tray.AddMenuItem("ambilight/video", "")
		audioMode := tray.AddMenuItem("ambilight/audio", "")

		tray.AddSeparator()
		
		quit := tray.AddMenuItem("quit", "")
		// Run an infinite loop on goroutine to detect button presses
		go func() {
			for {
				select {
				case <-quit.ClickedCh:
					tray.Quit()
					return
				// Change the amb.Mode once a different mode is clicked
				case <-videoMode.ClickedCh:
					c.DefaultMode = "video"
					// TODO: refine mode changing through tray
					c.Stop()
					c.Start()
				case <-audioMode.ClickedCh:
					c.DefaultMode = "audio"
					c.Stop()
					c.Start()
				}
			}
		}()
	}, func() {
		exit <- os.Interrupt
	})

	<-exit

	fmt.Println("exit")
}
