package main

import (
	_ "embed"
	"fmt"
	"os"
	"os/signal"

	tray "github.com/getlantern/systray"

	"ledctl3/internal/client"
	"ledctl3/internal/client/config"
)

//go:embed icon.ico
var icon []byte

func main() {
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

	exit := make(chan os.Signal, 1)
	signal.Notify(exit, os.Interrupt)

	go tray.Run(func() {
		tray.SetIcon(icon)

		// Setup menu items
		title := tray.AddMenuItem("ledctl", "")
		title.Disable()

		tray.AddSeparator()

		videoMode := tray.AddMenuItem("Ambilight - Video", "")
		audioMode := tray.AddMenuItem("Ambilight - Audio", "")

		tray.AddSeparator()

		quit := tray.AddMenuItem("Quit", "")
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
