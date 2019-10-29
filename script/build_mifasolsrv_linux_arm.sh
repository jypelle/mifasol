#!/bin/bash
export GOOS=linux
export GOARCH=arm
export GOARM=7
echo "Build"
go build -o assets/mifasolsrv-linux-arm ../cmd/mifasolsrv
