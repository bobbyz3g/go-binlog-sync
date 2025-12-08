all: linux-amd64 linux-arm64

linux-amd64:
	GOOS=linux GOARCH=amd64 go build -o bin/gbb-linux-amd64 ./cmd/gbb
	GOOS=linux GOARCH=amd64 go build -o bin/gbb-agent-linux-amd64 ./cmd/gbb-agent
	
linux-arm64:
	GOOS=linux GOARCH=arm64 go build -o bin/gbb-linux-arm64 ./cmd/gbb
	GOOS=linux GOARCH=arm64 go build -o bin/gbb-agent-linux-arm64 ./cmd/gbb-agent

clean:
	rm -rf bin/

test:
	go test -v ./...