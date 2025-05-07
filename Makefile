all: linux-amd64 linux-arm64

linux-amd64:
	GOOS=linux GOARCH=amd64 go build -o bin/gbb-linux-amd64 ./cmd/gbb
	
linux-arm64:
	GOOS=linux GOARCH=arm64 go build -o bin/gbb-linux-arm64 ./cmd/gbb
