package buses

import "context"

type BusFetcher interface {
	FetchBuses(ctx context.Context) ([]Bus, error)
	Ready(ctx context.Context) error
}

type Dot struct {
	Lat  float64
	Long float64
}

type Bus struct {
	// ID - vehicle identification #
	ID       string
	Route    Route
	Position Dot
}

type Route struct {
	ID        string
	ShortName string
	LongName  string
}
