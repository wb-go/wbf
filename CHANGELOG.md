# Changelog

All notable changes to this project will be documented in this file.

## [Unreleased]

### Fixed

- Fixed `Publisher.Publish` retry logic by correcting the closure signature passed to `retry.DoContext` (removed redundant `ctx` parameter).
- Fixed `WithExpiration` Fixed incorrect sending of the Expiration value
- Fixed `Consumer.consumeOnce` Fixed the freezing of 1 message
- Added `Publisher.GetExchangeName` method getting Exchange name
- Corrected message publishing logic and brought all RabbitMQ package code into compliance with `golangci-lint` standards.
