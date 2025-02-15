package buses_test

import (
	"context"
	"testing"

	"github.com/yanakipre/bot/app/telegramsearch/internal/pkg/client/buses"
	"github.com/yanakipre/bot/app/telegramsearch/internal/pkg/client/buses/cyprusbus"
)

func TestFullWorkflow(t *testing.T) {
	cfg := buses.DefaultConfig()
	fetcher := cyprusbus.NewProtobufFetcher(cfg)
	client := cyprusbus.NewClient(cfg, fetcher)

	testDot := buses.Dot{
		Lat:  34.700474470158184,
		Long: 33.100647034953774,
	}

	nearestBuses, err := client.GetNearest(context.Background(), testDot)

	if err != nil {
		t.Fatalf("Full workflow failed: %v", err)
	}

	if len(nearestBuses) > 0 {
		t.Logf("Found %d buses approaching the area", len(nearestBuses))
		for _, bus := range nearestBuses {
			t.Logf("Route %s %s (ID: %s, Route: %s) - Distance: %1.fm\n",
				bus.Route.ShortName, bus.Route.LongName, bus.ID, bus.Route.ID, cyprusbus.CalculateDistance(testDot, bus.Position))
		}
	} else {
		t.Log("No approaching buses found (might be expected behavior)")
	}
}
