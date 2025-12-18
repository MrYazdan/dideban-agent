# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- Initial project setup

## [0.1.0] - 2024-12-18

### Added
- ✅ Core metrics collection (CPU, Memory, Disk)
- ✅ HTTP transmission with retry logic
- ✅ Development mode with mock sender
- ✅ YAML configuration system
- ✅ Environment variable support
- ✅ Structured logging (JSON/pretty)
- ✅ Graceful shutdown handling
- ✅ Performance tracking
- ✅ Cross-platform support (Linux/Windows)
- ✅ Bearer token authentication

### Features
- 🖥️ CPU metrics: Usage percentage & load averages
- 🧠 Memory metrics: Used, total, available with percentages
- 💽 Disk metrics: Used, total space with percentages
- ⏱️ Collection duration measurement
- 🔁 Configurable collection intervals
- 📤 Push-based HTTP delivery
- 🔧 Dual mode operation (dev/prod)
- 🔐 Secure authentication
- 🛡️ Resilient design with exponential backoff

[Unreleased]: https://github.com/MrYazdan/dideban-agent/compare/v0.1.0...HEAD
[0.1.0]: https://github.com/MrYazdan/dideban-agent/releases/tag/v0.1.0