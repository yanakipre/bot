package buses

import (
	"time"

	"github.com/yanakipre/bot/internal/encodingtooling"
	"github.com/yanakipre/bot/internal/resttooling"
)

type Config struct {
	Timer         encodingtooling.Duration
	BoxSizeMeters float64
	BaseURL       string
	HTTPTransport resttooling.Config
}

const BaseUrl string = "http://20.19.98.194:8328/Api"

func DefaultConfig() Config {
	return Config{
		Timer:         encodingtooling.NewDuration(20 * time.Second),
		BoxSizeMeters: 1000,
		BaseURL:       BaseUrl,
		HTTPTransport: resttooling.DefaultTransportConfig(),
	}
}
