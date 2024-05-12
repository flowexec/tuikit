package io

import (
	"os"
)

type StdInReader struct {
	StdIn *os.File
}

func (r StdInReader) Read(p []byte) (n int, err error) {
	if len(p) == 0 {
		return 0, nil
	}

	if r.StdIn == nil {
		r.StdIn = os.Stdin
	}
	info, err := r.StdIn.Stat()
	if err != nil {
		return len(p), err
	}
	switch {
	case info.Size() != 0 && info.Mode()&os.ModeNamedPipe == 0:
		return len(p), nil
	default:
		return r.StdIn.Read(p)
	}
}
