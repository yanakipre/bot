package cyprusbus

import (
	"context"
	"fmt"
	"math"
	"sort"
	"time"
)

type BusFetcher interface {
	FetchBuses(ctx context.Context) ([]Bus, error)
}

type Dot struct {
	Lat  float64
	Long float64
}

type Bus struct {
	ID       string
	Route    string
	Position Dot
}

type Client struct {
	fetcher   BusFetcher
	sleepFunc func(time.Duration)
	timer     int
}

func NewClient(fetcher BusFetcher) *Client {
	return &Client{
		fetcher:   fetcher,
		sleepFunc: time.Sleep,
		timer:     20,
	}
}

func (c *Client) GetNearest(ctx context.Context, dot Dot) ([]Bus, error) {
	// Fetch Buses
	fetchedBuses, err := c.fetcher.FetchBuses(ctx)
	if err != nil {
		return nil, fmt.Errorf("initial fetch failed: %w", err)
	}

	// Calculate 1km bounding box
	latDelta := 1.0 / 111.0
	lonDelta := 1.0 / (111.0 * math.Cos(dot.Lat*math.Pi/180))

	latMin := dot.Lat - latDelta
	latMax := dot.Lat + latDelta
	lonMin := dot.Long - lonDelta
	lonMax := dot.Long + lonDelta

	// Filter fetched list if in bounds
	var initialBuses []Bus
	for _, bus := range fetchedBuses {
		if bus.Position.Lat >= latMin && bus.Position.Lat <= latMax &&
			bus.Position.Long >= lonMin && bus.Position.Long <= lonMax {
			initialBuses = append(initialBuses, bus)
		}
	}

	// Wait c.time seconds
	c.sleepFunc(time.Second * time.Duration(c.timer))

	// Second fetch
	newFetchedBuses, err := c.fetcher.FetchBuses(ctx)
	if err != nil {
		return nil, fmt.Errorf("second fetch failed: %w", err)
	}

	// Filter fetched list if in bounds
	var currentBuses []Bus
	for _, bus := range newFetchedBuses {
		if bus.Position.Lat >= latMin && bus.Position.Lat <= latMax &&
			bus.Position.Long >= lonMin && bus.Position.Long <= lonMax {
			currentBuses = append(currentBuses, bus)
		}
	}

	// Create current bus map
	currentMap := make(map[string]Bus)
	for _, bus := range currentBuses {
		currentMap[bus.ID] = bus
	}

	// Filter approaching buses
	var filteredBuses []Bus
	for _, initialBus := range initialBuses {
		currentBus, exists := currentMap[initialBus.ID]
		if !exists {
			filteredBuses = append(filteredBuses, currentBus)
			continue
		}

		// Compare distances
		initialDist := calculateDistance(dot, initialBus.Position)
		currentDist := calculateDistance(dot, currentBus.Position)
		if currentDist < initialDist {
			filteredBuses = append(filteredBuses, currentBus)
		}
	}

	// Sort by current distance
	sort.Slice(filteredBuses, func(i, j int) bool {
		return calculateDistance(dot, filteredBuses[i].Position) < calculateDistance(dot, filteredBuses[j].Position)
	})

	return filteredBuses, nil
}

func calculateDistance(a, b Dot) float64 {
	const earthRadius = 6371e3
	phi1 := a.Lat * math.Pi / 180
	phi2 := b.Lat * math.Pi / 180
	deltaPhi := (b.Lat - a.Lat) * math.Pi / 180
	deltaLambda := (b.Long - a.Long) * math.Pi / 180

	sinTerm := math.Sin(deltaPhi/2)*math.Sin(deltaPhi/2) +
		math.Cos(phi1)*math.Cos(phi2)*
			math.Sin(deltaLambda/2)*math.Sin(deltaLambda/2)
	c := 2 * math.Atan2(math.Sqrt(sinTerm), math.Sqrt(1-sinTerm))

	return earthRadius * c
}
