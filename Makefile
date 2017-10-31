VERSION := 0.11
VERSION_FILE := ./version/version.go
REVISION=$(shell git log -1 --pretty=format:"%H")

codegen:
	echo 'package version' > $(VERSION_FILE)
	echo '' >> $(VERSION_FILE)
	echo '// VERSION is the version of trireme-statistics' >> $(VERSION_FILE)
	echo 'const VERSION = "$(VERSION)"' >> $(VERSION_FILE)
	echo '' >> $(VERSION_FILE)
	echo '// REVISION is the revision of trireme-statistics' >> $(VERSION_FILE)
	echo 'const REVISION = "$(REVISION)"' >> $(VERSION_FILE)

build: codegen
	cd cmd/grafana-init && make build
	cd ../..
	cd cmd/trireme-graph && make build

clean:
	rm -rf vendor

docker_build:
	cd cmd/grafana-init && make docker_build
	cd ../..
	cd cmd/trireme-graph && make docker_build

docker_push:
	cd cmd/grafana-init && make docker_push
	cd ../..
	cd cmd/trireme-graph && make docker_push