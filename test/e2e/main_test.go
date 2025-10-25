//go:build e2e

package e2e

import (
	"os"
	"testing"
)

// TestMain sets up and tears down the test environment for all e2e tests
func TestMain(m *testing.M) {
	// Note: Individual test setup/teardown is handled per test
	// This TestMain is kept minimal to allow parallel test execution

	// Run all tests
	code := m.Run()

	// Exit with the test result code
	os.Exit(code)
}
