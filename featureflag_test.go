package featureflag

import (
	"fmt"
	"sync"
	"testing"
)

func TestSetAndEnabled(t *testing.T) {
	f := New()
	f.Set("feature_a", true)
	if !f.Enabled("feature_a") {
		t.Fatal("expected feature_a to be enabled")
	}

	f.Set("feature_b", false)
	if f.Enabled("feature_b") {
		t.Fatal("expected feature_b to be disabled")
	}
}

func TestEnabledUnknownFlag(t *testing.T) {
	f := New()
	if f.Enabled("nonexistent") {
		t.Fatal("expected unknown flag to return false")
	}
}

func TestSetPercentageAndEnabledFor(t *testing.T) {
	f := New()
	f.SetPercentage("rollout", 0.5)

	// Deterministic: same user+flag should always return the same result
	result1 := f.EnabledFor("rollout", "user-123")
	result2 := f.EnabledFor("rollout", "user-123")
	if result1 != result2 {
		t.Fatal("expected deterministic result for same user+flag")
	}
}

func TestEnabledForZeroPercent(t *testing.T) {
	f := New()
	f.SetPercentage("disabled", 0.0)

	for i := 0; i < 100; i++ {
		if f.EnabledFor("disabled", fmt.Sprintf("user-%d", i)) {
			t.Fatal("expected 0% rollout to always return false")
		}
	}
}

func TestEnabledForFullPercent(t *testing.T) {
	f := New()
	f.SetPercentage("full", 1.0)

	for i := 0; i < 100; i++ {
		if !f.EnabledFor("full", fmt.Sprintf("user-%d", i)) {
			t.Fatal("expected 100% rollout to always return true")
		}
	}
}

func TestEnabledForDistribution(t *testing.T) {
	f := New()
	f.SetPercentage("half", 0.5)

	enabled := 0
	total := 10000
	for i := 0; i < total; i++ {
		if f.EnabledFor("half", fmt.Sprintf("user-%d", i)) {
			enabled++
		}
	}

	ratio := float64(enabled) / float64(total)
	// Allow a generous margin: 0.5 ± 0.1
	if ratio < 0.4 || ratio > 0.6 {
		t.Fatalf("expected ~50%% rollout, got %.2f%%", ratio*100)
	}
}

func TestEnabledForFallsBackToEnabled(t *testing.T) {
	f := New()
	f.Set("simple", true)
	if !f.EnabledFor("simple", "any-user") {
		t.Fatal("expected EnabledFor to fall back to Enabled for simple flags")
	}

	f.Set("simple", false)
	if f.EnabledFor("simple", "any-user") {
		t.Fatal("expected EnabledFor to fall back to Enabled=false for simple flags")
	}
}

func TestEnabledForUnknownFlag(t *testing.T) {
	f := New()
	if f.EnabledFor("nonexistent", "user-1") {
		t.Fatal("expected unknown flag to return false")
	}
}

func TestRemove(t *testing.T) {
	f := New()
	f.Set("feature", true)
	f.Remove("feature")
	if f.Enabled("feature") {
		t.Fatal("expected removed flag to return false")
	}
	if f.Size() != 0 {
		t.Fatalf("expected size 0 after remove, got %d", f.Size())
	}
}

func TestRemoveNonexistent(t *testing.T) {
	f := New()
	f.Remove("nonexistent") // should not panic
}

func TestAll(t *testing.T) {
	f := New()
	f.Set("a", true)
	f.Set("b", false)
	f.SetPercentage("c", 0.5)

	all := f.All()
	if len(all) != 3 {
		t.Fatalf("expected 3 flags, got %d", len(all))
	}

	// Modifying the returned map should not affect the original
	delete(all, "a")
	if f.Size() != 3 {
		t.Fatal("expected All() to return a copy")
	}
}

