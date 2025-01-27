
all: fetch-all-env build

TAG?=latest
FLAGS=
ENVVAR=
GOOS?=darwin
REGISTRY?=686244538589.dkr.ecr.us-east-2.amazonaws.com
BASEIMAGE?=alpine:3.9
#BUILD_NUMBER=$$(date +'%Y%m%d-%H%M%S')
#BUILD_NUMBER := $(shell bash -c 'echo $$(date +'%Y%m%d-%H%M%S')')
include $(ENV_FILE)
export

build: clean wire
	$(ENVVAR) GOOS=$(GOOS) go build -o image-scanner

wire:
	wire

clean:
	rm -rf image-scanner

run: build
	./image-scanner

.PHONY: build
docker-build-image:  build
	 docker build -t image-scanner:$(TAG) .

dep-update-oss:
	go mod edit -replace=github.com/devtron-labs/common-lib=github.com/devtron-labs/devtron-services/common-lib@$(TARGET_BRANCH)
	go mod tidy
	go mod vendor

fetch-all-env:
	go run fetchAllEnv/fetchAllEnv.go
