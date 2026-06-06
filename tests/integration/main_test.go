// Package integration holds integration tests that need a full environment
// (miniredis + database mock + HTTP router) wired together.
//
// TestMain in this package initializes all shared dependencies once,
// so individual test functions don't need to repeat setup.
//
// Run with: go test -v ./tests/integration/
package integration

import (
	"fmt"
	"os"
	"testing"
)

// TestMain is the integration test entry point.
// Add shared setup here as integration tests are written.
func TestMain(m *testing.M) {
	// TODO: Initialize shared integration test dependencies:
	//   - miniredis (in-memory Redis)
	//   - sqlmock (in-memory DB)
	//   - HTTP test server
	//   - etc.

	fmt.Fprintln(os.Stderr, "integration: no tests yet")

	code := m.Run()
	os.Exit(code)
}
