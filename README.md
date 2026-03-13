# go-featureflag

Lightweight feature flags for Go with simple on/off toggles, percentage-based rollouts, and flexible loading from JSON or environment variables. Zero external dependencies, fully thread-safe.

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
| `Flags` | Thread-safe collection of feature flags |
| `New()` | Create a new empty Flags collection |
| `Set(name, enabled)` | Set a simple on/off flag |
| `SetPercentage(name, pct)` | Set a percentage rollout flag (0.0 to 1.0, clamped) |
| `Enabled(name)` | Check if a flag is enabled (returns false for unknown) |
| `EnabledFor(name, userID)` | Deterministic per-user check using FNV-32 hashing |
| `Remove(name)` | Remove a flag |
| `All()` | Return a copy of all flags |
| `Size()` | Return the number of flags |
| `FromEnv(prefix)` | Load flags from environment variables |
| `FromJSON(reader)` | Load flags from JSON |
| `MergeJSON(reader)` | Merge JSON flags into existing collection |

## License

MIT
