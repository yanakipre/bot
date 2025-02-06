package yanakipre

import (
	"context"
	"fmt"
	"math"
	"testing"
	"time"
)

type MockFetcher struct {
	responses [][]Bus
	errors    []error
	callCount int
}

func NewMockFetcher(responses [][]Bus, errors []error) *MockFetcher {
	return &MockFetcher{
		responses: responses,
		errors:    errors,
	}
}

func (m *MockFetcher) FetchBuses(ctx context.Context) ([]Bus, error) {
	if m.callCount >= len(m.responses) {
		return nil, fmt.Errorf("no more mock responses")
	}

	defer func() { m.callCount++ }()

	if m.errors[m.callCount] != nil {
		return nil, m.errors[m.callCount]
	}
	return m.responses[m.callCount], nil
}

func TestGetNearest_MovementScenarios(t *testing.T) {
	testDot := Dot{Lat: 34.707, Long: 33.022}

	tests := []struct {
		name          string
		mockResponses [][]Bus
		wantCount     int
	}{
		{
			name: "Approaching bus",
			mockResponses: [][]Bus{
				{ // First fetch
					{
						ID:    "bus1",
						Route: "100",
						Position: Dot{
							Lat:  testDot.Lat - 0.001,
							Long: testDot.Long - 0.001,
						},
					},
				},
				{ // Second fetch after c.timer
					{
						ID:    "bus1",
						Route: "100",
						Position: Dot{
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
			mockResponses: [][]Bus{
				{ // First fetch
					{
						ID:    "bus2",
						Route: "200",
						Position: Dot{
							Lat:  testDot.Lat - 0.0005,
							Long: testDot.Long - 0.0005,
						},
					},
				},
				{ // Second fetch after c.timer
					{
						ID:    "bus2",
						Route: "200",
						Position: Dot{
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
			mockResponses: [][]Bus{
				{ // First fetch
					{
						ID:    "bus3",
						Route: "300",
						Position: Dot{
							Lat:  testDot.Lat - 0.001,
							Long: testDot.Long - 0.001,
						},
					},
				},
				{ // Second fetch after c.timer
					{
						ID:    "bus3",
						Route: "300",
						Position: Dot{
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

			client := NewClient(mockFetcher)
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
		a         Dot
		b         Dot
		expected  float64
		tolerance float64
	}{
		{
			name:      "Same point",
			a:         Dot{Lat: 34.707, Long: 33.022},
			b:         Dot{Lat: 34.707, Long: 33.022},
			expected:  0,
			tolerance: 0.1,
		},
		{
			name:      "Short distance (143 meters)",
			a:         Dot{Lat: 34.707, Long: 33.022},
			b:         Dot{Lat: 34.708, Long: 33.023},
			expected:  143,
			tolerance: 1,
		},
		{
			name:      "Medium distance (1 km)",
			a:         Dot{Lat: 34.707, Long: 33.022},
			b:         Dot{Lat: 34.707 + 0.00899322, Long: 33.022},
			expected:  1000,
			tolerance: 1,
		},
		{
			name:      "Long distance (10 km)",
			a:         Dot{Lat: 34.707, Long: 33.022},
			b:         Dot{Lat: 34.707 + 0.0899322, Long: 33.022},
			expected:  10000,
			tolerance: 10,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := calculateDistance(tt.a, tt.b)
			if math.Abs(got-tt.expected) > tt.tolerance {
				t.Errorf(
					"calculateDistance(%v, %v) = %.2f meters, expected %.2f meters ±%.2f meters",
					tt.a, tt.b, got, tt.expected, tt.tolerance,
				)
			}
		})
	}
}
