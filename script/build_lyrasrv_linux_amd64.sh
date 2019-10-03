#!/bin/bash
export GOOS=linux
export GOARCH=amd64
echo "Build"
go build lyra/cmd/lyrasrv
