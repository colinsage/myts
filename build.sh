#!/bin/bash

echo "Begin build binary..."
GOOS=linux GOARCH=amd64 go build -o _build/linux_amd64/myts cmd/node/main.go
GOOS=linux GOARCH=arm64 go build -o _build/linux_arm64/myts cmd/node/main.go
GOOS=darwin GOARCH=amd64 go build -o _build/darwin_arm64/myts cmd/node/main.go

echo "Build Done!"