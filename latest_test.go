//go:build integration

package registry

import "testing"

func TestGetLatestReleaseVersion(t *testing.T) {
	b := GetLatestReleaseVersion()
	if string(b) != "v1.5" {
		t.Log("unexpected version received")
		t.Logf("expected %s, got %s", "v1.5", string(b))
		t.FailNow()
	}
}
