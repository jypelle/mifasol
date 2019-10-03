#!/bin/bash
export GOOS=linux
export GOARCH=arm
echo "Build"
go build lyra/cmd/lyracli
