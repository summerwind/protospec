VERSION ?= dev
COMMIT  ?= $(shell git rev-parse --verify HEAD | cut -c 1-7)

IMAGE_NAME  = ghcr.io/summerwind/protospec
BUILD_FLAGS = -ldflags "-X main._version=$(VERSION) -X main._commit=$(COMMIT)"

all: build

build:
	go build $(BUILD_FLAGS) .

test:
	go vet ./...
	go test -v ./...

release:
	mkdir -p release
	GOARCH=amd64 GOOS=darwin go build $(BUILD_FLAGS) -o protospec .
	tar -czf release/protospec_darwin_amd64.tar.gz protospec
	GOARCH=arm64 GOOS=darwin go build $(BUILD_FLAGS) -o protospec .
	tar -czf release/protospec_darwin_arm64.tar.gz protospec
	GOARCH=amd64 GOOS=linux go build $(BUILD_FLAGS) -o protospec .
	tar -czf release/protospec_linux_amd64.tar.gz protospec
	GOARCH=arm64 GOOS=linux go build $(BUILD_FLAGS) -o protospec .
	tar -czf release/protospec_linux_arm64.tar.gz protospec
	rm -rf protospec

latest-image:
	docker buildx build \
		--platform linux/amd64,linux/arm64 \
		--build-arg VERSION=$(VERSION) \
		--build-arg COMMIT=$(COMMIT) \
		-t $(IMAGE_NAME):latest --push .

release-image:
	docker buildx build \
		--platform linux/amd64,linux/arm64 \
		--build-arg VERSION=$(VERSION) \
		--build-arg COMMIT=$(COMMIT) \
		-t $(IMAGE_NAME):$(VERSION) --push .

clean:
	rm -rf protospec release
