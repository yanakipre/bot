package cyprusbus

import (
	"context"
	"github.com/yanakipre/bot/internal/testtooling"
	"testing"

	"github.com/yanakipre/bot/app/telegramsearch/internal/pkg/client/buses"
)

func TestFullWorkflow(t *testing.T) {
	testtooling.SetNewGlobalLoggerQuietly()
	cfg := DefaultConfig()
	client := NewClient(cfg)

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
				bus.Route.ShortName, bus.Route.LongName, bus.ID, bus.Route.ID, CalculateDistance(testDot, bus.Position))
		}
	} else {
		t.Log("No approaching buses found (might be expected behavior)")
	}
}
