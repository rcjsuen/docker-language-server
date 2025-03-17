.PHONY: test lint tidy install

build:
	./build.sh

test:
	gotestsum -- $$(go list ./... | grep -v e2e-tests) -timeout 30s
	go test $$(go list ./... | grep e2e-tests) -timeout 120s

build-docker-test:
	docker build -t docker/lsp:test --target test .

test-docker: build-docker-test
	docker run --rm -v /var/run/docker.sock:/var/run/docker.sock docker/lsp:test make test

test-docker-disconnected: build-docker-test
	docker run -e DOCKER_NETWORK_NONE=true --rm -v /var/run/docker.sock:/var/run/docker.sock --network none docker/lsp:test make test

lint:
	golangci-lint run --exclude-dirs lsp

tidy:
	go mod tidy

install:
	go install ./cmd/docker-language-server
