FROM golang:1.23.8-alpine3.21 AS base
WORKDIR /build
RUN apk update && apk upgrade && apk add --no-cache ca-certificates go make && update-ca-certificates
COPY . .

FROM base AS test
RUN <<SETUP
    apk add --no-cache make docker-cli && go install gotest.tools/gotestsum@latest
    wget https://github.com/docker/buildx/releases/download/v0.21.2/buildx-v0.21.2.linux-amd64
    chmod +x buildx-v0.21.2.linux-amd64
    mkdir ~/.docker/ ~/.docker/cli-plugins/
    mv buildx-v0.21.2.linux-amd64 ~/.docker/cli-plugins/docker-buildx
SETUP
# use -run=NON_EXISTING to download all the test dependencies
RUN make install && go test -run=NON_EXISTING ./...