func TestSize(t *testing.T) {
	f := New()
	if f.Size() != 0 {
		t.Fatalf("expected size 0, got %d", f.Size())
	}

	f.Set("a", true)
	f.Set("b", false)
	if f.Size() != 2 {
		t.Fatalf("expected size 2, got %d", f.Size())
	}
}

func TestSetPercentageClamping(t *testing.T) {
	f := New()
	f.SetPercentage("over", 1.5)
	f.SetPercentage("under", -0.5)

	all := f.All()
	if all["over"].Percentage != 1.0 {
		t.Fatalf("expected clamped to 1.0, got %f", all["over"].Percentage)
	}
	if all["under"].Percentage != 0.0 {
		t.Fatalf("expected clamped to 0.0, got %f", all["under"].Percentage)
	}
}

func TestSetConfigAndEnabledForContext(t *testing.T) {
	f := New()
	f.SetConfig("beta", FlagConfig{
		Enabled:      false,
		AllowedUsers: []string{"alice", "bob"},
		AllowedRoles: []string{"admin"},
	})

	// Allowed user should be enabled.
	ctx := FeatureFlagContext{UserID: "alice"}
	if !f.EnabledForContext("beta", ctx) {
		t.Fatal("expected allowed user alice to be enabled")
	}

	// Non-allowed user without matching role should fall back to Enabled (false).
	ctx = FeatureFlagContext{UserID: "charlie"}
	if f.EnabledForContext("beta", ctx) {
		t.Fatal("expected non-allowed user charlie to be disabled")
	}

	// Non-allowed user with matching role should be enabled.
	ctx = FeatureFlagContext{UserID: "charlie", Roles: []string{"admin"}}
	if !f.EnabledForContext("beta", ctx) {
		t.Fatal("expected user with admin role to be enabled")
	}
}

func TestEnabledForContextRoleMatching(t *testing.T) {
	f := New()
	f.SetConfig("internal", FlagConfig{
		Enabled:      false,
		AllowedRoles: []string{"staff", "moderator"},
	})

	// No roles at all.
	ctx := FeatureFlagContext{UserID: "user1"}
	if f.EnabledForContext("internal", ctx) {
		t.Fatal("expected user with no roles to be disabled")
	}

	// Non-matching role.
	ctx = FeatureFlagContext{UserID: "user1", Roles: []string{"viewer"}}
	if f.EnabledForContext("internal", ctx) {
		t.Fatal("expected user with viewer role to be disabled")
	}

	// One matching role among several.
	ctx = FeatureFlagContext{UserID: "user1", Roles: []string{"viewer", "moderator"}}
	if !f.EnabledForContext("internal", ctx) {
		t.Fatal("expected user with moderator role to be enabled")
	}
}

func TestEnabledForContextPercentage(t *testing.T) {
	f := New()
	f.SetConfig("gradual", FlagConfig{
		Percentage: 0.5,
	})

	enabled := 0
	total := 10000
	for i := 0; i < total; i++ {
		ctx := FeatureFlagContext{UserID: fmt.Sprintf("user-%d", i)}
		if f.EnabledForContext("gradual", ctx) {
			enabled++
		}
	}
	ratio := float64(enabled) / float64(total)
	if ratio < 0.4 || ratio > 0.6 {
		t.Fatalf("expected ~50%% rollout, got %.2f%%", ratio*100)
	}
}

func TestEnabledForContextFallback(t *testing.T) {
	f := New()
	// Simple flag (no config), should still work via fallback.
	f.Set("simple", true)
	ctx := FeatureFlagContext{UserID: "user1"}
	if !f.EnabledForContext("simple", ctx) {
		t.Fatal("expected fallback to simple flag enabled=true")
	}

	// Unknown flag.
	if f.EnabledForContext("unknown", ctx) {
		t.Fatal("expected unknown flag to return false")
	}
}

