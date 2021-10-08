GOFILES_NOVENDOR = $(shell find . -type f -name '*.go' -not -path "./vendor/*" -not -path "./.git/*")
GO_PACKAGES ?= $(shell go list ./... | grep -v /vendor/)

VERSION ?= next
ifneq ($(DRONE_TAG),)
	VERSION := $(DRONE_TAG:v%=%)
endif

# append commit-sha to next version
BUILD_VERSION := $(VERSION)
ifeq ($(BUILD_VERSION),next)
	BUILD_VERSION := $(shell echo "next-$(shell echo ${DRONE_COMMIT_SHA} | head -c 8)")
endif

LDFLAGS := -s -w -extldflags "-static" -X main.version=${BUILD_VERSION}

all: build

vendor:
	go mod tidy
	go mod vendor

formatcheck:
	@([ -z "$(shell gofmt -d $(GOFILES_NOVENDOR) | head)" ]) || (echo "Source is unformatted"; exit 1)

format:
	@gofmt -w ${GOFILES_NOVENDOR}

.PHONY: clean
clean:
	go clean -i ./...
	rm -rf release/

.PHONY: vet
vet:
	@echo "Running go vet..."
	@go vet $(GO_PACKAGES)

test:
	go test -cover ./...
# TODO: add '-race'

build:
	GOOS=linux   GOARCH=amd64 CGO_ENABLED=0 go build -ldflags '${LDFLAGS}' -o release/linux/amd64/plugin-git
	GOOS=linux   GOARCH=arm64 CGO_ENABLED=0 go build -ldflags '${LDFLAGS}' -o release/linux/arm64/plugin-git
	GOOS=linux   GOARCH=arm   CGO_ENABLED=0 go build -ldflags '${LDFLAGS}' -o release/linux/arm/plugin-git
	GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build -ldflags '${LDFLAGS}' -o release/windows/amd64/plugin-git
	GOOS=darwin  GOARCH=amd64 CGO_ENABLED=0 go build -ldflags '${LDFLAGS}' -o release/darwin/amd64/plugin-git
	GOOS=darwin  GOARCH=arm64 CGO_ENABLED=0 go build -ldflags '${LDFLAGS}' -o release/darwin/arm64/plugin-git

.PHONY: version
version:
	@echo ${VERSION}
