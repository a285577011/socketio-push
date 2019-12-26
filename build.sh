#!/bin/sh
go build -ldflags "-s -w" -o ./bin/pushservice ./cmd/main.go