func TestEnabledForContextAllowedUserTakesPriority(t *testing.T) {
	f := New()
	f.SetConfig("exclusive", FlagConfig{
		Enabled:      false,
		Percentage:   0,
		AllowedUsers: []string{"vip"},
	})

	// VIP user should be enabled even though Enabled=false and Percentage=0.
	ctx := FeatureFlagContext{UserID: "vip"}
	if !f.EnabledForContext("exclusive", ctx) {
		t.Fatal("expected allowed user to override disabled flag")
	}
}

func TestGetVariantConsistent(t *testing.T) {
	f := New()
	f.SetConfig("experiment", FlagConfig{
		Variants: []string{"control", "variant_a", "variant_b"},
	})

	// Same user always gets the same variant.
	v1 := f.GetVariant("experiment", "user-42")
	v2 := f.GetVariant("experiment", "user-42")
	if v1 != v2 {
		t.Fatalf("expected consistent variant, got %q and %q", v1, v2)
	}

	// Variant must be one of the configured values.
	valid := map[string]bool{"control": true, "variant_a": true, "variant_b": true}
	if !valid[v1] {
		t.Fatalf("unexpected variant %q", v1)
	}
}

func TestGetVariantDistribution(t *testing.T) {
	f := New()
	f.SetConfig("ab_test", FlagConfig{
		Variants: []string{"a", "b"},
	})

	counts := map[string]int{"a": 0, "b": 0}
	total := 10000
	for i := 0; i < total; i++ {
		v := f.GetVariant("ab_test", fmt.Sprintf("user-%d", i))
		counts[v]++
	}

	ratioA := float64(counts["a"]) / float64(total)
	if ratioA < 0.3 || ratioA > 0.7 {
		t.Fatalf("expected roughly even split, got a=%.2f%% b=%.2f%%", ratioA*100, (1-ratioA)*100)
	}
}

func TestGetVariantNoVariants(t *testing.T) {
	f := New()
	f.SetConfig("no_variants", FlagConfig{Enabled: true})

	if v := f.GetVariant("no_variants", "user-1"); v != "" {
		t.Fatalf("expected empty string for flag with no variants, got %q", v)
	}
}

func TestGetVariantUnknownFlag(t *testing.T) {
	f := New()
	if v := f.GetVariant("nonexistent", "user-1"); v != "" {
		t.Fatalf("expected empty string for unknown flag, got %q", v)
	}
}

func TestSetConfigClampPercentage(t *testing.T) {
	f := New()
	f.SetConfig("over", FlagConfig{Percentage: 1.5})
	f.SetConfig("under", FlagConfig{Percentage: -0.5})

	all := f.All()
	if all["over"].Percentage != 1.0 {
		t.Fatalf("expected clamped to 1.0, got %f", all["over"].Percentage)
	}
	if all["under"].Percentage != 0.0 {
		t.Fatalf("expected clamped to 0.0, got %f", all["under"].Percentage)
	}
}

func TestRemoveAlsoRemovesConfig(t *testing.T) {
	f := New()
	f.SetConfig("temp", FlagConfig{
		Enabled:  true,
		Variants: []string{"a", "b"},
	})
	f.Remove("temp")

	if f.Enabled("temp") {
		t.Fatal("expected removed flag to return false")
	}
	if v := f.GetVariant("temp", "user"); v != "" {
		t.Fatalf("expected empty variant after remove, got %q", v)
	}
}

func TestConcurrentAccess(t *testing.T) {
	f := New()
	var wg sync.WaitGroup

	// Concurrent writes
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			f.Set(fmt.Sprintf("flag-%d", i), i%2 == 0)
		}(i)
	}

	// Concurrent reads
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			f.Enabled(fmt.Sprintf("flag-%d", i))
			f.EnabledFor(fmt.Sprintf("flag-%d", i), "user")
			f.All()
			f.Size()
		}(i)
	}

	wg.Wait()

	if f.Size() != 100 {
		t.Fatalf("expected 100 flags after concurrent writes, got %d", f.Size())
	}
}
