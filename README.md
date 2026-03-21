# go-featureflag

[![CI](https://github.com/philiprehberger/go-featureflag/actions/workflows/ci.yml/badge.svg)](https://github.com/philiprehberger/go-featureflag/actions/workflows/ci.yml) [![Go Reference](https://pkg.go.dev/badge/github.com/philiprehberger/go-featureflag.svg)](https://pkg.go.dev/github.com/philiprehberger/go-featureflag) [![License](https://img.shields.io/github/license/philiprehberger/go-featureflag)](LICENSE)

Lightweight feature flags for Go with toggles, percentage rollouts, user targeting, and A/B variants

## Installation

```bash
go get github.com/philiprehberger/go-featureflag
```

## Usage

### Basic on/off flags

```go
import "github.com/philiprehberger/go-featureflag"

flags := featureflag.New()
flags.Set("dark_mode", true)
flags.Set("legacy_api", false)

if flags.Enabled("dark_mode") {
    // dark mode is on
}
```

### Percentage rollout

```go
flags.SetPercentage("new_checkout", 0.25) // 25% of users

// Deterministic per-user: same user always gets the same result
if flags.EnabledFor("new_checkout", userID) {
    // show new checkout flow
}
```

### Role & User Targeting

```go
flags.SetConfig("beta_feature", featureflag.FlagConfig{
    Enabled:      false,
    AllowedUsers: []string{"alice", "bob"},
    AllowedRoles: []string{"admin", "staff"},
})

ctx := featureflag.FeatureFlagContext{
    UserID: "alice",
    Roles:  []string{"viewer"},
}

flags.EnabledForContext("beta_feature", ctx) // true (alice is in AllowedUsers)
```

### A/B Variants

```go
flags.SetConfig("checkout_experiment", featureflag.FlagConfig{
    Enabled:  true,
    Variants: []string{"control", "variant_a", "variant_b"},
})

variant := flags.GetVariant("checkout_experiment", userID)
// Returns a consistent variant per user using FNV-32 hashing
```

### Context-Based Evaluation

```go
flags.SetConfig("gradual_rollout", featureflag.FlagConfig{
    Enabled:      false,
    Percentage:   0.25,
    AllowedUsers: []string{"vip_user"},
    AllowedRoles: []string{"beta_tester"},
})

ctx := featureflag.FeatureFlagContext{
    UserID: "some_user",
    Roles:  []string{"beta_tester"},
    Properties: map[string]string{
        "region": "eu-west",
    },
}

// Evaluation order: AllowedUsers -> AllowedRoles -> Percentage -> Enabled
flags.EnabledForContext("gradual_rollout", ctx) // true (role matches)
```

### Load from JSON

```go
import "strings"

data := `{"dark_mode": true, "new_checkout": 0.5}`
flags, err := featureflag.FromJSON(strings.NewReader(data))
if err != nil {
    log.Fatal(err)
}
```

### Load from environment variables

```bash
export MYAPP_DARK_MODE=true
export MYAPP_NEW_CHECKOUT=0.5
```

```go
flags := featureflag.FromEnv("MYAPP")
// flags now contains "dark_mode" (enabled) and "new_checkout" (50% rollout)
```

## API

| Function / Type | Description |
|-----------------|-------------|
| `Flag` | Struct with `Enabled bool` and `Percentage float64` fields |
| `FlagConfig` | Extended config with `Enabled`, `Percentage`, `AllowedUsers`, `AllowedRoles`, `Variants` |
| `FeatureFlagContext` | Evaluation context with `UserID`, `Roles`, `Properties` |
| `Flags` | Thread-safe collection of feature flags |
| `New()` | Create a new empty Flags collection |
| `Set(name, enabled)` | Set a simple on/off flag |
| `SetPercentage(name, pct)` | Set a percentage rollout flag (0.0 to 1.0, clamped) |
| `SetConfig(name, config)` | Set a flag with full targeting rules |
| `Enabled(name)` | Check if a flag is enabled (returns false for unknown) |
| `EnabledFor(name, userID)` | Deterministic per-user check using FNV-32 hashing |
| `EnabledForContext(name, ctx)` | Context-aware evaluation (users, roles, percentage, enabled) |
| `GetVariant(name, userID)` | Consistent A/B variant selection using FNV-32 hashing |
| `Remove(name)` | Remove a flag (including config) |
| `All()` | Return a copy of all flags |
| `Size()` | Return the number of flags |
| `FromEnv(prefix)` | Load flags from environment variables |
| `FromJSON(reader)` | Load flags from JSON |
| `MergeJSON(reader)` | Merge JSON flags into existing collection |

## Development

```bash
go test ./...
go vet ./...
```

## License

MIT
