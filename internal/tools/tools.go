//go:build tools
// +build tools

package tools

import (
	// required for custom golangci-lint plugin to work correctly
	_ "github.com/yanakipe/bot/pkg/lint/errlint"
)
