package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

type Config struct {
	PlatformToken        string `json:"platform_token"` // Bearer token from platform.deepseek.com
	DefaultPeriod        string `json:"default_period,omitempty"`
	QueryIntervalSeconds int    `json:"query_interval_seconds,omitempty"`
	ShowRequests         *bool  `json:"show_requests,omitempty"`
	ShowCached           *bool  `json:"show_cached,omitempty"`
	ShowNonCached        *bool  `json:"show_non_cached,omitempty"`
	TokenSetAt           string `json:"token_set_at"` // ISO 8601 timestamp of when the bearer token was saved
	configPath           string // path the config was loaded from, for saving back
}

// Load reads config from these locations (first found wins):
//
//  1. ./deepseekMon.json (working directory)
//  2. ~/.config/deepseekMon/config.json
//  3. DEEPSEEK_PLATFORM_TOKEN environment variable (fallback)
//
// PlatformToken is required. UI preferences are optional.
func Load() (*Config, error) {
	paths := []string{
		"deepseekMon.json",
	}

	if home, err := os.UserHomeDir(); err == nil {
		paths = append(paths, filepath.Join(home, ".config", "deepseekMon", "config.json"))
	}

	for _, p := range paths {
		cfg, err := loadFile(p)
		if err == nil && cfg.PlatformToken != "" {
			return cfg, nil
		}
	}

	if token := os.Getenv("DEEPSEEK_PLATFORM_TOKEN"); token != "" {
		return &Config{PlatformToken: token, configPath: "deepseekMon.json"}, nil
	}

	return nil, fmt.Errorf(
		"no platform bearer token found. Create deepseekMon.json with {\"platform_token\":\"...\"}, " +
			"or ~/.config/deepseekMon/config.json, or set DEEPSEEK_PLATFORM_TOKEN",
	)
}

func loadFile(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("%s: %w", path, err)
	}
	cfg.configPath = path
	return &cfg, nil
}

// Save writes the config back to the file it was loaded from.
func (c *Config) Save() error {
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal config: %w", err)
	}
	data = append(data, '\n')
	if err := os.WriteFile(c.configPath, data, 0600); err != nil {
		return fmt.Errorf("write config: %w", err)
	}
	return nil
}

// SetPlatformToken updates the platform bearer token, sets the timestamp, and saves.
func (c *Config) SetPlatformToken(token string) error {
	c.PlatformToken = token
	c.TokenSetAt = time.Now().UTC().Format(time.RFC3339)
	return c.Save()
}

const defaultQueryInterval = 30 * time.Second

func (c *Config) SelectedTab() int {
	if c.DefaultPeriod == "1d" {
		return 1
	}
	return 0
}

func (c *Config) SetSelectedTab(tab int) {
	if tab == 1 {
		c.DefaultPeriod = "1d"
		return
	}
	c.DefaultPeriod = "30d"
}

func (c *Config) QueryInterval() time.Duration {
	if c.QueryIntervalSeconds > 0 {
		return time.Duration(c.QueryIntervalSeconds) * time.Second
	}
	return defaultQueryInterval
}

func (c *Config) RequestsVisible() bool {
	return boolOrDefault(c.ShowRequests, true)
}

func (c *Config) CachedVisible() bool {
	return boolOrDefault(c.ShowCached, true)
}

func (c *Config) NonCachedVisible() bool {
	return boolOrDefault(c.ShowNonCached, true)
}

func boolOrDefault(value *bool, fallback bool) bool {
	if value == nil {
		return fallback
	}
	return *value
}

// ExampleConfig returns a sample config JSON for bootstrapping.
func ExampleConfig() []byte {
	return []byte(`{
  "platform_token": "your-platform-bearer-token-here",
  "default_period": "30d",
  "query_interval_seconds": 30,
  "show_requests": true,
  "show_cached": true,
  "show_non_cached": true,
  "token_set_at": "2026-05-08T00:00:00Z"
}
`)
}
