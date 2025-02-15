// nolint
//
// NOTE: this is copied from lib/pq.
//
// If we switch from database/sql to pgx native interface,
// we will not need StringArray or Float32Array anymore and we'll be able to just use []string and []float32.
package rdb

import (
	"bytes"
	"database/sql/driver"
	"fmt"
	"strconv"
	"strings"
)

// ArrayDelimiter may be optionally implemented by driver.Valuer or sql.Scanner
// to override the array delimiter used by GenericArray.
type ArrayDelimiter interface {
	// ArrayDelimiter returns the delimiter character(s) for this element's type.
	ArrayDelimiter() string
}

// Float32Array represents a one-dimensional array of the PostgreSQL double
// precision type.
//
// NOTE: this is copied from lib/pq.
//
// If we switch from database/sql to pgx native interface,
// we will not need StringArray or Float32Array anymore and we'll be able to just use []string and []float32.
type Float32Array []float32

// Scan implements the sql.Scanner interface.
func (a *Float32Array) Scan(src any) error {
	switch src := src.(type) {
	case []byte:
		return a.scanBytes(src)
	case string:
		return a.scanBytes([]byte(src))
	case nil:
		*a = nil
		return nil
	}

	return fmt.Errorf("rdb: cannot convert %T to Float32Array", src)
}

func (a *Float32Array) scanBytes(src []byte) error {
	elems, err := scanLinearArray(src, []byte{','}, "Float32Array")
	if err != nil {
		return err
	}
	if *a != nil && len(elems) == 0 {
		*a = (*a)[:0]
	} else {
		b := make(Float32Array, len(elems))
		for i, v := range elems {
			var x float64
			if x, err = strconv.ParseFloat(string(v), 32); err != nil {
				return fmt.Errorf("rdb: parsing array element index %d: %v", i, err)
			}
			b[i] = float32(x)
		}
		*a = b
	}
	return nil
}

// Value implements the driver.Valuer interface.
func (a Float32Array) Value() (driver.Value, error) {
	if a == nil {
		return nil, nil
	}

	if n := len(a); n > 0 {
		// There will be at least two curly brackets, N bytes of values,
		// and N-1 bytes of delimiters.
		b := make([]byte, 1, 1+2*n)
		b[0] = '{'

		b = strconv.AppendFloat(b, float64(a[0]), 'f', -1, 32)
		for i := 1; i < n; i++ {
			b = append(b, ',')
			b = strconv.AppendFloat(b, float64(a[i]), 'f', -1, 32)
		}

		return string(append(b, '}')), nil
	}

	return "{}", nil
}

// StringArray represents a one-dimensional array of the PostgreSQL character types.
//
// NOTE: this is copied from lib/pq.
//
// If we switch from database/sql to pgx native interface,
// we will not need StringArray or Float32Array anymore and we'll be able to just use []string and []float32.
type StringArray []string

// Scan implements the sql.Scanner interface.
func (a *StringArray) Scan(src any) error {
	switch src := src.(type) {
	case []byte:
		return a.scanBytes(src)
	case string:
		return a.scanBytes([]byte(src))
	case nil:
		*a = nil
		return nil
	}

	return fmt.Errorf("rdb: cannot convert %T to StringArray", src)
}

func (a *StringArray) scanBytes(src []byte) error {
	elems, err := scanLinearArray(src, []byte{','}, "StringArray")
	if err != nil {
		return err
	}
	if *a != nil && len(elems) == 0 {
		*a = (*a)[:0]
	} else {
		b := make(StringArray, len(elems))
		for i, v := range elems {
			if b[i] = string(v); v == nil {
				return fmt.Errorf("rdb: parsing array element index %d: cannot convert nil to string", i)
			}
		}
		*a = b
	}
	return nil
}

