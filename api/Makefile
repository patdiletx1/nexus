DOCKER_GO_IMAGE ?= golang:1.23-alpine
WORKDIR ?= /workspace/api

.PHONY: docker-go-version docker-go-tidy docker-go-test

docker-go-version:
	docker run --rm $(DOCKER_GO_IMAGE) /usr/local/go/bin/go version

docker-go-tidy:
	docker run --rm -v "$(PWD)/..:/workspace" -w $(WORKDIR) $(DOCKER_GO_IMAGE) sh -lc "/usr/local/go/bin/go mod tidy"

docker-go-test:
	docker run --rm -v "$(PWD)/..:/workspace" -w $(WORKDIR) $(DOCKER_GO_IMAGE) sh -lc "/usr/local/go/bin/go test ./..."
