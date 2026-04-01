# Changelog

## 0.2.2

- Standardize README to 3-badge format with emoji Support section
- Update CI checkout action to v5 for Node.js 24 compatibility
- Add GitHub issue templates, dependabot config, and PR template

## 0.2.1

- Consolidate README badges onto single line

## 0.2.0

- Add `FlagConfig` struct with targeting rules (allowed users, roles, percentage, variants)
- Add `FeatureFlagContext` struct for rich evaluation context
- Add `SetConfig()` for configuring flags with full targeting rules
- Add `EnabledForContext()` for context-aware flag evaluation
- Add `GetVariant()` for consistent A/B variant selection using FNV-32 hashing
- Refactor internal hash logic into shared `hashCheck` helper

## 0.1.1

- Add badges and Development section to README

## 0.1.0

- Initial release
- Simple on/off feature flags
- Percentage rollout with deterministic per-user evaluation
- Load from environment variables or JSON