// Value implements the driver.Valuer interface.
func (a StringArray) Value() (driver.Value, error) {
	if a == nil {
		return nil, nil
	}

	if n := len(a); n > 0 {
		// There will be at least two curly brackets, 2*N bytes of quotes,
		// and N-1 bytes of delimiters.
		b := make([]byte, 1, 1+3*n)
		b[0] = '{'

		b = appendArrayQuotedBytes(b, []byte(a[0]))
		for i := 1; i < n; i++ {
			b = append(b, ',')
			b = appendArrayQuotedBytes(b, []byte(a[i]))
		}

		return string(append(b, '}')), nil
	}

	return "{}", nil
}

func appendArrayQuotedBytes(b, v []byte) []byte {
	b = append(b, '"')
	for {
		i := bytes.IndexAny(v, `"\`)
		if i < 0 {
			b = append(b, v...)
			break
		}
		if i > 0 {
			b = append(b, v[:i]...)
		}
		b = append(b, '\\', v[i])
		v = v[i+1:]
	}
	return append(b, '"')
}

// parseArray extracts the dimensions and elements of an array represented in
// text format. Only representations emitted by the backend are supported.
// Notably, whitespace around brackets and delimiters is significant, and NULL
// is case-sensitive.
//
// See http://www.postgresql.org/docs/current/static/arrays.html#ARRAYS-IO
func parseArray(src, del []byte) (dims []int, elems [][]byte, err error) {
	var depth, i int

	if len(src) < 1 || src[0] != '{' {
		return nil, nil, fmt.Errorf("rdb: unable to parse array; expected %q at offset %d", '{', 0)
	}

Open:
	for i < len(src) {
		switch src[i] {
		case '{':
			depth++
			i++
		case '}':
			elems = make([][]byte, 0)
			goto Close
		default:
			break Open
		}
	}
	dims = make([]int, i)

Element:
	for i < len(src) {
		switch src[i] {
		case '{':
			if depth == len(dims) {
				break Element
			}
			depth++
			dims[depth-1] = 0
			i++
		case '"':
			elem := []byte{}
			var escape bool
			for i++; i < len(src); i++ {
				if escape {
					elem = append(elem, src[i])
					escape = false
				} else {
					switch src[i] {
					default:
						elem = append(elem, src[i])
					case '\\':
						escape = true
					case '"':
						elems = append(elems, elem)
						i++
						break Element
					}
				}
			}
		default:
			for start := i; i < len(src); i++ {
				if bytes.HasPrefix(src[i:], del) || src[i] == '}' {
					elem := src[start:i]
					if len(elem) == 0 {
						return nil, nil, fmt.Errorf("rdb: unable to parse array; unexpected %q at offset %d", src[i], i)
					}
					if bytes.Equal(elem, []byte("NULL")) {
						elem = nil
					}
					elems = append(elems, elem)
					break Element
				}
			}
		}
	}

	for i < len(src) {
		if bytes.HasPrefix(src[i:], del) && depth > 0 {
			dims[depth-1]++
			i += len(del)
			goto Element
		} else if src[i] == '}' && depth > 0 {
			dims[depth-1]++
			depth--
			i++
		} else {
			return nil, nil, fmt.Errorf("rdb: unable to parse array; unexpected %q at offset %d", src[i], i)
		}
	}

Close:
	for i < len(src) {
		if src[i] == '}' && depth > 0 {
			depth--
			i++
		} else {
			return nil, nil, fmt.Errorf("rdb: unable to parse array; unexpected %q at offset %d", src[i], i)
		}
	}
	if depth > 0 {
		err = fmt.Errorf("rdb: unable to parse array; expected %q at offset %d", '}', i)
	}
	if err == nil {
		for _, d := range dims {
			if (len(elems) % d) != 0 {
				err = fmt.Errorf("rdb: multidimensional arrays must have elements with matching dimensions")
			}
		}
	}
	return
}

func scanLinearArray(src, del []byte, typ string) (elems [][]byte, err error) {
	dims, elems, err := parseArray(src, del)
	if err != nil {
		return nil, err
	}
	if len(dims) > 1 {
		return nil, fmt.Errorf(
			"rdb: cannot convert ARRAY%s to %s",
			strings.Replace(fmt.Sprint(dims), " ", "][", -1),
			typ,
		)
	}
	return elems, err
}
