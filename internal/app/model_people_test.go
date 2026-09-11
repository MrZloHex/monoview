package app

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/MrZloHex/monolink/marshal"
)

// Nobody signed in is the panel as it always was. Signed in, it sends only
// what the person's grants cover — except PING, and MARSHAL, which judges
// its own requests.
func TestPermittedFollowsTheGrants(t *testing.T) {
	m := NewModel()
	if !m.permitted("GOVERNOR", "NEW", "EVENT") {
		t.Fatal("refused with nobody signed in")
	}

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

// The session outlives the panel's process, readable by its owner only.
func TestSessionIsKeptBetweenRuns(t *testing.T) {
	path := filepath.Join(t.TempDir(), "session.json")
	m := NewModel()
	m.SessionPath = path
	m.adopt(marshal.Session{Token: "tok", User: "mzh", Expires: time.Now().Add(time.Hour).Truncate(time.Second)}, []string{"*"})
	m.saveSession()

	if fi, err := os.Stat(path); err != nil || fi.Mode().Perm() != 0o600 {
		t.Fatalf("session file: %v, %v", fi, err)
	}

	next := NewModel()
	next.SessionPath = path
	next.RestoreSession()
	if next.Session.Token != "tok" || next.Session.User != "mzh" || len(next.Grants) != 1 {
		t.Fatalf("restored %+v, grants %q", next.Session, next.Grants)
	}

	next.dropSession()
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Fatal("signing out left the session file behind")
	}
}

func TestExpiredSessionIsNotRestored(t *testing.T) {
	path := filepath.Join(t.TempDir(), "session.json")
	m := NewModel()
	m.SessionPath = path
	m.adopt(marshal.Session{Token: "tok", User: "mzh", Expires: time.Now().Add(-time.Minute)}, nil)
	m.saveSession()

	next := NewModel()
	next.SessionPath = path
	next.RestoreSession()
	if next.signedIn() {
		t.Fatal("restored an expired session")
	}
}
