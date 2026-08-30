package config

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"
)

// Config represents the application configuration
type Config struct {
	API          APIConfig       `json:"api"`
	Cache        CacheConfig     `json:"cache"`
	Memory       MemoryConfig    `json:"memory"`
	Export       ExportConfig    `json:"export"`
	Validation   ValidationConfig `json:"validation"`
	Logging      LoggingConfig   `json:"logging"`
}

// APIConfig contains API-specific settings
type APIConfig struct {
	Port              int           `json:"port"`
	Host              string        `json:"host"`
	ReadTimeout       string        `json:"read_timeout"`
	WriteTimeout      string        `json:"write_timeout"`
	MaxRequestSize    int64         `json:"max_request_size"`
	EnableCORS        bool          `json:"enable_cors"`
	AllowedOrigins    []string      `json:"allowed_origins"`
}

// CacheConfig contains cache settings
type CacheConfig struct {
	Enabled       bool          `json:"enabled"`
	MaxSize       int           `json:"max_size"`
	TTL           string        `json:"ttl"`
	CleanupInterval string       `json:"cleanup_interval"`
}

// MemoryConfig contains memory/embeddings settings
type MemoryConfig struct {
	Enabled           bool   `json:"enabled"`
	DatabasePath      string `json:"database_path"`
	MaxEmbeddingSize  int    `json:"max_embedding_size"`
	SearchLimit       int    `json:"search_limit"`
	EnableAutoBackup  bool   `json:"enable_auto_backup"`
}

// ExportConfig contains export settings
type ExportConfig struct {
	EnabledFormats    []string `json:"enabled_formats"`
	CompressionLevel  int      `json:"compression_level"`
	MaxExportSize     int64    `json:"max_export_size"`
	DefaultFormat     string   `json:"default_format"`
}

// ValidationConfig contains validation settings
type ValidationConfig struct {
	Enabled              bool   `json:"enabled"`
	StrictMode           bool   `json:"strict_mode"`
	AutoFixMinorIssues   bool   `json:"auto_fix_minor_issues"`
	MaxIssuesToReport    int    `json:"max_issues_to_report"`
}

// LoggingConfig contains logging settings
type LoggingConfig struct {
	Level           string `json:"level"`
	Format          string `json:"format"`
	EnableFileLog   bool   `json:"enable_file_log"`
	LogDirectory    string `json:"log_directory"`
	MaxLogSizeMB    int    `json:"max_log_size_mb"`
	RetentionDays   int    `json:"retention_days"`
}

// DefaultConfig returns the default configuration
func DefaultConfig() *Config {
	return &Config{
		API: APIConfig{
			Port:           8080,
			Host:           "localhost",
			ReadTimeout:    "10s",
			WriteTimeout:   "10s",
			MaxRequestSize: 100 * 1024 * 1024, // 100MB
			EnableCORS:     true,
			AllowedOrigins: []string{"*"},
		},
		Cache: CacheConfig{
			Enabled:         true,
			MaxSize:         1000,
			TTL:             "5m",
			CleanupInterval: "1m",
		},
		Memory: MemoryConfig{
			Enabled:          true,
			DatabasePath:     ".agent/memory/vectors.db",
			MaxEmbeddingSize: 1024,
			SearchLimit:      100,
			EnableAutoBackup: true,
		},
		Export: ExportConfig{
			EnabledFormats:   []string{"json", "csv", "markdown"},
			CompressionLevel: 6,
			MaxExportSize:    1000 * 1024 * 1024, // 1GB
			DefaultFormat:    "json",
		},
		Validation: ValidationConfig{
			Enabled:            true,
			StrictMode:         false,
			AutoFixMinorIssues: false,
			MaxIssuesToReport:  100,
		},
		Logging: LoggingConfig{
			Level:         "info",
			Format:        "json",
			EnableFileLog: true,
			LogDirectory:  ".agent/logs",
			MaxLogSizeMB:  100,
			RetentionDays: 30,
		},
	}
}

// LoadConfig loads configuration from a file
func LoadConfig(path string) (*Config, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return DefaultConfig(), nil
		}
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	config := DefaultConfig()
	if err := json.Unmarshal(data, config); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	return config, nil
}

// SaveConfig saves configuration to a file
func (c *Config) SaveConfig(path string) error {
	// Create directory if it doesn't exist
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	return ioutil.WriteFile(path, data, 0644)
}

// Validate checks if configuration is valid
func (c *Config) Validate() error {
	if c.API.Port < 1 || c.API.Port > 65535 {
		return fmt.Errorf("invalid API port: %d", c.API.Port)
	}

	if c.Cache.MaxSize < 1 {
		return fmt.Errorf("cache max size must be at least 1")
	}

	if _, err := time.ParseDuration(c.Cache.TTL); err != nil {
		return fmt.Errorf("invalid cache TTL: %s", c.Cache.TTL)
	}

	if _, err := time.ParseDuration(c.API.ReadTimeout); err != nil {
		return fmt.Errorf("invalid API read timeout: %s", c.API.ReadTimeout)
	}

	if _, err := time.ParseDuration(c.API.WriteTimeout); err != nil {
		return fmt.Errorf("invalid API write timeout: %s", c.API.WriteTimeout)
	}

	return nil
}

// GetReadTimeout returns the read timeout as a Duration
func (c *Config) GetReadTimeout() time.Duration {
	d, _ := time.ParseDuration(c.API.ReadTimeout)
	return d
}

// GetWriteTimeout returns the write timeout as a Duration
func (c *Config) GetWriteTimeout() time.Duration {
	d, _ := time.ParseDuration(c.API.WriteTimeout)
	return d
}

// GetCacheTTL returns the cache TTL as a Duration
func (c *Config) GetCacheTTL() time.Duration {
	d, _ := time.ParseDuration(c.Cache.TTL)
	return d
}

// Merge merges another config into this one (second config takes precedence)
func (c *Config) Merge(other *Config) {
	if other == nil {
		return
	}

	if other.API.Port != 0 {
		c.API.Port = other.API.Port
	}
	if other.API.Host != "" {
		c.API.Host = other.API.Host
	}
	if other.Cache.MaxSize != 0 {
		c.Cache.MaxSize = other.Cache.MaxSize
	}
}
