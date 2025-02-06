package yanakipre

import (
	"context"
	"testing"
)

func TestFullWorkflow(t *testing.T) {
	fetcher := NewProtobufFetcher(baseURL)
	client := NewClient(fetcher)
	client.timer = 20

	testDot := Dot{
		Lat:  34.700474470158184,
		Long: 33.100647034953774,
	}

	buses, err := client.GetNearest(context.Background(), testDot)

	if err != nil {
		t.Fatalf("Full workflow failed: %v", err)
	}

	if len(buses) > 0 {
		t.Logf("Found %d buses approaching the area", len(buses))
		for i, bus := range buses {
			t.Logf("Bus %d: %s (Route %s) - Distance: %.0fm",
				i+1, bus.ID, bus.Route, calculateDistance(testDot, bus.Position))
		}
	} else {
		t.Log("No approaching buses found (this might be expected behavior)")
	}
}
