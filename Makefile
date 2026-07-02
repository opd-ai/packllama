BINARY ?= packllama
DIST ?= dist

.PHONY: build test vet run release clean

build:
	mkdir -p bin
	go build -o bin/$(BINARY) ./cmd/packllama

test:
	go test -race ./...

vet:
	go vet ./...

run:
	go run ./cmd/packllama

release: clean
	mkdir -p $(DIST)
	GOOS=linux GOARCH=amd64 go build -o $(DIST)/$(BINARY)-linux-amd64 ./cmd/packllama
	GOOS=linux GOARCH=arm64 go build -o $(DIST)/$(BINARY)-linux-arm64 ./cmd/packllama
	GOOS=darwin GOARCH=amd64 go build -o $(DIST)/$(BINARY)-darwin-amd64 ./cmd/packllama
	GOOS=darwin GOARCH=arm64 go build -o $(DIST)/$(BINARY)-darwin-arm64 ./cmd/packllama
	GOOS=windows GOARCH=amd64 go build -o $(DIST)/$(BINARY)-windows-amd64.exe ./cmd/packllama
	GOOS=windows GOARCH=arm64 go build -o $(DIST)/$(BINARY)-windows-arm64.exe ./cmd/packllama

clean:
	rm -rf bin $(DIST)
