package encodingtooling

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

var jsonData = []struct {
	raw  time.Duration
	JSON string
	YAML string
}{
	{raw: time.Millisecond, JSON: "\"1ms\"", YAML: "1ms"},
	{raw: time.Second, JSON: "\"1s\"", YAML: "1s"},
	{raw: 0, JSON: "\"0s\"", YAML: "0s"},
	{raw: time.Hour + time.Minute + time.Second, JSON: "\"1h1m1s\"", YAML: "1h1m1s"},
	{raw: -time.Second, JSON: "\"-1s\"", YAML: "-1s"},
}

func TestDuration_MarshalYAML(t *testing.T) {
	for _, tt := range jsonData {
		t.Run(fmt.Sprintf("marshal %s", tt.JSON), func(t *testing.T) {
			d := Duration{
				Duration: tt.raw,
			}
			got, err := d.MarshalYAML()
			require.NoError(t, err)
			require.Equal(t, tt.YAML, got.(string))
		})
	}
}

func TestDuration_UnmarshalYAML(t *testing.T) {
	for _, tt := range jsonData {
		result := tt.YAML
		t.Run(fmt.Sprintf("unmarshal %s", tt.YAML), func(t *testing.T) {
			var d Duration
			err := d.UnmarshalYAML(func(i any) error {
				r, ok := i.(*string)
				if !ok {
					panic("cannot cast to string pointer")
				}
				*r = result
				return nil
			})
			require.NoError(t, err)
			require.Equal(t, tt.raw, d.Duration)
		})
	}
}

func TestDuration_MarshalJSON(t *testing.T) {
	for _, tt := range jsonData {
		t.Run("marshal test", func(t *testing.T) {
			d := Duration{
				Duration: tt.raw,
			}
			got, err := d.MarshalJSON()
			require.NoError(t, err)
			require.Equal(t, tt.JSON, string(got))
		})
	}
}

func TestDuration_UnmarshalJSON(t *testing.T) {
	for _, tt := range jsonData {
		t.Run("unmarshal test", func(t *testing.T) {
			var d Duration
			err := d.UnmarshalJSON([]byte(tt.JSON))
			require.NoError(t, err)
			require.Equal(t, tt.raw, d.Duration)
		})
	}
}
