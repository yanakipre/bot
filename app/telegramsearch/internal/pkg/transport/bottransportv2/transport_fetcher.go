package yanakipre

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/MobilityData/gtfs-realtime-bindings/golang/gtfs"
	"google.golang.org/protobuf/proto"
)

type ProtobufFetcher struct {
	baseURL    string
	httpClient *http.Client
}

func NewProtobufFetcher(baseURL string) *ProtobufFetcher {
	return &ProtobufFetcher{
		baseURL:    baseURL,
		httpClient: &http.Client{Timeout: 10 * time.Second},
	}
}

func (f *ProtobufFetcher) FetchBuses(ctx context.Context) ([]Bus, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", f.baseURL+"/api/gtfs-realtime", nil)
	if err != nil {
		return nil, err
	}

	resp, err := f.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	feed := &gtfs.FeedMessage{}
	if err := proto.Unmarshal(data, feed); err != nil {
		return nil, err
	}

	var buses []Bus
	for _, entity := range feed.Entity {
		if vp := entity.GetVehicle(); vp != nil {
			pos := vp.GetPosition()
			trip := vp.GetTrip()
			if pos != nil && trip != nil {
				buses = append(buses, Bus{
					ID:    vp.GetVehicle().GetId(),
					Route: trip.GetRouteId(),
					Position: Dot{
						Lat:  float64(pos.GetLatitude()),
						Long: float64(pos.GetLongitude()),
					},
				})
			}
		}
	}
	return buses, nil
}
