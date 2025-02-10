package earthquakes

import (
	"github.com/yanakipre/bot/internal/resttooling"
)

type Config struct {
	ApiURL        string
	HTTPTransport resttooling.Config
}

const defaultApiURL = "http://www.gsd-seismology.org.cy/events/feed.rss"

func (c *Config) Default() {
	transport := resttooling.DefaultTransportConfig()
	transport.ClientName = "earthquakes"
	*c = Config{
		ApiURL:        defaultApiURL,
		HTTPTransport: transport,
	}
}
