# Changelog

All notable changes to this project will be documented in this file.

## [Unreleased]

### Fixed

- Fixed `Publisher.Publish` retry logic by correcting the closure signature passed to `retry.DoContext` (removed redundant `ctx` parameter).
- Corrected message publishing logic and brought all RabbitMQ package code into compliance with `golangci-lint` standards.
