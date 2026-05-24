package config

import (
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Server   ServerConfig   `yaml:"server"`
	Database DBConfig       `yaml:"database"`
	Log      LogConfig      `yaml:"log"`
	LLM      LLMConfig      `yaml:"llm"`
	Session  SessionConfig  `yaml:"session"`
	Analysis AnalysisConfig `yaml:"analysis"`
	Auth     AuthConfig     `yaml:"auth"`
}

type SessionConfig struct {
	MemoryWindowSize int `yaml:"memory_window_size"`
}

type AnalysisConfig struct {
	AutoTrigger bool `yaml:"auto_trigger"`
}

type AuthConfig struct {
	Username      string        `yaml:"username"`
	Password      string        `yaml:"password"`
	JWTSecret     string        `yaml:"jwt_secret"`
	AccessExpiry  time.Duration `yaml:"access_expiry"`
	RefreshExpiry time.Duration `yaml:"refresh_expiry"`
}

type ServerConfig struct {
	Host           string        `yaml:"host"`
	Port           int           `yaml:"port"`
	RequestTimeout time.Duration `yaml:"request_timeout"`
}

type DBConfig struct {
	Path string `yaml:"path"`
}

type LogConfig struct {
	Level string `yaml:"level"`
	File  string `yaml:"file"`
}

type LLMConfig struct {
	Provider string        `yaml:"provider"`
	BaseURL  string        `yaml:"base_url"`
	APIKey   string        `yaml:"api_key"`
	Model    string        `yaml:"model"`
	Timeout  time.Duration `yaml:"timeout"`
}

func Load(path string) (*Config, error) {
	cfg := &Config{
		Server: ServerConfig{
			Host:           "0.0.0.0",
			Port:           8080,
			RequestTimeout: 30 * time.Second,
		},
		Database: DBConfig{
			Path: "./talkent.db",
		},
		Log: LogConfig{
			Level: "info",
		},
		LLM: LLMConfig{
			Timeout: 30 * time.Second,
		},
		Session: SessionConfig{
			MemoryWindowSize: 10,
		},
		Analysis: AnalysisConfig{
			AutoTrigger: true,
		},
		Auth: AuthConfig{
			Username:      "admin",
			Password:      "admin",
			JWTSecret:     "change-me-in-production",
			AccessExpiry:  1 * time.Hour,
			RefreshExpiry: 7 * 24 * time.Hour,
		},
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read config file %s: %w", path, err)
	}

	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("parse config file: %w", err)
	}

	applyEnvOverrides(cfg)

	return cfg, nil
}

func applyEnvOverrides(cfg *Config) {
	if v := os.Getenv("TALKENT_SERVER_HOST"); v != "" {
		cfg.Server.Host = v
	}
	if v := os.Getenv("TALKENT_SERVER_PORT"); v != "" {
		fmt.Sscanf(v, "%d", &cfg.Server.Port)
	}
	if v := os.Getenv("TALKENT_DATABASE_PATH"); v != "" {
		cfg.Database.Path = v
	}
	if v := os.Getenv("TALKENT_LOG_LEVEL"); v != "" {
		cfg.Log.Level = v
	}
	if v := os.Getenv("TALKENT_LOG_FILE"); v != "" {
		cfg.Log.File = v
	}
	if v := os.Getenv("TALKENT_LLM_PROVIDER"); v != "" {
		cfg.LLM.Provider = v
	}
	if v := os.Getenv("TALKENT_LLM_BASE_URL"); v != "" {
		cfg.LLM.BaseURL = v
	}
	if v := os.Getenv("TALKENT_LLM_API_KEY"); v != "" {
		cfg.LLM.APIKey = v
	}
	if v := os.Getenv("TALKENT_LLM_MODEL"); v != "" {
		cfg.LLM.Model = v
	}
	if v := os.Getenv("TALKENT_ANALYSIS_AUTO_TRIGGER"); v != "" {
		cfg.Analysis.AutoTrigger = v == "true" || v == "1"
	}
	if v := os.Getenv("TALKENT_SESSION_MEMORY_WINDOW_SIZE"); v != "" {
		fmt.Sscanf(v, "%d", &cfg.Session.MemoryWindowSize)
	}
	if v := os.Getenv("TALKENT_AUTH_USERNAME"); v != "" {
		cfg.Auth.Username = v
	}
	if v := os.Getenv("TALKENT_AUTH_PASSWORD"); v != "" {
		cfg.Auth.Password = v
	}
	if v := os.Getenv("TALKENT_AUTH_JWT_SECRET"); v != "" {
		cfg.Auth.JWTSecret = v
	}
	if v := os.Getenv("TALKENT_AUTH_JWT_ACCESS_EXPIRY"); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			cfg.Auth.AccessExpiry = d
		}
	}
	if v := os.Getenv("TALKENT_AUTH_JWT_REFRESH_EXPIRY"); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			cfg.Auth.RefreshExpiry = d
		}
	}
}

func (c *Config) Addr() string {
	return fmt.Sprintf("%s:%d", c.Server.Host, c.Server.Port)
}

