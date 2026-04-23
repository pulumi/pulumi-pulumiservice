// Copyright 2016-2026, Pulumi Corporation.

package gen

import "os"

// readFile is a thin wrapper so tests can fake it without stubbing the whole
// package. Kept unexported on purpose.
func readFile(path string) ([]byte, error) {
	return os.ReadFile(path)
}
