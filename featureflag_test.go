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
