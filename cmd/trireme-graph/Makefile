VERSION_FILE := ./version/version.go
PROJECT_NAME := trireme-statistics
BUILD_NUMBER := latest
VERSION := 0.11
REVISION=$(shell git log -1 --pretty=format:"%H")
DOCKER_REGISTRY?=aporeto
DOCKER_IMAGE_NAME?=$(PROJECT_NAME)
DOCKER_IMAGE_TAG?=$(BUILD_NUMBER)

build:
	CGO_ENABLED=1 go build -o trireme-statistics

package: build
	mv trireme-statistics docker/trireme-statistics

bindata:
	cd graph; go-bindata -pkg=graph html/...

clean:
	rm -rf ./vendor
	rm -rf ./docker/collector

docker_build: package
		docker \
			build \
			-t $(DOCKER_REGISTRY)/$(DOCKER_IMAGE_NAME):$(DOCKER_IMAGE_TAG) docker

docker_push: docker_build
		docker \
			push \
			$(DOCKER_REGISTRY)/$(DOCKER_IMAGE_NAME):$(DOCKER_IMAGE_TAG)
