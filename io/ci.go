package io

import "os"

// isCI returns true if running in a CI environment.
func isCI() bool {
	return os.Getenv("GITHUB_ACTIONS") == "true" || os.Getenv("CI") == "true"
}
