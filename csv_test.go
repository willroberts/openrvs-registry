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

	expected := "MyServer,127.0.0.1,6777,MyGameMode"
	if lines[1] != expected {
		t.Log("unexpected server line")
		t.Logf("expected %s, got %s", expected, lines[1])
		t.FailNow()
	}
}

func TestCSVSerializer_SerializeDebug(t *testing.T) {
	csv := NewCSVSerializer(true)
	b := csv.Serialize(ServerMap{
		NewHostport("127.0.0.1", 6777): Server{
			Name:     "MyServer",
			IP:       "127.0.0.1",
			Port:     6777,
			GameMode: "MyGameMode",
		},
	})

	s := strings.Split(string(b), "\n")[1]
	expected := "MyServer,127.0.0.1,6777,MyGameMode,healthy=false,expired=false,passed=0,failed=0"
	if s != expected {
		t.Log("unexpected server line")
		t.Logf("expected %s, got %s", expected, s)
		t.FailNow()
	}

}

func TestCSVSerializer_Deserialize(t *testing.T) {
	//csv := NewCSVSerializer(false)
}
