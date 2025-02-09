package earthquakes

import (
	"time"

	"github.com/yanakipre/bot/internal/encodingtooling"
	"github.com/yanakipre/bot/internal/resttooling"
)

type Config struct {
	ApiURL        string
	Timeout       encodingtooling.Duration
	HTTPTransport resttooling.Config
}

const defaultApiURL = "http://www.gsd-seismology.org.cy/events/feed.rss"

func DefaultConfig() Config {
	return Config{
		ApiURL:        defaultApiURL,
		Timeout:       encodingtooling.NewDuration(100 * time.Second),
		HTTPTransport: resttooling.DefaultTransportConfig(),
	}
}
