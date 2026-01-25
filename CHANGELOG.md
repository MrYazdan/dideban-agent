## [0.1.1] - 2026-01-25

### Changed
- 🔁 Renamed agent identifier from `id` to `name` across the entire codebase
- 🧩 Updated `Config` structure to use `Agent.Name` instead of `Agent.ID`
- ⚙️ Default agent identifier generator renamed to `getDefaultAgentName`
- 🧪 Updated validation logic to check `Agent.Name`
- 📝 Updated example configuration (`config.example.yaml`)
- 📉 Removed `agent_id` metric references from mock and metrics collection
- 🧹 Internal refactoring to align naming consistency


### Breaking Changes
- ⚠️ Configuration field `agent.id` has been renamed to `agent.name`

[Unreleased]: https://github.com/MrYazdan/dideban-agent/compare/v0.1.1...HEAD
[0.1.1]: https://github.com/MrYazdan/dideban-agent/compare/v0.1.0...v0.1.1