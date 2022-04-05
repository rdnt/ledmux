#!/bin/bash

docker buildx build \
    --cache-to=type=inline,mode=max \
    --platform linux/arm64 \
    --tag rpi-ws281x-builder-arm64 \
    --load \
    --file builder/Dockerfile .
