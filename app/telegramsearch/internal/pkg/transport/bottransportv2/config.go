package bottransportv2

// for userbot?
import (
	"errors"

	"github.com/yanakipre/bot/internal/secret"
)

type Config struct {
	AppID   secret.Value[int]
	AppHash secret.String
}

func (c *Config) Validate() error {
	if c.AppID.Unmask() == 0 {
		return errors.New("app_id is required")
	}
	if c.AppHash.Unmask() == "" {
		return errors.New("app_hash is required")
	}
	return nil
}

func DefaultConfig() Config {
	return Config{
		AppID:   secret.NewValue[int](24144218),
		AppHash: secret.NewString("b1602e8d49775f7a212037c00e29ed6d"), // ?
	}
}
