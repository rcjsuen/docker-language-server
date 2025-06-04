# Change Log

All notable changes to the Docker Language Server will be documented in this file.

## [Unreleased]

### Fixed

- Compose
  - textDocument/completion
    - fix incorrect snippet item that was generated even if there were no choices to suggest ([#283](https://github.com/docker/docker-language-server/issues/283))

## [0.10.0] - 2025-06-03

### Added

- errors will now be reported to BugSnag if telemetry is not disabled
- Compose
  - textDocument/definition
    - support navigating to the defined YAML anchor from an alias reference ([#264](https://github.com/docker/docker-language-server/issues/264))
  - textDocument/documentHighlight
    - support highlighting YAML anchor and alias references ([#264](https://github.com/docker/docker-language-server/issues/264))
  - textDocument/documentLink
    - support opening a referenced Dockerfile from the `build` object's `dockerfile` attribute ([#69](https://github.com/docker/docker-language-server/issues/69))
    - support opening a referenced file from a config's `file` attribute ([#271](https://github.com/docker/docker-language-server/issues/271))
    - support opening a referenced file from a secret's `file` attribute ([#272](https://github.com/docker/docker-language-server/issues/272))
    - provide document links when an included file is also a YAML anchor ([#275](https://github.com/docker/docker-language-server/issues/275))
  - textDocument/hover
    - render the referenced network's YAML content as a hover result ([#246](https://github.com/docker/docker-language-server/issues/246))
    - render the referenced config's YAML content as a hover result ([#249](https://github.com/docker/docker-language-server/issues/249))
    - render the referenced secret's YAML content as a hover result ([#250](https://github.com/docker/docker-language-server/issues/250))
    - render the referenced volume's YAML content as a hover result ([#251](https://github.com/docker/docker-language-server/issues/251))
    - include the range of the hovered element to clearly identify what is being hovered over for the client ([#256](https://github.com/docker/docker-language-server/issues/256))
    - render the referenced anchor's YAML content as a hover result ([#268](https://github.com/docker/docker-language-server/issues/268))
  - textDocument/prepareRename
    - support renaming YAML anchor and alias references ([#264](https://github.com/docker/docker-language-server/issues/264))
  - textDocument/rename
    - preparing rename operations for YAML anchor and alias references ([#264](https://github.com/docker/docker-language-server/issues/264))

### Fixed

- Compose
  - textDocument/completion
    - include the array definition in the inserted text so we do not make the YAML content malformed ([#278](https://github.com/docker/docker-language-server/issues/278))
  - textDocument/definition
    - fix range calculation when the element is quoted ([#255](https://github.com/docker/docker-language-server/issues/255))
  - textDocument/documentHighlight
    - fix range calculation when the element is quoted ([#255](https://github.com/docker/docker-language-server/issues/255))
  - textDocument/documentLink
    - consider quotes when calculating the link's range ([#242](https://github.com/docker/docker-language-server/issues/242))
    - consider anchors and aliases instead of assuming everything are strings ([#266](https://github.com/docker/docker-language-server/issues/266))
  - textDocument/hover
    - prevent YAML hover issues caused by whitespace ([#244](https://github.com/docker/docker-language-server/issues/244))
    - ignore hover requests that are outside the file to prevent panics ([#261](https://github.com/docker/docker-language-server/issues/261))
  - textDocument/prepareRename
    - fix range calculation when the element is quoted ([#255](https://github.com/docker/docker-language-server/issues/255))
  - textDocument/rename
    - fix range calculation when the element is quoted ([#255](https://github.com/docker/docker-language-server/issues/255))
- Bake
  - textDocument/publishDiagnostics
    - filter out variables when resolving Dockerfile paths to prevent false positives from being reported ([#263](https://github.com/docker/docker-language-server/issues/263))

## [0.9.0] - 2025-05-26

## Added

- global initialization option to disable all Compose features ([#230](https://github.com/docker/docker-language-server/issues/230))
- Compose
  - textDocument/completion
    - include the attribute's schema description when providing enum suggestions ([#235](https://github.com/docker/docker-language-server/issues/235))

### Changed

- Dockerfile
  - textDocument/hover
    - `recommended_tag` diagnostics are now hidden by default ([#223](https://github.com/docker/docker-language-server/issues/223))
  - textDocument/publishDiagnostics
    - recommended tag hovers are now hidden by default ([#223](https://github.com/docker/docker-language-server/issues/223))

### Fixed

- correct initialize request handling for clients that do not support dynamic registrations ([#229](https://github.com/docker/docker-language-server/issues/229))
- Dockerfile
  - textDocument/hover
    - hide vulnerability hovers if the top level setting is disabled ([#226](https://github.com/docker/docker-language-server/issues/226))
  - textDocument/publishDiagnostics
    - consider flag changes when determining whether to scan a file again or not ([#224](https://github.com/docker/docker-language-server/issues/224))
- Compose
  - textDocument/hover
    - fixed a case where an object reference's description would not be returned in a hover result ([#233](https://github.com/docker/docker-language-server/issues/233))

## [0.8.0] - 2025-05-23

### Added

- Dockerfile
  - textDocument/hover
    - support configuring vulnerability hovers with an experimental setting ([#192](https://github.com/docker/docker-language-server/issues/192))
  - textDocument/publishDiagnostics
    - support filtering vulnerability diagnostics with an experimental setting ([#192](https://github.com/docker/docker-language-server/issues/192))
- Compose
  - updated Compose schema to the latest version
  - textDocument/definition
    - support navigating to a dependency that is defined in another file ([#190](https://github.com/docker/docker-language-server/issues/190))
  - textDocument/hover
    - improve hover result by linking to the schema and the online documentation ([#199](https://github.com/docker/docker-language-server/issues/199))
    - add support for hovering over service names that are defined in a different file ([#207](https://github.com/docker/docker-language-server/issues/207))
- Bake
  - textDocument/publishDiagnostics
    - support filtering vulnerability diagnostics with an experimental setting ([#192](https://github.com/docker/docker-language-server/issues/192))

### Changed

- Dockerfile
  - textDocument/publishDiagnostics
    - hide `not_pinned_digest` diagnostics from Scout by default ([#216](https://github.com/docker/docker-language-server/issues/216))

### Fixed

- Dockerfile
  - textDocument/publishDiagnostics
    - ignore the diagnostic's URL and do not set it if it is evaluated to be the empty string ([#219](https://github.com/docker/docker-language-server/issues/219))
- Compose
  - textDocument/completion
    - fix panic in code completion in an empty file ([#196](https://github.com/docker/docker-language-server/issues/196))
    - fix line number assumption issues when using code completion for build targets ([#210](https://github.com/docker/docker-language-server/issues/210))
  - textDocument/hover
    - ensure results are returned even if the file has CRLFs ([#205](https://github.com/docker/docker-language-server/issues/205))
- Bake
  - textDocument/publishDiagnostics
    - stop flagging `BUILDKIT_SYNTAX` as an unrecognized `ARG` ([#187](https://github.com/docker/docker-language-server/issues/187))
    - use inheritance to determine if an `ARG` is truly unused ([#198](https://github.com/docker/docker-language-server/issues/198))
    - correct range calculations for malformed variable interpolation errors ([#203](https://github.com/docker/docker-language-server/issues/203))

## [0.7.0] - 2025-05-09

### Added

- Compose
  - textDocument/completion
    - support build stage names for the `target` attribute ([#173](https://github.com/docker/docker-language-server/issues/173))
    - set schema documentation to the completion items ([#176](https://github.com/docker/docker-language-server/issues/176))
    - automatically suggest boolean values for simple boolean attributes ([#179](https://github.com/docker/docker-language-server/issues/179))
    - suggest service names for a service's `extends` or `extends.service` attribute ([#184](https://github.com/docker/docker-language-server/issues/184))
  - textDocument/hover
    - render a referenced service's YAML content as a hover ([#157](https://github.com/docker/docker-language-server/issues/157))

### Fixed

- Compose
  - textDocument/inlayHint
    - prevent circular service dependencies from crashing the server ([#182](https://github.com/docker/docker-language-server/issues/182))

## [0.6.0] - 2025-05-07

### Added

- Compose
  - updated Compose schema to the latest version
  - textDocument/completion
    - improve code completion by automatically including required attributes in completion items ([#155](https://github.com/docker/docker-language-server/issues/155))
  - textDocument/inlayHint
    - show the parent service's value if it is being overridden and they are not object attributes ([#156](https://github.com/docker/docker-language-server/issues/156))
  - textDocument/formatting
    - add support to format YAML files that do not have clear syntactical errors ([#165](https://github.com/docker/docker-language-server/issues/165))
  - textDocument/publishDiagnostics
    - report YAML syntax errors ([#167](https://github.com/docker/docker-language-server/issues/167))

### Fixed

- Compose
  - textDocument/completion
    - suggest completion items for array items that use an object schema directly ([#161](https://github.com/docker/docker-language-server/issues/161))
  - textDocument/definition
    - consider `extends` when looking up a service reference ([#170](https://github.com/docker/docker-language-server/issues/170))
  - textDocument/documentHighlight
    - consider `extends` when looking up a service reference ([#170](https://github.com/docker/docker-language-server/issues/170))
  - textDocument/prepareRename
    - consider `extends` when looking up a service reference ([#170](https://github.com/docker/docker-language-server/issues/170))
  - textDocument/rename
    - consider `extends` when looking up a service reference ([#170](https://github.com/docker/docker-language-server/issues/170))

## [0.5.0] - 2025-05-05

### Added

- Compose
  - updated Compose schema to the latest version ([#117](https://github.com/docker/docker-language-server/issues/117))
  - textDocument/completion
    - suggest dependent service names for the `depends_on` attribute ([#131](https://github.com/docker/docker-language-server/issues/131))
    - suggest dependent network names for the `networks` attribute ([#132](https://github.com/docker/docker-language-server/issues/132))
    - suggest dependent volume names for the `volumes` attribute ([#133](https://github.com/docker/docker-language-server/issues/133))
    - suggest dependent config names for the `configs` attribute ([#134](https://github.com/docker/docker-language-server/issues/134))
    - suggest dependent secret names for the `secrets` attribute ([#135](https://github.com/docker/docker-language-server/issues/135))
  - textDocument/definition
    - support looking up volume references ([#147](https://github.com/docker/docker-language-server/issues/147))
  - textDocument/documentHighlight
    - support highlighting the short form `depends_on` syntax for services ([#70](https://github.com/docker/docker-language-server/issues/70))
    - support highlighting the long form `depends_on` syntax for services ([#71](https://github.com/docker/docker-language-server/issues/71))
    - support highlighting referenced networks, volumes, configs, and secrets ([#145](https://github.com/docker/docker-language-server/issues/145))
  - textDocument/prepareRename
    - support rename preparation requests ([#150](https://github.com/docker/docker-language-server/issues/150))
  - textDocument/rename
    - support renaming named references of services, networks, volumes, configs, and secrets ([#149](https://github.com/docker/docker-language-server/issues/149))

### Fixed

- Dockerfile
  - textDocument/codeAction
    - preserve instruction flags when fixing a `not_pinned_digest` diagnostic ([#123](https://github.com/docker/docker-language-server/issues/123))
- Compose
  - textDocument/completion
    - resolved a spacing offset issue with object or array completions ([#115](https://github.com/docker/docker-language-server/issues/115))
  - textDocument/hover
    - return the hover results for Compose files

## [0.4.1] - 2025-04-29

### Fixed

- Compose
  - textDocument/completion
    - protect the completion calculation code from throwing errors when encountering empty array items ([#112](https://github.com/docker/docker-language-server/issues/112))

## [0.4.0] - 2025-04-28

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
  - textDocument/hover
    - extract descriptions and enum values from the Compose specification and display them as hovers ([#101](https://github.com/docker/docker-language-server/issues/101))

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
  - textDocument/inlayHint
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

[Unreleased]: https://github.com/docker/docker-language-server/compare/v0.10.0...main
[0.10.0]: https://github.com/docker/docker-language-server/compare/v0.9.0...v0.10.0
[0.9.0]: https://github.com/docker/docker-language-server/compare/v0.8.0...v0.9.0
[0.8.0]: https://github.com/docker/docker-language-server/compare/v0.7.0...v0.8.0
[0.7.0]: https://github.com/docker/docker-language-server/compare/v0.6.0...v0.7.0
[0.6.0]: https://github.com/docker/docker-language-server/compare/v0.5.0...v0.6.0
[0.5.0]: https://github.com/docker/docker-language-server/compare/v0.4.1...v0.5.0
[0.4.1]: https://github.com/docker/docker-language-server/compare/v0.4.0...v0.4.1
[0.4.0]: https://github.com/docker/docker-language-server/compare/v0.3.8...v0.4.0
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
