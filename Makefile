APP     := openclaw-honeypot
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
REGISTRY := ghcr.io/shart-cloud
IMAGE   := $(REGISTRY)/$(APP)

.PHONY: build test fmt vet tidy deps docker-build docker-push ci clean

build:
	CGO_ENABLED=0 go build -ldflags="-s -w -X main.version=$(VERSION)" -o honeypot ./cmd/honeypot

test:
	go test -v -race ./...

fmt:
	gofmt -w .

vet:
	go vet ./...

tidy:
	go mod tidy

deps:
	go mod download

docker-build:
	docker build -t $(IMAGE):$(VERSION) -t $(IMAGE):latest .

docker-push: docker-build
	docker push $(IMAGE):$(VERSION)
	docker push $(IMAGE):latest

ci: deps fmt vet test build

clean:
	rm -f honeypot
