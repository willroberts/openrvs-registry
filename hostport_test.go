package registry

import "testing"

func TestNewHostport(t *testing.T) {
	h := NewHostport("127.0.0.1", 6777)
	expected := Hostport("127.0.0.1:6777")
	if h != expected {
		t.Log("failed to generate hostport")
		t.Logf("expected %s, got %s", expected, h)
		t.FailNow()
	}
}
