# Change Log

All notable changes to the Docker Language Server will be documented in this file.

## [Unreleased]

### Added

- Compose
  - textDocument/completion
    - add code completion support based on the JSON schema, extracting out attribute names and enum values ([#86](https://github.com/docker/docker-language-server/issues/86))
    - completion items are populated with a detail that corresponds to the possible types of the item ([#93](https://github.com/docker/docker-language-server/issues/93))
    - suggests completion items for the attributes of an object inside an array ([#95](https://github.com/docker/docker-language-server/issues/95))
  - textDocument/definition
    - support lookup of `configs`, `networks`, and `secrets` referenced inside `services` object ([#91](https://github.com/docker/docker-language-server/issues/91))
  - textDocument/documentLink
    - support opening a referenced image's page as a link ([#91](https://github.com/docker/docker-language-server/issues/91))

## [0.3.8] - 2025-04-24

### Added
- Bake
  - textDocument/definition
    - support code navigation when a single attribute of a target has been reused ([#78](https://github.com/docker/docker-language-server/issues/78))
  - textDocument/semanticTokens/full
    - ensure only Bake files will respond to a textDocument/semanticTokens/full request ([#84](https://github.com/docker/docker-language-server/issues/84))
- Compose
  - textDocument/definition
    - support lookup of `services` referenced by the short form syntax of `depends_on` ([#67](https://github.com/docker/docker-language-server/issues/67))
    - support lookup of `services` referenced by the long form syntax of `depends_on` ([#68](https://github.com/docker/docker-language-server/issues/68))

### Fixed
- ensure file validation is skipped if the file has since been closed by the editor ([#79](https://github.com/docker/docker-language-server/issues/79))

## [0.3.7] - 2025-04-21

### Fixed
- ensure file validation is skipped if the file has since been closed by the editor ([#79](https://github.com/docker/docker-language-server/issues/79))

## [0.3.6] - 2025-04-18

### Changed
- get the JSON structure of a Bake target with Go APIs instead of spawning a separate child process ([#63](https://github.com/docker/docker-language-server/issues/63))
- Update `moby/buildkit` to v0.21.0 and `docker/buildx` to v0.23.0 ([#64](https://github.com/docker/docker-language-server/issues/64))

### Fixed

- Bake
  - textDocument/publishDiagnostics
    - consider the context attribute when determining which Dockerfile the Bake target is for ([#57](https://github.com/docker/docker-language-server/issues/57))
  - textDocument/inlayHints
    - consider the context attribute when determining which Dockerfile to use for inlaying default values of `ARG` variables ([#60](https://github.com/docker/docker-language-server/pull/60))
  - textDocument/completion
    - consider the context attribute when determining which Dockerfile to use for looking up build stages ([#61](https://github.com/docker/docker-language-server/pull/61))
  - textDocument/definition
    - consider the context attribute when trying to resolve the Dockerfile to use for `ARG` variable definitions ([#62](https://github.com/docker/docker-language-server/pull/62))
    - fix a panic that may occur if a for loop did not have a conditional expression ([#65](https://github.com/docker/docker-language-server/pull/65))

## [0.3.5] - 2025-04-13

### Fixed

- initialize
  - when responding to the initialize request, we should send an empty array back for tokenModifiers instead of a null

## [0.3.4] - 2025-04-11

### Fixed

- Compose
  - textDocument/documentSymbol
    - prevent scalar values from showing up as a document symbol

## [0.3.3] - 2025-04-09

### Fixed

- refactored the panic handler so that crashes from handling the JSON-RPC messages would no longer cause the language server to crash

## [0.3.2] - 2025-04-09

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

[Unreleased]: https://github.com/docker/docker-language-server/compare/v0.3.8...main
[0.3.8]: https://github.com/docker/docker-language-server/compare/v0.3.7...v0.3.8
[0.3.7]: https://github.com/docker/docker-language-server/compare/v0.3.6...v0.3.7
[0.3.6]: https://github.com/docker/docker-language-server/compare/v0.3.5...v0.3.6
[0.3.5]: https://github.com/docker/docker-language-server/compare/v0.3.4...v0.3.5
[0.3.4]: https://github.com/docker/docker-language-server/compare/v0.3.3...v0.3.4
[0.3.3]: https://github.com/docker/docker-language-server/compare/v0.3.2...v0.3.3
[0.3.2]: https://github.com/docker/docker-language-server/compare/v0.3.1...v0.3.2
[0.3.1]: https://github.com/docker/docker-language-server/compare/v0.3.0...v0.3.1
[0.3.0]: https://github.com/docker/docker-language-server/compare/v0.2.0...v0.3.0
[0.2.0]: https://github.com/docker/docker-language-server/compare/v0.1.0...v0.2.0
