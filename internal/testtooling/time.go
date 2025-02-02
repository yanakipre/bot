package testtooling

import "time"

// GetLocalTimeFromStr
// Timezones are tricky here, one can time in UTC here.
//
// http://golang.org/pkg/time/#Parse:
// When parsing a time with a zone abbreviation like MST, if the zone abbreviation
// has a defined offset in the current location, then that offset is used.
// If the zone abbreviation is unknown, Parse records the time as being in a fabricated
// location with the given zone abbreviation and a zero offset.
// See https://github.com/golang/go/issues/9617 for more.
//
// Deprecated.
func GetLocalTimeFromStr(timeStr string) time.Time {
	r, err := time.Parse(time.RFC822, timeStr)
	if err != nil {
		panic(err)
	}
	// the DB driver WILL return time in Local timezone.
	// and we need to compare time in same timezones
	// see https://github.com/stretchr/testify/issues/666
	return r.Local()
}

func GetLocalTimeFromStrWithSeconds(timeStr string) time.Time {
	r, err := time.Parse(time.RFC1123, timeStr)
	if err != nil {
		panic(err)
	}
	// the DB driver WILL return time in Local timezone.
	// and we need to compare time in same timezones
	// see https://github.com/stretchr/testify/issues/666
	return r.Local()
}

func GetLocalTimeFromRFC3339Str(timeStr string) time.Time {
	r, err := time.Parse(time.RFC3339, timeStr)
	if err != nil {
		panic(err)
	}
	// the DB driver WILL return time in Local timezone.
	// and we need to compare time in same timezones
	// see https://github.com/stretchr/testify/issues/666
	return r.Local()
}

func GetLocalTimeFromRFC3339StrPtr(timeStr string) *time.Time {
	r, err := time.Parse(time.RFC3339, timeStr)
	if err != nil {
		panic(err)
	}
	// the DB driver WILL return time in Local timezone.
	// and we need to compare time in same timezones
	// see https://github.com/stretchr/testify/issues/666
	r = r.Local()
	return &r
}
