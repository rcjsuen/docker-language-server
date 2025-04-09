# Change Log

All notable changes to the Docker Language Server will be documented in this file.

## [Unreleased]

### Fixed

- Docker Bake
  - textDocument/semanticTokens/full
    - prevent single line comments from crashing the server

## [0.3.1] - 2025-04-09

### Changed

- binaries are now built with `CGO_ENABLED=0` to allow for greater interoperability

## [0.3.0] - 2025-04-08

### Fixed

- textDocument/publishDiagnostics
  - stop diagnostics from being sent to the client if a file with errors or warnings were opened by the client and then quickly closed

## [0.2.0] - 2025-04-03

### Added

- Dockerfile
  - textDocument/publishDiagnostics
    - introduce a setting to ignore certain diagnostics to not duplicate the ones from the Dockerfile Language Server

- Docker Bake
  - textDocument/completion
    - suggest network attributes when the text cursor is inside of a string

- telemetry
  - records the language identifier of modified files, this will only include Dockerfiles, Bake files, and Compose files

### Fixed

- Docker Bake
  - textDocument/definition
    - always return LocationLinks to help disambiguate word boundaries for clients ([#31](https://github.com/docker/docker-language-server/issues/31))

## 0.1.0 - 2025-03-31

### Added

- Dockerfile
  - textDocument/codeAction
    - suggest remediation actions for build warnings
  - textDocument/hover
    - provide vulnerability information of referenced images (experimental)
  - textDocument/publishDiagnostics
    - report build check warnings from BuildKit and BuildX
    - scan images for vulnerabilities (experimental)
- Compose
  - textDocument/documentLink
    - allow jumping between included files
  - textDocument/documentSymbol
    - provide a document outline for Compose files
- Docker Bake
  - textDocument/codeAction
    - provide remediation actions for some detected errors
  - textDocument/codeLens
    - relays information to the client to run Bake on a specific target
  - textDocument/completion
    - code completion of block and attribute names
  - textDocument/definition
    - code navigation between variables, referenced targets, and referenced build stages
  - textDocument/documentHighlight
    - highlights the same variable or target references
  - textDocument/documentLink
    - jump from the Bake file to the Dockerfile
  - textDocument/documentSymbol
    - provide an outline for Bake files
  - textDocument/formatting
    - provide rudimentary support for formatting
  - textDocument/hover
    - show variable values
  - textDocument/inlayHint
    - inlay ARG values from the Dockerfile
  - textDocument/inlineCompletion
    - scans build stages from the Dockerfile and suggests target blocks
  - textDocument/publishDiagnostics
    - scan and report the Bake file for errors
  - textDocument/semanticTokens/full
    - provide syntax highlighting for Bake files

[Unreleased]: https://github.com/docker/docker-language-server/compare/v0.3.0...main
[0.3.0]: https://github.com/docker/docker-language-server/compare/v0.2.0...v0.3.0
[0.2.0]: https://github.com/docker/docker-language-server/compare/v0.1.0...v0.2.0
