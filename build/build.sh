#!/usr/bin/env bash
set -xe
GOOS=linux GOARCH=amd64 go build -o ./bin/stoilo_streams ./cmd/main.go
