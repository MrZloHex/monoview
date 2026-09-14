package app

import (
	"testing"
	"time"

	"github.com/MrZloHex/monolink/marshal"
)

// Nobody signed in, nothing goes but signing in. Signed in, it sends only
// what the person's grants cover — except PING, and MARSHAL, which judges
// its own requests.
func TestPermittedFollowsTheGrants(t *testing.T) {
	m := NewModel()
	if m.permitted("GOVERNOR", "NEW", "EVENT") || m.permitted("UKAZ", "PING", "PING") {
		t.Fatal("something was permitted with nobody signed in")
	}
	if !m.permitted("MARSHAL", "AUTH", "KEY") {
		t.Fatal("signing in was refused")
	}
	m.Logs = nil

	m.adopt(marshal.Session{Token: "t", User: "dasha", Expires: time.Now().Add(time.Hour)},
		[]string{"VERTEX.*", "GOVERNOR.GET.*"})
	for _, c := range []struct {
		to, verb, noun string
		want           bool
	}{
		{"VERTEX", "SET", "LED", true},
		{"GOVERNOR", "GET", "EVENTS", true},
		{"GOVERNOR", "NEW", "EVENT", false},
		{"UKAZ", "PRINT", "AGENDA", false},
		{"UKAZ", "PING", "PING", true},
		{"MARSHAL", "GET", "USERS", true},
	} {
		if got := m.permitted(c.to, c.verb, c.noun); got != c.want {
			t.Errorf("%s.%s.%s permitted = %v, want %v", c.to, c.verb, c.noun, got, c.want)
		}
	}
	if len(m.Logs) != 2 {
		t.Fatalf("refusals logged %d times, want once each: %+v", len(m.Logs), m.Logs)
	}
	m.permitted("UKAZ", "PRINT", "AGENDA")
	if len(m.Logs) != 2 {
		t.Fatal("a repeated refusal was logged again")
	}
}
