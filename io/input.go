package io

import (
	"os"
)

type StdInReader struct {
	StdIn *os.File
}

func (r StdInReader) Read(p []byte) (n int, err error) {
	return r.StdIn.Read(p)
}
