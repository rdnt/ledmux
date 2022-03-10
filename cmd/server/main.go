package main

import (
	"fmt"
	"ledctl3/internal/server"
	"os"
	"os/signal"
)

func main() {
	ctl, err := server.New()
	if err != nil {
		panic(err)
	}

	err = ctl.Start()
	if err != nil {
		panic(err)
	}

	exit := make(chan os.Signal, 1)
	signal.Notify(exit, os.Interrupt)
	<-exit

	fmt.Println("exit")
}
