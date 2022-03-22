package main

import (
	"context"
	"fmt"
	"ledctl3/internal/client/controller/ambilight/capturer/dxgi"
	"ledctl3/internal/client/interfaces"
	"sync"
	"time"
)

func main() {
	repo, err := dxgi.New()
	if err != nil {
		panic(err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var wg sync.WaitGroup
	wg.Add(2)

	for {
		now := time.Now()

		displays, err := repo.All()
		if err != nil {
			panic(err)
		}

		fmt.Println(time.Since(now))
		fmt.Println(displays)

		for _, d := range displays {
			frames := d.Capture(ctx, 1000)
			fmt.Println(time.Since(now))
			go func(d interfaces.Display) {
				for range frames {
					wg.Done()
					break
				}
			}(d)
		}

		wg.Wait()
		cancel()
		time.Sleep(1 * time.Second)
	}

}
