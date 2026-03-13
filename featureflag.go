// Package featureflag provides lightweight feature flags for Go.
package featureflag

import (
	"hash/fnv"
	"sync"
)

// Flag represents a single feature flag configuration.
// If Percentage is 0, the Enabled field determines whether the flag is on or off.
// If Percentage is greater than 0, it represents a percentage rollout (0.0 to 1.0).
type Flag struct {
	Enabled    bool
	Percentage float64
}

// Flags is a thread-safe collection of feature flags.
type Flags struct {
	mu    sync.RWMutex
	flags map[string]Flag
}

// New creates a new empty Flags collection.
func New() *Flags {
	return &Flags{
		flags: make(map[string]Flag),
	}
}

// Set sets a simple on/off feature flag.
func (f *Flags) Set(name string, enabled bool) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.flags[name] = Flag{Enabled: enabled}
}

// SetPercentage sets a percentage rollout flag.
// The percentage is clamped to the range [0.0, 1.0].
func (f *Flags) SetPercentage(name string, pct float64) {
	if pct < 0 {
		pct = 0
	}
	if pct > 1 {
		pct = 1
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	f.flags[name] = Flag{Percentage: pct}
}

// Enabled checks if a simple on/off flag is enabled.
// Returns false for unknown flags.
func (f *Flags) Enabled(name string) bool {
	f.mu.RLock()
	defer f.mu.RUnlock()
	flag, ok := f.flags[name]
	if !ok {
		return false
	}
	return flag.Enabled
}

// EnabledFor checks if a flag is enabled for a specific user.
// For percentage rollout flags, the result is deterministic based on the
// combination of userID and flag name using FNV-32 hashing.
// For simple flags (Percentage == 0), it falls back to the Enabled field.
// Returns false for unknown flags.
func (f *Flags) EnabledFor(name string, userID string) bool {
	f.mu.RLock()
	defer f.mu.RUnlock()
	flag, ok := f.flags[name]
	if !ok {
		return false
	}
	if flag.Percentage == 0 {
		return flag.Enabled
	}
	h := fnv.New32a()
	h.Write([]byte(userID + name))
	hash := h.Sum32()
	return hash%1000 < uint32(flag.Percentage*1000)
}

// Remove removes a flag by name. No-op if the flag does not exist.
func (f *Flags) Remove(name string) {
	f.mu.Lock()
	defer f.mu.Unlock()
	delete(f.flags, name)
}

// All returns a copy of all flags.
func (f *Flags) All() map[string]Flag {
	f.mu.RLock()
	defer f.mu.RUnlock()
	result := make(map[string]Flag, len(f.flags))
	for k, v := range f.flags {
		result[k] = v
	}
	return result
}

// Size returns the number of flags.
func (f *Flags) Size() int {
	f.mu.RLock()
	defer f.mu.RUnlock()
	return len(f.flags)
}
