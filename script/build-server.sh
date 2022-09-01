#!/bin/bash

go mod vendor -v

docker run \
    -it \
    --rm \
    --mount type=bind,src="$(pwd)",dst="//app" \
    -w "//app" \
    --platform linux/arm64 \
    rpi-ws281x-builder-arm64 \
    go build -v -o build/ledctld-linux-arm64 cmd/server/main.go

rm -rf ./vendor
