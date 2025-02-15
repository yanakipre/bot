package logger

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_RawJSON(t *testing.T) {
	field := RawJSON("test", rawStruct{Hello: "world"})
	require.Equal(t, "test", field.Key)
	// makes sure that there is no trailing newline
	require.Equal(t, []byte(`{"Hello":"world"}`), field.Interface)
}

type rawStruct struct {
	Hello string
}
