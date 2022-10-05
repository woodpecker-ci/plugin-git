GOFILES_NOVENDOR = $(shell find . -type f -name '*.go' -not -path "./vendor/*" -not -path "./.git/*")
GO_PACKAGES ?= $(shell go list ./... | grep -v /vendor/)

TARGETOS ?= linux
TARGETARCH ?= amd64

VERSION ?= next
ifneq ($(CI_COMMIT_TAG),)
	VERSION := $(CI_COMMIT_TAG:v%=%)
endif

# append commit-sha to next version
BUILD_VERSION := $(VERSION)
ifeq ($(BUILD_VERSION),next)
	CI_COMMIT_SHA ?= $(shell git rev-parse HEAD)
	BUILD_VERSION := $(shell echo "next-$(shell echo ${CI_COMMIT_SHA} | head -c 8)")
endif

LDFLAGS := -s -w -extldflags "-static" -X main.version=${BUILD_VERSION}

all: build

.PHONY: vendor
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
	# we can not use "-race" as test trigger write to os.stdout

build:
	CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} go build -ldflags '${LDFLAGS}' -o release/plugin-git

.PHONY: version
version:
	@echo ${BUILD_VERSION}

release-binaries:
	GOOS=linux   GOARCH=amd64 CGO_ENABLED=0 go build -ldflags '${LDFLAGS}' -o release/linux-amd64_plugin-git
	GOOS=linux   GOARCH=arm64 CGO_ENABLED=0 go build -ldflags '${LDFLAGS}' -o release/linux-arm64_plugin-git
	GOOS=linux   GOARCH=arm   CGO_ENABLED=0 go build -ldflags '${LDFLAGS}' -o release/linux-arm_plugin-git
	GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build -ldflags '${LDFLAGS}' -o release/windows-amd64_plugin-git.exe
	GOOS=windows GOARCH=arm64 CGO_ENABLED=0 go build -ldflags '${LDFLAGS}' -o release/windows-arm64_plugin-git.exe
	GOOS=darwin  GOARCH=amd64 CGO_ENABLED=0 go build -ldflags '${LDFLAGS}' -o release/darwin-amd64_plugin-git
	GOOS=darwin  GOARCH=arm64 CGO_ENABLED=0 go build -ldflags '${LDFLAGS}' -o release/darwin-arm64_plugin-git

release-tarball:
	mkdir -p release
	tar -cvzf release/plugin-git-src-$(BUILD_VERSION).tar.gz \
	  *.go \
	  go.??? \
	  LICENSE \
	  Makefile

release-checksums:
	# generate shas for tar files
	(cd release/; sha256sum *plugin-git* > checksums.txt)

.PHONY: release
release: release-binaries release-tarball
	$(MAKE) release-checksums
