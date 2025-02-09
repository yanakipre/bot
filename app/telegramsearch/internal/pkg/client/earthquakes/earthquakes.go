package earthquakes

import (
	"context"
	"time"
)

// ;)
type Earthquaker interface {
	LatestNEarthquakes(ctx context.Context, n int, minMagnitude float32) ([]Earthquake, error)
}

type Earthquake struct {
	Magnitude float32
	When      time.Time
	// Human-readable location of earthquake
	Location string
	// Geo coordinates
	Position Coordinate
}

// TODO: Extract to common package
type Coordinate struct {
	Latitude, Longitude float64
}
