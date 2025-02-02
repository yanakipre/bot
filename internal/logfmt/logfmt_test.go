package main

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLogfmt_ToSnakeCase(t *testing.T) {
	tt := []struct {
		in, out string
	}{
		{"", ""},
		{"snake_case_input_is_untouched", "snake_case_input_is_untouched"},
		{`___:/\trailing_and_leading_symbols_are_trimmed_://\___`, "trailing_and_leading_symbols_are_trimmed"},
		{"repeating___separators_are_reduced", "repeating_separators_are_reduced"},
		{"CamelCaseIsConverted", "camel_case_is_converted"},
		{"camelCaseIsConverted", "camel_case_is_converted"},
		{"AbbreviationsLikeURLorHTTPareNotSeparated", "abbreviations_like_url_or_http_are_not_separated"},
		{"all.inner \t Separators&//Are:\nnormalized", "all_inner_separators_are_normalized"},
	}

	for _, tc := range tt {
		t.Run(tc.in, func(t *testing.T) {
			require.Equal(t, tc.out, toSnakeCase(tc.in))
		})
	}
}
