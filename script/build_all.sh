#!/bin/bash
echo "Build server"
GOOS=windows GOARCH=amd64 go build -o assets/mifasolsrv-windows-amd64.exe mifasol/cmd/mifasolsrv
GOOS=linux GOARCH=amd64 go build -o assets/mifasolsrv-linux-amd64 mifasol/cmd/mifasolsrv
GOOS=linux GOARCH=arm go build -o assets/mifasolsrv-linux-arm mifasol/cmd/mifasolsrv
GOOS=darwin GOARCH=amd64 go build -o assets/mifasolsrv-darwin-amd64 mifasol/cmd/mifasolsrv
echo "Build client"
GOOS=windows GOARCH=amd64 go build -o assets/mifasolcli-windows-amd64.exe mifasol/cmd/mifasolcli
GOOS=linux GOARCH=amd64 go build -o assets/mifasolcli-linux-amd64 mifasol/cmd/mifasolcli
GOOS=linux GOARCH=arm go build -o assets/mifasolcli-linux-arm mifasol/cmd/mifasolcli
#GOOS=darwin GOARCH=amd64 go build -o assets/mifasolcli-darwin-amd64 mifasol/cmd/mifasolcli
