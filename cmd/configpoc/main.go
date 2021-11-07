package main

import (
	"github.com/sanity-io/litter"
	"ledctl3/internal/client/config"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		panic(err)
	}

	litter.Dump(cfg)
}
