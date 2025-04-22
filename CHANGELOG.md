# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [v2.0.0] - Apr 23, 2025

### Added

- Add web interface

## [v1.1.0] - Jan 27, 2025

### Added

- Allow consuming messages without committing them
- Allow persisting a single message

## [v1.0.0] - Jan 26, 2025

### Added

- Allow dynamic parsing of protobuf encoded messages
- Allow setting up of "profiles" for various Kafka topics, each with its own
    details related to brokers, message encoding, authentication, etc.
- Allow authentication via AWS MSK IAM
- Show topic and consumer group info in the footer

### Changed

- Message metadata and value are now shown in a single viewport
- The command line interface; most of the configuration is now taken from the
    config file

### Removed

- Keymaps to maximize message metadata or value viewport

## [v0.1.0] - Mar 6, 2024

### Added

- A TUI that allows pulling messages from a kafka topic on demand, and viewing
  their metadata and value
- Allow persisting messages to the local filesystem
- Allow skipping messages

[unreleased]: https://github.com/dhth/kplay/compare/v2.0.0...HEAD
[v2.0.0]: https://github.com/dhth/kplay/compare/v1.1.0...v2.0.0
[v1.1.0]: https://github.com/dhth/kplay/compare/v1.0.0...v1.1.0
[v1.0.0]: https://github.com/dhth/kplay/compare/v0.1.0...v1.0.0
[v0.1.0]: https://github.com/dhth/kplay/commits/v0.1.0
