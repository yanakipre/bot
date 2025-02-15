package cyprusbus

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
