package config

import (
	"os"
	"testing"
	"time"
)

func TestSelectedTabDefaultsTo30d(t *testing.T) {
	var cfg Config
	if got := cfg.SelectedTab(); got != 0 {
		t.Fatalf("expected default tab 0, got %d", got)
	}
}

func TestSelectedTabSupports1d(t *testing.T) {
	cfg := Config{DefaultPeriod: "1d"}
	if got := cfg.SelectedTab(); got != 1 {
		t.Fatalf("expected tab 1, got %d", got)
	}
}

func TestUsageVisibilityDefaultsToShown(t *testing.T) {
	var cfg Config
	if !cfg.RequestsVisible() {
		t.Fatal("expected requests to be visible by default")
	}
	if !cfg.CachedVisible() {
		t.Fatal("expected cached tokens to be visible by default")
	}
	if !cfg.NonCachedVisible() {
		t.Fatal("expected non-cached tokens to be visible by default")
	}
}

func TestLoadSupportsPlatformTokenEnv(t *testing.T) {
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	t.Cleanup(func() {
		_ = os.Chdir(wd)
	})

	if err := os.Chdir(t.TempDir()); err != nil {
		t.Fatalf("chdir: %v", err)
	}
	t.Setenv("DEEPSEEK_PLATFORM_TOKEN", "token-from-env")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.PlatformToken != "token-from-env" {
		t.Fatalf("expected env token, got %q", cfg.PlatformToken)
	}
}

func TestLoadRequiresPlatformToken(t *testing.T) {
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	t.Cleanup(func() {
		_ = os.Chdir(wd)
	})

	if err := os.Chdir(t.TempDir()); err != nil {
		t.Fatalf("chdir: %v", err)
	}
	t.Setenv("DEEPSEEK_PLATFORM_TOKEN", "")

	if _, err := Load(); err == nil {
		t.Fatal("expected missing token error")
	}
}

func TestQueryIntervalUsesConfiguredSeconds(t *testing.T) {
	cfg := Config{QueryIntervalSeconds: 11}
	if got := cfg.QueryInterval(); got != 11*time.Second {
		t.Fatalf("expected 11s, got %v", got)
	}
}

func TestQueryIntervalDefaultsTo30s(t *testing.T) {
	var cfg Config
	if got := cfg.QueryInterval(); got != 30*time.Second {
		t.Fatalf("expected 30s default, got %v", got)
	}
}
