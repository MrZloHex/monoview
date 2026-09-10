package app

import (
	"testing"
	"time"

	"github.com/MrZloHex/monolink"
)

// Every hub message used to start a tick chain of its own. Chains never end, so after
// hours of bus traffic thousands of them were re-rendering the UI every second.
func TestHubTrafficDoesNotMultiplyTickChains(t *testing.T) {
	pong, err := monolink.Parse("MONOVIEW:PONG:PING:VERTEX")
	if err != nil {
		t.Fatal(err)
	}

	m := NewModel()
	for i := 0; i < 100; i++ {
		next, cmd := m.Update(HubMsg(pong))
		m = next.(Model)
		if cmd != nil {
			t.Fatalf("hub message %d started a tick chain", i)
		}
	}

	next, cmd := m.Update(TickMsg{Time: time.Now(), Gen: m.tickGen})
	m = next.(Model)
	if cmd == nil {
		t.Fatal("live tick did not continue its chain")
	}

	if _, cmd := m.Update(TickMsg{Time: time.Now(), Gen: m.tickGen - 1}); cmd != nil {
		t.Fatal("superseded tick chain kept running")
	}
}
