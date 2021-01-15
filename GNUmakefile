GOOS := $(shell go env GOOS)
GOARCH := $(shell go env GOARCH)
TAG := 1.1.5.1
TF_PLUGIN_PATH := $(HOME)/.terraform.d/plugins/$(GOOS)_$(GOARCH)
PLUGIN_NAME := terraform-provider-pingdom

default: build

build: mod
	go build -o build/$(GOOS)_$(GOARCH)/$(PLUGIN_NAME)_v$(TAG)

install: build
	install -d $(TF_PLUGIN_PATH) && \
		install build/$(GOOS)_$(GOARCH)/$(PLUGIN_NAME)_v$(TAG) $(TF_PLUGIN_PATH)

lint:
	golangci-lint run

test:
	go test -v -cover ./...

clean:
	rm -rf build/

build-linux: mod
	@mkdir -p build/linux_amd64
	@docker build -t build .
	@docker run --detach --name build build
	@docker cp build:/app/$(PLUGIN_NAME) ./build/linux_amd64/$(PLUGIN_NAME)_v$(TAG)
	@docker rm -f build

create-release:
	rm -rf build/release
	mkdir -p build/release
	zip -j build/release/$(PLUGIN_NAME)_$(TAG)_linux_amd64.zip build/linux_amd64/$(PLUGIN_NAME)_v$(TAG)
	zip -j build/release/$(PLUGIN_NAME)_$(TAG)_darwin_amd64.zip build/darwin_amd64/$(PLUGIN_NAME)_v$(TAG)
	cd build/release && shasum -a 256 *.zip > $(PLUGIN_NAME)_$(TAG)_SHA256SUMS
	gpg --detach-sign build/release/$(PLUGIN_NAME)_$(TAG)_SHA256SUMS

mod:
	@go mod tidy
	@go mod vendor

.PHONY: build install lint test clean build-linux mod
