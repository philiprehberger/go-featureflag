package featureflag

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
)

// FromEnv loads feature flags from environment variables matching the given prefix.
// Variables should follow the pattern {PREFIX}_FEATURE_NAME=true/false for simple flags
// or {PREFIX}_FEATURE_NAME=0.5 for percentage rollout flags.
// The flag name is derived from the variable name by stripping the prefix and
// converting to lowercase.
func FromEnv(prefix string) *Flags {
	f := New()
	prefix = strings.ToUpper(prefix) + "_"
	for _, env := range os.Environ() {
		parts := strings.SplitN(env, "=", 2)
		if len(parts) != 2 {
			continue
		}
		key, val := parts[0], parts[1]
		if !strings.HasPrefix(strings.ToUpper(key), prefix) {
			continue
		}
		name := strings.ToLower(key[len(prefix):])
		if name == "" {
			continue
		}

		// Try boolean first
		if b, err := strconv.ParseBool(val); err == nil {
			f.flags[name] = Flag{Enabled: b}
			continue
		}

		// Try float (percentage)
		if pct, err := strconv.ParseFloat(val, 64); err == nil {
			if pct < 0 {
				pct = 0
			}
			if pct > 1 {
				pct = 1
			}
			f.flags[name] = Flag{Percentage: pct}
			continue
		}
	}
	return f
}

// FromJSON loads feature flags from a JSON reader.
// The JSON format is a map of flag names to values:
//
//	{"feature_name": true, "other_feature": 0.75}
//
// Boolean values create simple on/off flags. Numeric values create percentage rollout flags.
func FromJSON(r io.Reader) (*Flags, error) {
	f := New()
	if err := parseJSON(r, f); err != nil {
		return nil, err
	}
	return f, nil
}

// MergeJSON merges flags from a JSON reader into the existing Flags collection.
// New flags are added and existing flags are overwritten.
func (f *Flags) MergeJSON(r io.Reader) error {
	return parseJSON(r, f)
}

func parseJSON(r io.Reader, f *Flags) error {
	var raw map[string]json.RawMessage
	if err := json.NewDecoder(r).Decode(&raw); err != nil {
		return fmt.Errorf("featureflag: invalid JSON: %w", err)
	}

	f.mu.Lock()
	defer f.mu.Unlock()

	for name, val := range raw {
		// Try boolean
		var b bool
		if err := json.Unmarshal(val, &b); err == nil {
			f.flags[name] = Flag{Enabled: b}
			continue
		}

		// Try float
		var pct float64
		if err := json.Unmarshal(val, &pct); err == nil {
			if pct < 0 {
				pct = 0
			}
			if pct > 1 {
				pct = 1
			}
			f.flags[name] = Flag{Percentage: pct}
			continue
		}
	}
	return nil
}
