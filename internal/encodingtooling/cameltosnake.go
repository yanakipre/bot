package encodingtooling

import (
	"strings"
	"unicode"
)

// CamelToSnake is used for sqlx.MapperFunc to map UpperCamelCase struct fields to snake_case
// database columns
// no need to use `db:` struct tags
func CamelToSnake(s string) string {
	var b strings.Builder
	r := strings.NewReader(s)
	const (
		_ = iota
		lower
		upper
	)
	var prev int
	for {
		cur, _, err := r.ReadRune()
		if err != nil {
			break
		}
		if unicode.IsUpper(cur) {
			switch prev {
			case lower:
				// on rising edge: xY -> x_y
				b.WriteRune('_')
			case upper:
				// pick next rune and check
				var next rune
				if next, _, err = r.ReadRune(); err == nil {
					err = r.UnreadRune()
					if err != nil {
						return err.Error()
					}
					if unicode.IsLower(next) {
						// before falling edge: XYz -> x_yz
						b.WriteRune('_')
					}
				}
			}
			prev = upper
			cur = unicode.ToLower(cur)
		} else {
			prev = lower
		}
		b.WriteRune(cur)
	}
	return b.String()
}
