#!/bin/bash

GOOS=windows GOARCH=amd64 go build -mod=mod -v -o build/ledctl-windows-amd64.exe cmd/client/main.go
GOOS=linux GOARCH=amd64 go build -mod=mod -v -o build/ledctl-linux-amd64.exe cmd/client/main.go
