#!/bin/bash

docker buildx build \
    --cache-to=type=inline,mode=max \
    --platform linux/arm/v6 \
    --tag rpi-ws281x-builder-armv6 \
    --load \
    --file builder/Dockerfile .
