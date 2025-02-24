package cyprusbus

import (
	"time"

	"github.com/yanakipre/bot/internal/encodingtooling"
	"github.com/yanakipre/bot/internal/resttooling"
)

type Config struct {
	// Timer is the time interval between 2 requests to the API.
	// It is used to understand the directions of the buses.
	Timer encodingtooling.Duration `yaml:"timer"`
	// BoxSizeMeters limits the size of the bounding box around the user's location.
	// Buses outside of this box are not considered.
	BoxSizeMeters float64            `yaml:"box_size_meters"`
	BaseURL       string             `yaml:"base_url"`
	HTTPTransport resttooling.Config `yaml:"http_transport"`
}

// this URL is retrieved from data.gov.cy.
const baseUrl string = "http://20.19.98.194:8328/Api"

func DefaultConfig() Config {
	tr := resttooling.DefaultTransportConfig()
	tr.ClientName = "cyprusbus"
	return Config{
		Timer:         encodingtooling.NewDuration(20 * time.Second),
		BoxSizeMeters: 1000,
		BaseURL:       baseUrl,
		HTTPTransport: tr,
	}
}
