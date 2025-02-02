package projectpath

import (
	"path"
	"path/filepath"
	"runtime"
)

var (
	_, b, _, _ = runtime.Caller(0)
	RootPath   = path.Clean(path.Join(filepath.Dir(b), "../.."))
)
