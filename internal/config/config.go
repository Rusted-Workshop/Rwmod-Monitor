package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

type S3Config struct {
	Endpoint  string `json:"endpoint"`
	AccessKey string `json:"access_key"`
	SecretKey string `json:"secret_key"`
	Bucket    string `json:"bucket"`
}

type Config struct {
	MonitorDir    string   `json:"monitor_dir"`
	DelayMinutes  int      `json:"delay_minutes"`
	MaxRetries    int      `json:"max_retries"`
	S3            S3Config `json:"s3"`
}

func GetConfigPath() (string, error) {
	execPath, err := os.Executable()
	if err != nil {
		return "", err
	}
	execDir := filepath.Dir(execPath)
	return filepath.Join(execDir, "config.json"), nil
}

func CreateTemplate() (*Config, error) {
	template := &Config{
		MonitorDir:   "/path/to/monitor",
		DelayMinutes: 5,
		MaxRetries:   3,
		S3: S3Config{
			Endpoint:  "https://s3.example.com",
			AccessKey: "your-access-key-id",
			SecretKey: "your-secret-key",
			Bucket:    "your-bucket-name",
		},
	}
	return template, nil
}

func Load() (*Config, error) {
	configPath, err := GetConfigPath()
	if err != nil {
		return nil, fmt.Errorf("failed to get config path: %w", err)
	}

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return createDefaultConfig(configPath)
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config: %w", err)
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	return &cfg, nil
}

func createDefaultConfig(configPath string) (*Config, error) {
	cfg, err := CreateTemplate()
	if err != nil {
		return nil, err
	}

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return nil, fmt.Errorf("failed to write config: %w", err)
	}

	return cfg, nil
}

func (c *Config) Validate() error {
	if c.MonitorDir == "" || c.MonitorDir == "/path/to/monitor" {
		return fmt.Errorf("monitor_dir not configured")
	}
	if c.S3.Endpoint == "" || c.S3.Endpoint == "https://s3.example.com" {
		return fmt.Errorf("S3 endpoint not configured")
	}
	if c.S3.AccessKey == "" || c.S3.AccessKey == "your-access-key-id" {
		return fmt.Errorf("S3 access key not configured")
	}
	if c.S3.SecretKey == "" || c.S3.SecretKey == "your-secret-key" {
		return fmt.Errorf("S3 secret key not configured")
	}
	if c.S3.Bucket == "" || c.S3.Bucket == "your-bucket-name" {
		return fmt.Errorf("S3 bucket not configured")
	}
	return nil
}
