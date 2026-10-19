package daemon

import (
	"errors"
	"path/filepath"
	"testing"
)

func TestResolveStateRoot_ExpandsTilde(t *testing.T) {
	t.Setenv("HOME", "/home/testuser")

	got, err := resolveStateRoot("~/.devsy/data")
	if err != nil {
		t.Fatalf("resolveStateRoot() error = %v", err)
	}
	want := "/home/testuser/.devsy/data"
	if got != want {
		t.Errorf("resolveStateRoot() = %q, want %q", got, want)
	}
}

func TestResolveStateRoot_AlreadyAbsolute(t *testing.T) {
	t.Setenv("HOME", "/home/testuser")

	got, err := resolveStateRoot("/opt/devsy/data")
	if err != nil {
		t.Fatalf("resolveStateRoot() error = %v", err)
	}
	want := "/opt/devsy/data"
	if got != want {
		t.Errorf("resolveStateRoot() = %q, want %q", got, want)
	}
}

func TestResolveStateRoot_RejectsRelative(t *testing.T) {
	got, err := resolveStateRoot("relative/path")
	if err == nil {
		t.Fatalf("resolveStateRoot() expected error, got nil")
	}
	var relativeErr *relativeRootError
	if !errors.As(err, &relativeErr) {
		t.Fatalf("resolveStateRoot() expected *relativeRootError, got %v", err)
	}
	if relativeErr.raw != "relative/path" {
		t.Errorf("relativeRootError.raw = %q, want %q", relativeErr.raw, "relative/path")
	}
	if filepath.IsAbs(got) {
		t.Errorf("resolveStateRoot() returned absolute root %q, want non-absolute", got)
	}
}

// relativeRootError is returned when the raw state root is not absolute.
type relativeRootError struct {
	raw string
}

func (e *relativeRootError) Error() string {
	return "daemon state root must be absolute, got " + e.raw +
		"; set DEVSSY_DATA_ROOT to an absolute path"
}

