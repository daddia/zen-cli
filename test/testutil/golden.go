package testutil

import (
	"flag"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

var updateGolden = flag.Bool("update-golden", false, "update golden files")

// Golden compares actual output with golden file
// Usage: Golden(t, actual, "testdata/expected.golden")
func Golden(t *testing.T, actual, goldenPath string) {
	t.Helper()

	if *updateGolden {
		UpdateGolden(t, actual, goldenPath)
		return
	}

	expected := ReadGolden(t, goldenPath)
	require.Equal(t, expected, actual, "output doesn't match golden file: %s", goldenPath)
}

// UpdateGolden updates a golden file with new content
func UpdateGolden(t *testing.T, content, goldenPath string) {
	t.Helper()

	dir := filepath.Dir(goldenPath)
	err := os.MkdirAll(dir, 0755)
	require.NoError(t, err, "failed to create golden file directory")

	err = os.WriteFile(goldenPath, []byte(content), 0644)
	require.NoError(t, err, "failed to write golden file")

	t.Logf("Updated golden file: %s", goldenPath)
}

// ReadGolden reads content from a golden file
func ReadGolden(t *testing.T, goldenPath string) string {
	t.Helper()

	content, err := os.ReadFile(goldenPath)
	require.NoError(t, err, "failed to read golden file: %s", goldenPath)

	return string(content)
}

// GoldenExists checks if a golden file exists
func GoldenExists(goldenPath string) bool {
	_, err := os.Stat(goldenPath)
	return err == nil
}
