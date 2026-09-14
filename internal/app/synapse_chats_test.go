package app

import (
	"fmt"
	"testing"
)

// CHATS comes a frame's worth at a time: a full page goes on from its oldest
// conversation, and a short one is the last.
func TestChatsPageOnFromTheOldest(t *testing.T) {
	page := map[string]synapseChat{}
	for i := range 16 {
		page[fmt.Sprintf("p%d", i)] = synapseChat{last: uint64(100 - i)}
	}
	if got := chatsBefore(16, page); got != "85" {
		t.Fatalf("the next page starts before %q, want 85", got)
	}
	if got := chatsBefore(15, page); got != "" {
		t.Fatalf("a short page is followed by one before %q", got)
	}
}
