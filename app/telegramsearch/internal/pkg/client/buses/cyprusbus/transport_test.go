package cyprusbus

import (
	"context"
	"fmt"
	"math"
	"testing"
	"time"

	"github.com/yanakipre/bot/app/telegramsearch/internal/pkg/client/buses"
)

type MockFetcher struct {
	responses [][]buses.Bus
	errors    []error
	callCount int
}

func NewMockFetcher(responses [][]buses.Bus, errors []error) *MockFetcher {
	return &MockFetcher{
		responses: responses,
		errors:    errors,
	}
}

func (m *MockFetcher) FetchBuses(ctx context.Context) ([]buses.Bus, error) {
	if m.callCount >= len(m.responses) {
		return nil, fmt.Errorf("no more mock responses")
	}

	defer func() { m.callCount++ }()

	if m.errors[m.callCount] != nil {
		return nil, m.errors[m.callCount]
	}
	return m.responses[m.callCount], nil
}

// Ready implements check for readinesschecker
func (pf *MockFetcher) Ready(ctx context.Context) error {
	return nil
}

func TestGetNearest_MovementScenarios(t *testing.T) {
	testDot := buses.Dot{Lat: 34.707, Long: 33.022}

	tests := []struct {
		name          string
		mockResponses [][]buses.Bus
		wantCount     int
	}{
		{
			name: "Approaching bus",
			mockResponses: [][]buses.Bus{
				{ // First fetch
					{
						ID:    "bus1",
						Route: buses.Route{ID: "100"},
						Position: buses.Dot{
							Lat:  testDot.Lat - 0.001,
							Long: testDot.Long - 0.001,
						},
					},
				},
				{ // Second fetch after c.timer
					{
						ID:    "bus1",
						Route: buses.Route{ID: "100"},
						Position: buses.Dot{
							Lat:  testDot.Lat - 0.0005, // Moving closer
							Long: testDot.Long - 0.0005,
						},
					},
				},
			},
			wantCount: 1,
		},
		{
			name: "Departing bus",
			mockResponses: [][]buses.Bus{
				{ // First fetch
					{
						ID:    "bus2",
						Route: buses.Route{ID: "200"},
						Position: buses.Dot{
							Lat:  testDot.Lat - 0.0005,
							Long: testDot.Long - 0.0005,
						},
					},
				},
				{ // Second fetch after c.timer
					{
						ID:    "bus2",
						Route: buses.Route{ID: "200"},
						Position: buses.Dot{
							Lat:  testDot.Lat - 0.001, // Moving away
							Long: testDot.Long - 0.001,
						},
					},
				},
			},
			wantCount: 0,
		},
		{
			name: "Stationary bus",
			mockResponses: [][]buses.Bus{
				{ // First fetch
					{
						ID:    "bus3",
						Route: buses.Route{ID: "300"},
						Position: buses.Dot{
							Lat:  testDot.Lat - 0.001,
							Long: testDot.Long - 0.001,
						},
					},
				},
				{ // Second fetch after c.timer
					{
						ID:    "bus3",
						Route: buses.Route{ID: "300"},
						Position: buses.Dot{
							Lat:  testDot.Lat - 0.001, // Stationary
							Long: testDot.Long - 0.001,
						},
					},
				},
			},
			wantCount: 0, // Zero distance change
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockFetcher := NewMockFetcher(
				tt.mockResponses,
				make([]error, len(tt.mockResponses)),
			)

			cfg := DefaultConfig()

			client := NewClient(cfg)
			client.fetcher = mockFetcher
			client.sleepFunc = func(time.Duration) {}

			result, err := client.GetNearest(context.Background(), testDot)
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			if len(result) != tt.wantCount {
				t.Errorf("Expected %d buses, got %d", tt.wantCount, len(result))
			}
		})
	}
}

func TestCalculateDistance(t *testing.T) {
	tests := []struct {
		name      string
		a         buses.Dot
		b         buses.Dot
		expected  float64
		tolerance float64
	}{
		{
			name:      "Same point",
			a:         buses.Dot{Lat: 34.707, Long: 33.022},
			b:         buses.Dot{Lat: 34.707, Long: 33.022},
			expected:  0,
			tolerance: 0.1,
		},
		{
			name:      "Short distance (143 meters)",
			a:         buses.Dot{Lat: 34.707, Long: 33.022},
			b:         buses.Dot{Lat: 34.708, Long: 33.023},
			expected:  144,
			tolerance: 1,
		},
		{
			name:      "Medium distance (1 km)",
			a:         buses.Dot{Lat: 34.707, Long: 33.022},
			b:         buses.Dot{Lat: 34.707 + 0.00899322, Long: 33.022},
			expected:  1001,
			tolerance: 1,
		},
		{
			name:      "Long distance (10 km)",
			a:         buses.Dot{Lat: 34.707, Long: 33.022},
			b:         buses.Dot{Lat: 34.707 + 0.0899322, Long: 33.022},
			expected:  10011,
			tolerance: 10,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CalculateDistance(tt.a, tt.b)
			if math.Abs(got-tt.expected) > tt.tolerance {
				t.Errorf(
					"calculateDistance(%v, %v) = %.2f meters, expected %.2f meters ±%.2f meters",
					tt.a, tt.b, got, tt.expected, tt.tolerance,
				)
			}
		})
	}
}
