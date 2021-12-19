GOFILES_NOVENDOR = $(shell find . -type f -name '*.go' -not -path "./vendor/*" -not -path "./.git/*")
GO_PACKAGES ?= $(shell go list ./... | grep -v /vendor/)

TARGETOS ?= linux
TARGETARCH ?= amd64

VERSION ?= next
ifneq ($(CI_TAG),)
	VERSION := $(CI_TAG:v%=%)
endif

# append commit-sha to next version
BUILD_VERSION := $(VERSION)
ifeq ($(BUILD_VERSION),next)
	BUILD_VERSION := $(shell echo "next-$(shell echo ${CI_COMMIT_SHA} | head -c 8)")
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
	CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} go build -ldflags '${LDFLAGS}' -o release/plugin-git

.PHONY: version
version:
	@echo ${VERSION}
