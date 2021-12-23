#!/bin/bash

go mod vendor -v

docker run \
    -it \
    --rm \
    --mount type=bind,src="$(pwd)",dst="//app" \
    -w "//app" \
    --platform linux/arm/v6 \
    rpi-ws281x-builder-armv6 \
    go build -mod=vendor -v -o build/ledctld-linux-armv6 cmd/server/main.go
