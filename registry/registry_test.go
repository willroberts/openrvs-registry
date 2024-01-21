//go:build integration

package registry

import (
	"testing"
	"time"

	beacon "github.com/willroberts/openrvs-beacon"
)

func TestAddServer(t *testing.T) {
	var (
		ip         = "184.73.85.28" // openrvs.org
		beaconPort = 7776
	)

	data, err := beacon.GetServerReport(ip, beaconPort, 5*time.Second)
	if err != nil {
		t.Log(err)
		t.FailNow()
	}

	reg := NewRegistry(Config{})
	if err := reg.AddServer(ip, data); err != nil {
		t.Log(err)
		t.FailNow()
	}

	if reg.ServerCount() != 1 {
		t.Logf("incorrect server count; expected %d, got %d", 1, reg.ServerCount())
		t.FailNow()
	}
}
