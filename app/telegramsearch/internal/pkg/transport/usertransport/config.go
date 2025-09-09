package usertransport

import (
	"errors"
	"os"
	"path/filepath"

	"github.com/yanakipre/bot/internal/secret"
)

// Config represents configuration for user transport
type Config struct {
	SessionStoragePath string                `yaml:"session_storage_path"`
	AppID              secret.Value[int]     `yaml:"appid"`
	AppHash            secret.String         `yaml:"apphash"`
	Phone              secret.String         `yaml:"phone"`
	ChatIDs            []secret.Value[int64] `yaml:"chat_ids"`
}

// DefaultConfig returns default configuration for user transport
func DefaultConfig() Config {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		homeDir = "."
	}
	return Config{
		SessionStoragePath: filepath.Join(homeDir, ".telegramsearch", "session"),
		AppID:              secret.NewValue[int](10),
		AppHash:            secret.NewString("apphash"),
		Phone:              secret.NewString("phone"),
		ChatIDs:            []secret.Value[int64]{},
	}
}

// Validate validates the configuration
func (c *Config) Validate() error {
	var errs []error

	if c.AppID.Unmask() == 0 {
		errs = append(errs, errors.New("app_id is required"))
	}

	if c.AppHash.Unmask() == "" {
		errs = append(errs, errors.New("app_hash is required"))
	}

	if c.Phone.Unmask() == "" {
		errs = append(errs, errors.New("phone is required"))
	}

	if c.SessionStoragePath == "" {
		errs = append(errs, errors.New("session_storage_path is required"))
	}

	if len(c.ChatIDs) == 0 {
		errs = append(errs, errors.New("chat_ids is required"))
	}

	return errors.Join(errs...)
}
