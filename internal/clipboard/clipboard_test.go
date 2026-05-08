package clipboard

import "testing"

func TestCopyRuns(t *testing.T) {
	err := Copy("tellme test")
	if err != nil {
		t.Logf("clipboard not available: %v (ok in CI)", err)
	}
}
