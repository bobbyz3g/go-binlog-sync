all: clean build

build:
	GOOS=linux GOARCH=arm64 go build -o bin/gbs-arm64 ./cmd/gbs
	GOOS=linux GOARCH=amd64 go build -o bin/gbs-amd64 ./cmd/gbs

clean:
	rm -rf bin/

test:
	go test -v ./...