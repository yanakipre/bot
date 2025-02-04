package chdb

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCHDB_SplitQuery(t *testing.T) {
	query := `
SELECT * FROM events;



-- some comment
SELECT * FROM consumption_events;
`
	queries := splitQuery(query)
	require.Len(t, queries, 2)
	require.EqualValues(
		t,
		[]string{"SELECT * FROM events", "-- some comment\nSELECT * FROM consumption_events"},
		queries,
	)
}
