#!/bin/bash
echo "Build server"
GOOS=windows GOARCH=amd64 go build -o assets/lyrasrv-windows-amd64.exe lyra/cmd/lyrasrv
GOOS=linux GOARCH=amd64 go build -o assets/lyrasrv-linux-amd64 lyra/cmd/lyrasrv
GOOS=linux GOARCH=arm go build -o assets/lyrasrv-linux-arm lyra/cmd/lyrasrv
GOOS=darwin GOARCH=amd64 go build -o assets/lyrasrv-darwin-amd64 lyra/cmd/lyrasrv
echo "Build client"
GOOS=windows GOARCH=amd64 go build -o assets/lyracli-windows-amd64.exe lyra/cmd/lyracli
GOOS=linux GOARCH=amd64 go build -o assets/lyracli-linux-amd64 lyra/cmd/lyracli
GOOS=linux GOARCH=arm go build -o assets/lyracli-linux-arm lyra/cmd/lyracli
#GOOS=darwin GOARCH=amd64 go build -o assets/lyracli-darwin-amd64 lyra/cmd/lyracli
