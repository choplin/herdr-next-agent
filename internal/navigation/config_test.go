package navigation

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestLoadConfigUsesDefaults(t *testing.T) {
	t.Parallel()

	for _, dir := range []string{"", t.TempDir()} {
		cfg, err := LoadConfig(dir)
		if err != nil {
			t.Fatalf("LoadConfig(%q) error = %v", dir, err)
		}
		if cfg.Order != DisplayOrder {
			t.Fatalf("LoadConfig(%q).Order = %q, want %q", dir, cfg.Order, DisplayOrder)
		}
		if len(cfg.States) != 1 || cfg.States[0] != "blocked" {
			t.Fatalf("LoadConfig(%q).States = %q, want [blocked]", dir, cfg.States)
		}
	}
}

func TestLoadConfigReadsWaitingOrder(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, configFileName), []byte("order = \"waiting\"\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	cfg, err := LoadConfig(dir)
	if err != nil {
		t.Fatalf("LoadConfig() error = %v", err)
	}
	if cfg.Order != WaitingOrder {
		t.Fatalf("Order = %q, want %q", cfg.Order, WaitingOrder)
	}
}

func TestLoadConfigReadsStates(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	contents := "states = [\"blocked\", \"done\"]\n"
	if err := os.WriteFile(filepath.Join(dir, configFileName), []byte(contents), 0o600); err != nil {
		t.Fatal(err)
	}
	cfg, err := LoadConfig(dir)
	if err != nil {
		t.Fatalf("LoadConfig() error = %v", err)
	}
	if len(cfg.States) != 2 || cfg.States[0] != "blocked" || cfg.States[1] != "done" {
		t.Fatalf("States = %q, want [blocked done]", cfg.States)
	}
}

func TestLoadConfigRejectsInvalidConfiguration(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		contents string
	}{
		{name: "unsupported order", contents: "order = \"newest\"\n"},
		{name: "empty states", contents: "states = []\n"},
		{name: "unsupported state", contents: "states = [\"paused\"]\n"},
		{name: "duplicate state", contents: "states = [\"blocked\", \"blocked\"]\n"},
		{name: "unknown field", contents: "ordering = \"display\"\n"},
		{name: "malformed TOML", contents: "order = [\n"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			dir := t.TempDir()
			if err := os.WriteFile(filepath.Join(dir, configFileName), []byte(test.contents), 0o600); err != nil {
				t.Fatal(err)
			}
			_, err := LoadConfig(dir)
			var readErr *ConfigReadError
			var invalidErr *InvalidConfigError
			if !errors.As(err, &readErr) && !errors.As(err, &invalidErr) {
				t.Fatalf("LoadConfig() error = %v, want typed config error", err)
			}
		})
	}
}
