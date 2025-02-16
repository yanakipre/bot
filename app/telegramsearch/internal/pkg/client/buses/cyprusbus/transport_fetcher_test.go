package cyprusbus

import (
	"bytes"
	"context"
	"io"
	"strings"
	"testing"
	"time"
)

func TestRoutesFileEmbedding(t *testing.T) {
	// Test that we can open the embedded file
	file, err := routeFS.Open("data/routes.txt")
	if err != nil {
		t.Fatalf("Failed to open embedded routes.txt: %v", err)
	}
	defer file.Close()

	// Read the content to verify it's a valid CSV
	data, err := io.ReadAll(file)
	if err != nil {
		t.Fatalf("Failed to read embedded routes.txt: %v", err)
	}

	// Verify the file contains expected header
	if len(data) == 0 {
		t.Error("Embedded routes.txt is empty")
	}

	// Check if the file starts with a valid CSV header
	lines := bytes.Split(data, []byte("\n"))
	if len(lines) < 2 { // At least header and one route
		t.Error("routes.txt should contain at least a header and one route")
	}

	// Check if the file has all the required fields
	expectedFields := []string{"route_id", "route_short_name", "route_long_name"}
	header := string(lines[0])
	for _, field := range expectedFields {
		if !strings.Contains(header, field) {
			t.Errorf("Header missing required field: %s", field)
		}
	}
}

func TestNewRouteCache(t *testing.T) {
	t.Run("successful route cache creation", func(t *testing.T) {
		cache, err := newRouteCache()
		if err != nil {
			t.Fatalf("Failed to create route cache: %v", err)
		}

		if cache == nil {
			t.Fatal("Expected non-nil route cache")
		}

		if cache.routes == nil {
			t.Error("Expected non-nil routes map")
		}

		// Verify that at least one route was loaded
		if len(cache.routes) == 0 {
			t.Error("Expected at least one route to be loaded")
		}

		// Test route retrieval
		for routeID, route := range cache.routes {
			if route.ID != routeID {
				t.Errorf("Route ID mismatch: got %s, want %s", route.ID, routeID)
			}
			if route.ShortName == "" {
				t.Errorf("Route %s has empty short name", routeID)
			}
			if route.LongName == "" {
				t.Errorf("Route %s has empty long name", routeID)
			}
		}
	})

	t.Run("route cache getRoutes", func(t *testing.T) {
		cache, err := newRouteCache()
		if err != nil {
			t.Fatalf("Failed to create route cache: %v", err)
		}

		// Test existing route
		for routeID := range cache.routes {
			route, exists := cache.getRoutes(routeID)
			if !exists {
				t.Errorf("Expected route %s to exist", routeID)
			}
			if route.ID != routeID {
				t.Errorf("Route ID mismatch: got %s, want %s", route.ID, routeID)
			}
		}

		// Test non-existent route
		_, exists := cache.getRoutes("non-existent-route")
		if exists {
			t.Error("Expected non-existent route to return false")
		}
	})
}

func TestProtobufFetcher(t *testing.T) {
	cfg := DefaultConfig()

	fetcher := newProtobufFetcher(cfg)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	err := fetcher.Ready(ctx)
	if err != nil {
		t.Fatalf("Failed to initialize the routes cache: %v", err)
	}

	t.Run("should fetch bus positions", func(t *testing.T) {
		fetchedBuses, err := fetcher.FetchBuses(ctx)
		if err != nil {
			if isConnectionError(err) {
				t.Skip("Skipping test due to connection error:", err)
			}
			t.Fatalf("Failed to fetch buses: %v", err)
		}

		if len(fetchedBuses) == 0 {
			t.Log("Warning: Received empty bus list")
			return
		}

		for i, bus := range fetchedBuses {
			if bus.ID == "" {
				t.Errorf("Bus %d has empty ID", i)
			}
		}
	})

	t.Run("should handle server errors", func(t *testing.T) {
		cfg.BaseURL += "/invalid-endpoint"
		invalidFetcher := newProtobufFetcher(cfg)
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
