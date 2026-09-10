package app

import (
	"testing"
	"time"
)

func TestParseAchtungEndTimeAcceptsCanonicalDue(t *testing.T) {
	// What achtung actually sends today.
	for _, kind := range []string{"ALARM", "EVERY", "DAILY"} {
		got := parseAchtungEndTime(kind, "6h22m53s", "2026.09.10.07.05")
		if got == nil {
			t.Fatalf("%s: canonical due token did not parse", kind)
		}
		if got.Year() != 2026 || got.Month() != time.September || got.Day() != 10 ||
			got.Hour() != 7 || got.Minute() != 5 {
			t.Fatalf("%s: parsed to %v", kind, got)
		}
	}
	// Unpadded, as older achtung builds emit.
	if got := parseAchtungEndTime("ALARM", "1h", "2026.9.10.7.5"); got == nil {
		t.Fatal("unpadded due token did not parse")
	}
	// TIMER still uses the remaining duration, not the due.
	if got := parseAchtungEndTime("TIMER", "10s", ""); got == nil {
		t.Fatal("TIMER should derive its end time from remaining")
	}
	// Garbage stays nil rather than a wrong time.
	if got := parseAchtungEndTime("ALARM", "1h", "banana"); got != nil {
		t.Fatalf("garbage due parsed to %v", got)
	}
}
