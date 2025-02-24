package cyprusbus

import (
	"context"
	"fmt"
	"math"
	"sort"
	"time"

	"github.com/yanakipre/bot/internal/buses"
)

type Client struct {
	fetcher   buses.BusFetcher
	sleepFunc func(time.Duration)
	timer     time.Duration
	boxSize   float64
}

func NewClient(cfg Config) *Client {
	return &Client{
		fetcher:   newProtobufFetcher(cfg),
		sleepFunc: time.Sleep,
		timer:     cfg.Timer.Duration,
		boxSize:   cfg.BoxSizeMeters,
	}
}

func (c *Client) GetNearest(ctx context.Context, dot buses.Dot) ([]buses.Bus, error) {
	// Fetch Buses
	fetchedBuses, err := c.fetcher.FetchBuses(ctx)
	if err != nil {
		return nil, fmt.Errorf("initial fetch failed: %w", err)
	}

	// Calculate 1km bounding box
	latDelta := c.boxSize / 111000.0
	longDelta := c.boxSize / (111000.0 * math.Cos(dot.Lat*math.Pi/180))

	latMin := dot.Lat - latDelta
	latMax := dot.Lat + latDelta
	longMin := dot.Long - longDelta
	longMax := dot.Long + longDelta

	// Filter fetched list if in bounds
	var initialBuses []buses.Bus
	for _, bus := range fetchedBuses {
		if bus.Position.Lat >= latMin && bus.Position.Lat <= latMax &&
			bus.Position.Long >= longMin && bus.Position.Long <= longMax {
			initialBuses = append(initialBuses, bus)
		}
	}

	// Wait c.time seconds
	c.sleepFunc(c.timer)

	// Second fetch
	secondFetchedBuses, err := c.fetcher.FetchBuses(ctx)
	if err != nil {
		return nil, fmt.Errorf("second fetch failed: %w", err)
	}

	// Filter fetched list if in bounds
	var currentBuses []buses.Bus
	for _, bus := range secondFetchedBuses {
		if bus.Position.Lat >= latMin && bus.Position.Lat <= latMax &&
			bus.Position.Long >= longMin && bus.Position.Long <= longMax {
			currentBuses = append(currentBuses, bus)
		}
	}

	// Create current bus map
	initialMap := make(map[string]buses.Bus)
	for _, bus := range initialBuses {
		initialMap[bus.ID] = bus
	}

	// Filter approaching buses
	var filteredBuses []buses.Bus
	for _, currentBus := range currentBuses {
		initialBus, exists := initialMap[currentBus.ID]
		if !exists {
			//filteredBuses = append(filteredBuses, currentBus) // This line may cause adding filteredBuses with a departing Bus, if GPS data is flowed in a margin larger than a tolerance treshold
			continue
		}

		// Compare distances
		initialDist := CalculateDistance(dot, initialBus.Position)
		currentDist := CalculateDistance(dot, currentBus.Position)
		if currentDist < initialDist {
			filteredBuses = append(filteredBuses, currentBus)
		}
	}

	// Sort by current distance
	sort.Slice(filteredBuses, func(i, j int) bool {
		return CalculateDistance(dot, filteredBuses[i].Position) < CalculateDistance(dot, filteredBuses[j].Position)
	})

	return filteredBuses, nil
}

func CalculateDistance(a, b buses.Dot) float64 {
	const earthRadius = 6378137
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
