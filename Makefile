.PHONY: release runcliui buildwa

release:
	echo "Build windows amd64 client"; \
	GOOS=windows GOARCH=amd64 go build -o internal/srv/webSrv/clients/mifasolcli-windows-amd64.exe ./cmd/mifasolcli; \
	echo "Build linux amd64 client"; \
	GOOS=linux GOARCH=amd64 go build -o internal/srv/webSrv/clients/mifasolcli-linux-amd64 ./cmd/mifasolcli; \
	echo "Build linux arm client"; \
	GOOS=linux GOARCH=arm GOARM=7 go build -o internal/srv/webSrv/clients/mifasolcli-linux-arm ./cmd/mifasolcli; \
#	echo "Build darwin amd64 client"; \
#	GOOS=darwin GOARCH=amd64 go build -o internal/srv/webSrv/clients/mifasolcli-darwin-amd64 ./cmd/mifasolcli; \
	echo "Build webassembly client"; \
	GOOS=js GOARCH=wasm go build -o internal/srv/webSrv/clients/mifasolwa.wasm ./cmd/mifasolwa; \
	echo "Build windows amd64 server"; \
	GOOS=windows GOARCH=amd64 go build -o release/mifasolsrv-windows-amd64.exe ./cmd/mifasolsrv; \
	echo "Build linux amd64 server"; \
	GOOS=linux GOARCH=amd64 go build -o release/mifasolsrv-linux-amd64 ./cmd/mifasolsrv; \
	echo "Build linux arm server"; \
	GOOS=linux GOARCH=arm GOARM=7 go build -o release/mifasolsrv-linux-arm ./cmd/mifasolsrv; \
	echo "Build darwin amd64 server"; \
	GOOS=darwin GOARCH=amd64 go build -o release/mifasolsrv-darwin-amd64 ./cmd/mifasolsrv; \
	echo "Build darwin arm64 server"; \
	GOOS=darwin GOARCH=arm64 go build -o release/mifasolsrv-darwin-arm64 ./cmd/mifasolsrv;

runcliui:
	go run ./cmd/mifasolcli ui;

buildcliwa:
	GOOS=js GOARCH=wasm go build -o internal/srv/webSrv/clients/mifasolcliwa.wasm ./cmd/mifasolcliwa;
