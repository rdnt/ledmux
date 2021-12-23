package main

import (
	"ledctl3/internal/server"
	"time"
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

	time.Sleep(100 * time.Hour)
}
