package main

import (
	"fmt"
	"ledctl3/internal/client"
	"ledctl3/internal/client/config"
	"os"
	"os/signal"
)

func main() {
	//leds := 83
	//cap := "dxgi"

	//displays, err := dxgi.New()
	//if err != nil {
	//	panic(err)
	//}

	//amb, err := ambilight.New(
	//	ambilight.WithLedsCount(leds),
	//	ambilight.WithDisplayRepository(displays),
	//)
	//if err != nil {
	//	panic(err)
	//}

	//viz, err := audioviz.New(
	//	audioviz.WithLedsCount(leds),
	//	audioviz.WithAudioSource(nil),
	//)
	//if err != nil {
	//	panic(err)
	//}

	cfg, err := config.Load()
	if err != nil {
		panic(err)
	}

	s, err := client.New(
		client.WithConfig(cfg),
		//client.WithServerIP("192.168.1.101:4197"),
		////client.WithServerIP(":4197"),
		//client.WithLedsCount(leds),
		//client.WithDisplayCapturer(client.Capturer(cap)),
	)
	if err != nil {
		panic(err)
	}

	err = s.Start()
	if err != nil {
		panic(err)
	}
	defer s.Stop()

	//err = s.SetMode(serverrpi.Ambilight)
	//if err != nil {
	//	panic(err)
	//}

	exit := make(chan os.Signal, 1)
	signal.Notify(exit, os.Interrupt)
	<-exit

	fmt.Println("exit")
}
