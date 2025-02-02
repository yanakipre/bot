package encodingtooling

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

var testSizeData = []struct {
	raw  int64
	JSON string
	YAML string
}{
	{raw: 0, JSON: "\"0B\"", YAML: "0B"},
	{raw: 6272, JSON: "\"6.125KiB\"", YAML: "6.125KiB"},
	{raw: 1024 * 1024 * 10, JSON: "\"10MiB\"", YAML: "10MiB"},
	{raw: 1024*1024*10 + 10*1024 + 245, JSON: "\"10.01MiB\"", YAML: "10.01MiB"},
	{raw: 123 * 1024 * 1024 * 1024, JSON: "\"123GiB\"", YAML: "123GiB"},
	{raw: 123 * 1024 * 1024 * 1024 * 1024, JSON: "\"123TiB\"", YAML: "123TiB"},
}

func TestSize_MarshalYAML(t *testing.T) {
	for _, tt := range testSizeData {
		t.Run(fmt.Sprintf("marshal %s", tt.JSON), func(t *testing.T) {
			d := Size{
				Size: tt.raw,
			}
			got, err := d.MarshalYAML()
			require.NoError(t, err)
			require.Equal(t, tt.YAML, got.(string))
		})
	}
}

func TestSize_UnmarshalYAML(t *testing.T) {
	for _, tt := range testSizeData {
		result := tt.YAML
		t.Run(fmt.Sprintf("unmarshal %s", tt.YAML), func(t *testing.T) {
			var d Size
			err := d.UnmarshalYAML(func(i any) error {
				r, ok := i.(*string)
				if !ok {
					panic("cannot cast to string pointer")
				}
				*r = result
				return nil
			})
			require.NoError(t, err)
			require.Equal(t, tt.raw, d.Size)
		})
	}
}

func TestSize_MarshalJSON(t *testing.T) {
	for _, tt := range testSizeData {
		t.Run("marshal test", func(t *testing.T) {
			d := Size{
				Size: tt.raw,
			}
			got, err := d.MarshalJSON()
			require.NoError(t, err)
			require.Equal(t, tt.JSON, string(got))
		})
	}
}

func TestSize_UnmarshalJSON(t *testing.T) {
	for _, tt := range testSizeData {
		t.Run("unmarshal test", func(t *testing.T) {
			var d Size
			err := d.UnmarshalJSON([]byte(tt.JSON))
			require.NoError(t, err)
			require.Equal(t, tt.raw, d.Size)
		})
	}
}
