package datagovcy_test

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/yanakipre/bot/app/telegramsearch/internal/pkg/client/earthquakes"
	"github.com/yanakipre/bot/app/telegramsearch/internal/pkg/client/earthquakes/datagovcy"
	"github.com/yanakipre/bot/internal/resttooling"
)

func testLatestNEarthQuakes(t *testing.T, client earthquakes.Earthquaker) {

	tests := []struct {
		name         string
		n            int
		minMagnitude float32
		expectedErr  error
	}{
		{"Latest5Earthquakes", 5, 0, nil},
		{"Latest1Earthquakes", 1, 0, nil},
		{"Latest5EarthquakesMagn2", 5, 2, nil},
		{"Latest5EarthquakesMagnNegative", 5, -5, nil},
		{"Latest0Earthquakes", 0, 0, nil},
		{"LatestNegativeEarthquakes", -1, 0, datagovcy.ErrNegativeEarthquakes},
	}

	ctx := context.Background()

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			eqs, err := client.LatestNEarthquakes(ctx, tc.n, tc.minMagnitude)

			if tc.expectedErr != nil {
				if err == nil {
					t.Errorf("expected error %v, got nil", tc.expectedErr)
				} else if !errors.Is(err, tc.expectedErr) {
					t.Errorf("expected error %v, got %v", tc.expectedErr, err)
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if len(eqs) != tc.n {
				t.Errorf("expected %d earthquakes, got %d", tc.n, len(eqs))
			}

			for _, eq := range eqs {
				if eq.Magnitude < tc.minMagnitude {
					t.Errorf("expected min magnitude of %f, got %f", tc.minMagnitude, eq.Magnitude)
				}
			}
		})
	}

}

func TestIntegrationLatestEarthquakes(t *testing.T) {
	cfg := earthquakes.Config{}
	cfg.Default()
	client := datagovcy.NewClient(cfg)

	testLatestNEarthQuakes(t, client)
}

func TestLatestNEarthQuakes(t *testing.T) {
	content, err := os.ReadFile("./response.xml")
	if err != nil {
		t.Fatal(err)
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reader := bytes.NewReader(content)
		io.Copy(w, reader)
		if err != nil {
			t.Fatal(err)
		}
	}))

	client := datagovcy.NewClient(earthquakes.Config{
		ApiURL:        srv.URL,
		HTTPTransport: resttooling.DefaultTransportConfig(),
	})

	testLatestNEarthQuakes(t, client)
}
