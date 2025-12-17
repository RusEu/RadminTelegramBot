package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Config represents the application configuration
type Config struct {
	Telegram   TelegramConfig   `yaml:"telegram"`
	Security   SecurityConfig   `yaml:"security"`
	Server     ServerConfig     `yaml:"server"`
	Monitoring MonitoringConfig `yaml:"monitoring"`
}

// TelegramConfig holds Telegram bot settings
type TelegramConfig struct {
	BotToken     string  `yaml:"bot_token"`
	AllowedUsers []int64 `yaml:"allowed_users"`
	WebhookURL   string  `yaml:"webhook_url,omitempty"`
	WebhookPort  int     `yaml:"webhook_port,omitempty"`
}

// SecurityConfig holds security-related settings
type SecurityConfig struct {
	AdminPassword    string `yaml:"admin_password"`
	SessionTimeout   int    `yaml:"session_timeout"`
	MaxFileSize      int64  `yaml:"max_file_size"`
	RateLimitWindow  int    `yaml:"rate_limit_window"`
	RateLimitCommands int   `yaml:"rate_limit_commands"`
}

// ServerConfig holds server-specific settings
type ServerConfig struct {
	Name       string `yaml:"name"`
	Timezone   string `yaml:"timezone"`
	LogLevel   string `yaml:"log_level"`
	WorkingDir string `yaml:"working_dir"`
	LogFile    string `yaml:"log_file,omitempty"`
}

// MonitoringConfig holds monitoring thresholds and settings
type MonitoringConfig struct {
	CPUAlertThreshold    float64 `yaml:"cpu_alert_threshold"`
	MemoryAlertThreshold float64 `yaml:"memory_alert_threshold"`
	DiskAlertThreshold   float64 `yaml:"disk_alert_threshold"`
	MonitoringInterval   int     `yaml:"monitoring_interval"`
	AlertCooldown        int     `yaml:"alert_cooldown"`
}

// LoadConfig loads configuration from a YAML file
func LoadConfig(configPath string) (*Config, error) {
	// Check if config file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("configuration file not found: %s", configPath)
	}

	// Read config file
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	// Parse YAML
	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	// Set defaults
	config.setDefaults()

	// Expand environment variables
	config.expandEnvVars()

	return &config, nil
}

// setDefaults sets default values for configuration options
func (c *Config) setDefaults() {
	if c.Security.SessionTimeout == 0 {
		c.Security.SessionTimeout = 3600 // 1 hour
	}
	if c.Security.MaxFileSize == 0 {
		c.Security.MaxFileSize = 52428800 // 50MB
	}
	if c.Security.RateLimitWindow == 0 {
		c.Security.RateLimitWindow = 60 // 1 minute
	}
	if c.Security.RateLimitCommands == 0 {
		c.Security.RateLimitCommands = 10 // 10 commands per minute
	}
	if c.Server.LogLevel == "" {
		c.Server.LogLevel = "info"
	}
	if c.Server.Name == "" {
		c.Server.Name = "Server"
	}
	if c.Server.Timezone == "" {
		c.Server.Timezone = "UTC"
	}
	if c.Server.WorkingDir == "" {
		c.Server.WorkingDir = "/tmp"
	}
	if c.Monitoring.CPUAlertThreshold == 0 {
		c.Monitoring.CPUAlertThreshold = 85
	}
	if c.Monitoring.MemoryAlertThreshold == 0 {
		c.Monitoring.MemoryAlertThreshold = 90
	}
	if c.Monitoring.DiskAlertThreshold == 0 {
		c.Monitoring.DiskAlertThreshold = 95
	}
	if c.Monitoring.MonitoringInterval == 0 {
		c.Monitoring.MonitoringInterval = 300 // 5 minutes
	}
	if c.Monitoring.AlertCooldown == 0 {
		c.Monitoring.AlertCooldown = 1800 // 30 minutes
	}
}

// expandEnvVars expands environment variables in configuration
func (c *Config) expandEnvVars() {
	c.Telegram.BotToken = os.ExpandEnv(c.Telegram.BotToken)
	c.Security.AdminPassword = os.ExpandEnv(c.Security.AdminPassword)
	c.Server.WorkingDir = os.ExpandEnv(c.Server.WorkingDir)
	c.Server.LogFile = os.ExpandEnv(c.Server.LogFile)
}

// Validate validates the configuration
func (c *Config) Validate() error {
	if c.Telegram.BotToken == "" {
		return fmt.Errorf("telegram bot token is required")
	}
	
	if len(c.Telegram.AllowedUsers) == 0 {
		return fmt.Errorf("at least one allowed user must be specified")
	}

	if c.Security.AdminPassword == "" {
		return fmt.Errorf("admin password is required")
	}

	if c.Security.SessionTimeout < 60 {
		return fmt.Errorf("session timeout must be at least 60 seconds")
	}

	if c.Security.MaxFileSize < 1024 {
		return fmt.Errorf("max file size must be at least 1024 bytes")
	}

	if c.Server.WorkingDir != "" {
		if !filepath.IsAbs(c.Server.WorkingDir) {
			return fmt.Errorf("working directory must be an absolute path")
		}
	}

	// Validate log level
	validLogLevels := []string{"debug", "info", "warn", "error"}
	validLevel := false
	for _, level := range validLogLevels {
		if c.Server.LogLevel == level {
			validLevel = true
			break
		}
	}
	if !validLevel {
		return fmt.Errorf("invalid log level: %s (valid: debug, info, warn, error)", c.Server.LogLevel)
	}

	return nil
}

// SaveConfig saves the configuration to a file
func (c *Config) SaveConfig(configPath string) error {
	data, err := yaml.Marshal(c)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	return os.WriteFile(configPath, data, 0600)
}