# Docker Language Server

The Docker Language Server is a [language server](https://microsoft.github.io/language-server-protocol/) for providing language features for file types in the Docker ecosystem ([Dockerfiles](https://docs.docker.com/reference/dockerfile/), [Compose files](https://docs.docker.com/reference/compose-file/), and [Bake files](https://docs.docker.com/build/bake/reference/)).

## Requirements

The Docker Language Server relies on some features that are dependent on [Buildx](https://github.com/docker/buildx). If Buildx is not available as a Docker CLI plugin then those features will not be available.

## Features

- Dockerfile
  - hover support for images to show vulnerability information from Docker Scout
  - suggested image tag updates from Docker Scout
  - Dockerfile linting support from BuildKit and Buildx
- Compose files
  - document outline support
- Bake files
  - code completion
  - inferring variable values
  - formatting
  - code navigation
  - hover tooltips

## Installing

Ensure you have Go 1.23 or greater installed, check out this repository and then run `make install`.

## Building

Run `make build` to generate a binary of the Docker Language Server.

## Testing

Run `make test` to run the unit tests. If the BuildKit tests do not work, make sure that `docker buildx build` works from the command line.

If you would like to run the tests inside of a Docker container, run `make test-docker` which will build a Docker image with the test code and then execute the tests from within the Docker container. Note that this requires the Docker daemon's UNIX socket to be mounted.

## CLI

The main command for docker-language-server is `docker-language-server start` with `--stdio` or `--address :12345`:

When run in stdio mode, requests and responses will be written to stdin and stdout. All logging is _always_ written to stderr.

```
Language server for Docker

Usage:
  docker-language-server [command]

Available Commands:
  completion  Generate the autocompletion script for the specified shell
  help        Help about any command
  start       Start the Docker LSP server

Flags:
      --debug     Enable debug logging
  -h, --help      help for docker-language-server
      --verbose   Enable verbose logging
  -v, --version   version for docker-language-server

Use "docker-language-server [command] --help" for more information about a command.
```

## Configuration

### Telemetry

The Docker Language Server has telemetry which is **disabled** by default. Telemetry is only set on startup with the first incoming `initialized` request from the client. Please read our [privacy policy](https://www.docker.com/legal/docker-privacy-policy/) to learn more about how your data will be collected and used.

```JSONC
{
  "clientInfo": {
      "name": "clientName",
      "version": "1.2.3"
  },
  "initializationOptions": {
    // you can send enable all telemetry, only send errors, or disable it completely
    "telemetry": "all" | "error" | "off"
  }
}
```

### Experimental Capabilities

To support `textDocument/codeLens`, the client must provide a command with the id `dockerLspClient.bake.build` for executing the build. If this is supported, the client can define its experimental capabilities as follows. The server will then respond that it supports code lens requests and return results for `textDocument/codeLens` requests for Bake HCL files.

```JSONC
{
  "capabilities": {
    "experiemntal:": {
      "dockerLanguageServerCapabilities": {
          "commands": [
            "dockerLspClient.bake.build"
          ]
      }
    }
  }
}
```

## tliron/glsp Fork

This project includes a fork of the [tliron/glsp GitHub repository](https://github.com/tliron/glsp/). The code can be found inside the `internal/tliron/glsp` folder.
