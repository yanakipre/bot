package jsontooling

import (
	"bytes"
	"encoding/json"
)

// StrictUnmarshal complains if JSON contains fields, not present in structure.
func StrictUnmarshal(data []byte, v any) error {
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.DisallowUnknownFields()
	return dec.Decode(v)
}
