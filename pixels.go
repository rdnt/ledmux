package main

import (
	"fmt"
	"time"

	"github.com/kbinani/screenshot"
)

func main() {

	ticker := time.NewTicker(10 * time.Millisecond)
	go func() {
		for _ = range ticker.C {
			capture()
		}
	}()

	time.Sleep(3000 * time.Millisecond)
	ticker.Stop()
	fmt.Println("Stop")

}
