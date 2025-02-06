package cyprusbus

import (
	"context"
	"testing"
	"time"
)

const baseURL string = "http://20.19.98.194:8328/Api"

func TestProtobufFetcher(t *testing.T) {
	fetcher := NewProtobufFetcher(baseURL)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	t.Run("should fetch bus positions", func(t *testing.T) {
		buses, err := fetcher.FetchBuses(ctx)
		if err != nil {
			if isConnectionError(err) {
				t.Skip("Skipping test due to connection error:", err)
			}
			t.Fatalf("Failed to fetch buses: %v", err)
		}

		if len(buses) == 0 {
			t.Log("Warning: Received empty bus list")
			return
		}

		for i, bus := range buses {
			// t.Logf("Bus %v has route %v and position %+v\n", bus.ID, bus.Route, bus.Position)
			if bus.ID == "" {
				t.Errorf("Bus %d has empty ID", i)
			}
			if bus.Route == "" {
				t.Errorf("Bus %d has empty Route", i)
			}
			// if bus.Position.Lat == 0 || bus.Position.Long == 0 {
			// 	t.Errorf("Bus %d has invalid coordinates: %+v", i, bus.Position)
			// }
		}
	})

	t.Run("should handle server errors", func(t *testing.T) {
		invalidFetcher := NewProtobufFetcher(baseURL + "/invalid-endpoint")
		_, err := invalidFetcher.FetchBuses(ctx)

		if err == nil {
			t.Error("Expected error for invalid endpoint, got nil")
		}
	})
}

func isConnectionError(err error) bool {
	if err == nil {
		return false
	}
	return err.Error() == "context deadline exceeded" ||
		err.Error() == "connection refused"
}
