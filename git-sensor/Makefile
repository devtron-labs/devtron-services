
all: fetch-all-env build

TAG?=latest
FLAGS=
ENVVAR=
GOOS?=darwin
REGISTRY?=686244538589.dkr.ecr.us-east-2.amazonaws.com
BASEIMAGE?=alpine:3.9
GIT_COMMIT =$(shell sh -c 'git log --pretty=format:'%h' -n 1')
BUILD_TIME= $(shell sh -c 'date -u '+%Y-%m-%dT%H:%M:%SZ'')

include $(ENV_FILE)
export

build: clean wire
	$(ENVVAR) GOOS=$(GOOS) go build \
	    -o git-sensor \
		-ldflags="-X 'github.com/devtron-labs/git-sensor/util.GitCommit=${GIT_COMMIT}' -X 'github.com/devtron-labs/git-sensor/util.BuildTime=${BUILD_TIME}'"

wire:
	wire

clean:
	rm -rf git-sensor
	export GOFLAGS=-buildvcs=false

run: build
	./git-sensor

#.PHONY: build
#docker-build-image:  build
#	 docker build -t orchestrator:$(TAG) .

#.PHONY: build, all, wire, clean, run, set-docker-build-env, docker-build-push, orchestrator,
#docker-build-push: docker-build-image
#	docker tag orchestrator:${TAG}  ${REGISTRY}/orchestrator:${TAG}
#	docker push ${REGISTRY}/orchestrator:${TAG}

dep-update-oss:
	go mod edit -replace=github.com/devtron-labs/common-lib=github.com/devtron-labs/devtron-services/common-lib@$(TARGET_BRANCH)
	go mod tidy
	go mod vendor


fetch-all-env:
	go run fetchAllEnv/fetchAllEnv.go


