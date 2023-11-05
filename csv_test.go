package registry

import (
	"strings"
	"testing"
)

func TestCSVSerializer_New(t *testing.T) {
	_ = NewCSVSerializer(false)
}

func TestCSVSerializer_Serialize(t *testing.T) {
	csv := &csvSerializer{
		headerLine: "name,ip,port,mode",
		debugMode:  false,
	}
	b := csv.Serialize(ServerMap{
		NewHostport("127.0.0.1", 6777): Server{
			Name:     "MyServer",
			IP:       "127.0.0.1",
			Port:     6777,
			GameMode: "MyGameMode",
		},
	})

	lines := strings.Split(string(b), "\n")
	if len(lines) != 2 {
		t.Log("unexpected number of lines in serialized csv output")
		t.Logf("expected %d, got %d", 2, len(lines))
		t.FailNow()
	}

	if lines[0] != csv.headerLine {
		t.Log("unexpected header line")
		t.Logf("expected %s, got %s", csv.headerLine, lines[0])
		t.FailNow()
	}

	if lines[1] != "MyServer,127.0.0.1,6777,MyGameMode" {
		t.Log("unexpected server line")
		t.Logf("expected %s, got %s", "MyServer,127.0.0.1,6777,MyGameMode", lines[1])
		t.FailNow()
	}
}

func TestCSVSerializer_SerializeDebug(t *testing.T) {
	//csv := NewCSVSerializer(true)
}

func TestCSVSerializer_Deserialize(t *testing.T) {
	//csv := NewCSVSerializer(false)
}
