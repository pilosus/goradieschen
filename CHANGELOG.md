# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

...

## [v0.0.2]: 2025-08-03

`RESP2` protocol implementation allows to use `redis-cli` or any other clients that support the protocol
for interactions with `goradieschen`.

### Changed

- Plain text commands replaced with the `RESP2` protocol implementation

### Added

- `EXPIRE` command
- `TTL` command
- `KEYS` command
- `PING` command
- `COMMAND` command
- `FLUSHALL` command

## [v0.0.1]: 2025-07-29

### Added
 
- GET, SET, DEL commands
- Store
- Basic webserver

[Unreleased]: https://github.com/pilosus/goradieschen/compare/v0.0.2...HEAD
[v0.0.2]: https://github.com/pilosus/goradieschen/compare/v0.0.1...v0.0.2
[v0.0.1]: https://github.com/pilosus/goradieschen/compare/v0.0.0...v0.0.1
