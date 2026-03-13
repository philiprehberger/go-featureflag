package featureflag

import (
	"strings"
	"testing"
)

func TestFromJSONBooleanFlags(t *testing.T) {
	input := `{"feature_a": true, "feature_b": false}`
	f, err := FromJSON(strings.NewReader(input))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !f.Enabled("feature_a") {
		t.Fatal("expected feature_a to be enabled")
	}
	if f.Enabled("feature_b") {
		t.Fatal("expected feature_b to be disabled")
	}
}

func TestFromJSONPercentageFlags(t *testing.T) {
	input := `{"rollout": 0.75}`
	f, err := FromJSON(strings.NewReader(input))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	all := f.All()
	if all["rollout"].Percentage != 0.75 {
		t.Fatalf("expected percentage 0.75, got %f", all["rollout"].Percentage)
	}
}

func TestFromJSONMixed(t *testing.T) {
	input := `{"simple": true, "rollout": 0.5, "off": false}`
	f, err := FromJSON(strings.NewReader(input))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if f.Size() != 3 {
		t.Fatalf("expected 3 flags, got %d", f.Size())
	}
	if !f.Enabled("simple") {
		t.Fatal("expected simple to be enabled")
	}
	if f.Enabled("off") {
		t.Fatal("expected off to be disabled")
	}
	all := f.All()
	if all["rollout"].Percentage != 0.5 {
		t.Fatalf("expected rollout percentage 0.5, got %f", all["rollout"].Percentage)
	}
}

func TestFromJSONInvalid(t *testing.T) {
	input := `not valid json`
	_, err := FromJSON(strings.NewReader(input))
	if err == nil {
		t.Fatal("expected error for invalid JSON")
	}
}

func TestFromEnvMatchingPrefix(t *testing.T) {
	t.Setenv("MYAPP_DARK_MODE", "true")
	t.Setenv("MYAPP_BETA", "false")
	t.Setenv("MYAPP_ROLLOUT", "0.75")

	f := FromEnv("MYAPP")

	if !f.Enabled("dark_mode") {
		t.Fatal("expected dark_mode to be enabled")
	}
	if f.Enabled("beta") {
		t.Fatal("expected beta to be disabled")
	}
	all := f.All()
	if all["rollout"].Percentage != 0.75 {
		t.Fatalf("expected rollout percentage 0.75, got %f", all["rollout"].Percentage)
	}
}

func TestFromEnvIgnoresNonMatching(t *testing.T) {
	t.Setenv("MYAPP_FEATURE", "true")
	t.Setenv("OTHER_FEATURE", "true")
	t.Setenv("PATH_EXTRA", "/usr/bin")

	f := FromEnv("MYAPP")

	if f.Size() != 1 {
		t.Fatalf("expected 1 flag, got %d", f.Size())
	}
	if !f.Enabled("feature") {
		t.Fatal("expected feature to be enabled")
	}
}

func TestMergeJSONAddsAndOverwrites(t *testing.T) {
	f := New()
	f.Set("existing", true)
	f.Set("overwrite_me", false)

	input := `{"new_flag": true, "overwrite_me": true, "rollout": 0.3}`
	err := f.MergeJSON(strings.NewReader(input))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if f.Size() != 4 {
		t.Fatalf("expected 4 flags, got %d", f.Size())
	}
	if !f.Enabled("existing") {
		t.Fatal("expected existing to remain enabled")
	}
	if !f.Enabled("new_flag") {
		t.Fatal("expected new_flag to be enabled")
	}
	if !f.Enabled("overwrite_me") {
		t.Fatal("expected overwrite_me to be overwritten to true")
	}
	all := f.All()
	if all["rollout"].Percentage != 0.3 {
		t.Fatalf("expected rollout percentage 0.3, got %f", all["rollout"].Percentage)
	}
}

func TestMergeJSONInvalid(t *testing.T) {
	f := New()
	f.Set("existing", true)

	err := f.MergeJSON(strings.NewReader("bad json"))
	if err == nil {
		t.Fatal("expected error for invalid JSON")
	}

	// Existing flags should be untouched
	if !f.Enabled("existing") {
		t.Fatal("expected existing flag to remain after failed merge")
	}
}
