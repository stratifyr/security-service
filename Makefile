.PHONY: build vet lint run verify

build:
	go build ./...

vet:
	go vet ./...

lint:
	golangci-lint run ./...

run:
	go run .

verify: build vet lint
