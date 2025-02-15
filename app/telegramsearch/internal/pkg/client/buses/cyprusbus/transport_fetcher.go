package cyprusbus

import (
	"bytes"
	"context"
	"embed"
	"encoding/csv"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/MobilityData/gtfs-realtime-bindings/golang/gtfs"
	"github.com/yanakipre/bot/app/telegramsearch/internal/pkg/client/buses"
	"github.com/yanakipre/bot/internal/resttooling"
	"google.golang.org/protobuf/proto"
)

//go:embed data/routes.txt
var routeFS embed.FS

type protobufFetcher struct {
	baseURL    string
	httpClient *http.Client
	routeCache *routeCache
}

func newProtobufFetcher(cfg Config) *protobufFetcher {
	return &protobufFetcher{
		baseURL:    cfg.BaseURL,
		httpClient: resttooling.NewHTTPClientFromConfig(cfg.HTTPTransport),
	}
}

type routeCache struct {
	routes map[string]buses.Route
}

func (rc *routeCache) getRoutes(routeID string) (buses.Route, bool) {
	route, ok := rc.routes[routeID]
	return route, ok
}

func newRouteCache() (*routeCache, error) {
	rc := &routeCache{
		routes: make(map[string]buses.Route),
	}

	routesFile, err := routeFS.Open("data/routes.txt")
	if err != nil {
		return nil, fmt.Errorf("failed to open routes.txt: %w", err)
	}
	defer routesFile.Close()

	// Read the file content first to handle BOM (<feff>)
	data, err := io.ReadAll(routesFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read routes.txt: %w", err)
	}
	// Remove BOM if present
	data = bytes.TrimLeft(data, "\xef\xbb\xbf")

	reader := csv.NewReader(bytes.NewReader(data))

	// Read Header
	header, err := reader.Read()
	if err != nil {
		return nil, fmt.Errorf("failed to read routes.txt header: %w", err)
	}

	// Create index map for required fields
	fieldIndex := make(map[string]int)
	requiredFields := []string{"route_id", "route_short_name", "route_long_name"}

	for i, field := range header {
		fieldIndex[strings.TrimSpace(field)] = i
	}

	for _, field := range requiredFields {
		if _, ok := fieldIndex[field]; !ok {
			return nil, fmt.Errorf("required field is not present in routes.txt")
		}
	}

	// Read and parse routes
	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to read route record %w", err)
		}

		route := buses.Route{
			ID:        record[fieldIndex["route_id"]],
			ShortName: record[fieldIndex["route_short_name"]],
			LongName:  record[fieldIndex["route_long_name"]],
		}
		rc.routes[route.ID] = route
	}

	return rc, nil
}

func (f *protobufFetcher) FetchBuses(ctx context.Context) ([]buses.Bus, error) {

	// Initialize route cache if not already done
	if f.routeCache == nil {
		cache, err := newRouteCache()
		if err != nil {
			return nil, fmt.Errorf("failed to initialize route cache %w", err)
		}
		f.routeCache = cache
	}

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

	var fetchedBuses []buses.Bus
	for _, entity := range feed.Entity {
		if bp := entity.GetVehicle(); bp != nil {
			pos := bp.GetPosition()
			trip := bp.GetTrip()
			if pos != nil && trip != nil {
				routeID := trip.GetRouteId()
				route, ok := f.routeCache.getRoutes(routeID)
				if !ok {
					// If route is not found, route is routeID
					route = buses.Route{ID: routeID}
				}

				fetchedBuses = append(fetchedBuses, buses.Bus{
					ID:    bp.GetVehicle().GetId(),
					Route: route,
					Position: buses.Dot{
						Lat:  float64(pos.GetLatitude()),
						Long: float64(pos.GetLongitude()),
					},
				})
			}
		}
	}
	return fetchedBuses, nil
}
