#!/bin/bash
export GOOS=linux
export GOARCH=amd64
echo "Build"
go build -o assets/mifasolcli-linux-amd64 ../cmd/mifasolcli
