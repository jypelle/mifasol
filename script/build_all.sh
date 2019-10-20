#!/bin/bash
if [ -z "$1" ]; then
  echo "You should provide a version as parameter."
  exit
fi

echo "Build servers"
GOOS=windows GOARCH=amd64 go build -o assets/mifasolsrv-$1-windows-amd64.exe ../cmd/mifasolsrv
GOOS=linux GOARCH=amd64 go build -o assets/mifasolsrv-$1-linux-amd64 ../cmd/mifasolsrv
GOOS=linux GOARCH=arm go build -o assets/mifasolsrv-$1-linux-arm ../cmd/mifasolsrv
GOOS=darwin GOARCH=amd64 go build -o assets/mifasolsrv-$1-darwin-amd64 ../cmd/mifasolsrv
echo "Build clients"
GOOS=windows GOARCH=amd64 go build -o assets/mifasolcli-$1-windows-amd64.exe ../cmd/mifasolcli
GOOS=linux GOARCH=amd64 go build -o assets/mifasolcli-$1-linux-amd64 ../cmd/mifasolcli
GOOS=linux GOARCH=arm go build -o assets/mifasolcli-$1-linux-arm ../cmd/mifasolcli
#GOOS=darwin GOARCH=amd64 go build -o assets/mifasolcli-$1-darwin-amd64 ../cmd/mifasolcli
