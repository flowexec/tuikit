package io

import (
	"os"
)

type StdInReader struct{}

func (r StdInReader) Read(p []byte) (n int, err error) {
	if len(p) == 0 {
		return 0, nil
	}

	info, err := os.Stdin.Stat()
	if err != nil {
		return len(p), err
	}
	switch {
	case info.Size() == 0:
		return len(p), nil
	case info.Mode()&os.ModeNamedPipe == 0:
		return len(p), nil
	default:
		return os.Stdin.Read(p)
	}
}
