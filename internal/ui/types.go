package ui

import "time"

// Sheet represents a tab/page in the UI
type Sheet int

const (
	SheetCalendar Sheet = iota
	SheetDiary
	SheetHome
	SheetSystem
)

var SheetNames = []string{
	"[1] CALENDAR",
	"[2] DIARY",
	"[3] HOME",
	"[4] SYSTEM",
}

// Event represents a calendar event
type Event struct {
	Date     time.Time
	Title    string
	Category string
}

// DiaryEntry represents a diary entry
type DiaryEntry struct {
	Date    time.Time
	Content string
	Mood    string
}

// HomeDevice represents a controllable device reachable through the concentrator.
//
// Kind determines interaction:
//
//	"toggle" – Enter flips between two states (on/off)
//	"cycle"  – Enter advances through a loop of states (off→blink→fade→solid→off)
//	"value"  – Left/Right adjusts a numeric value, Enter sends it
type HomeDevice struct {
	Name    string
	Node    string            // target node (VERTEX, LUCH, ACHTUNG)
	Verb    string            // protocol verb (LAMP, LED, ...)
	Status  string            // current state key (e.g. "on", "off", "blink")
	Kind    string            // "toggle", "cycle", "value"
	Actions map[string]string // status -> noun to send (toggle & cycle)
	Noun    string            // fixed noun for value kind (e.g. "BRIGHT")
	Val     int               // current numeric value (value kind)
	Min     int
	Max     int
	Step    int
	Pending bool              // true while waiting for OK after sending a command
}

// SystemNode represents a system/server node
type SystemNode struct {
	Name   string
	Status string
	CPU    float64
	Memory float64
	Uptime string
}

// LogEntry represents a log line
type LogEntry struct {
	Time    time.Time
	Level   string
	Source  string
	Message string
}

// ScheduleEntry represents a university class/lecture
type ScheduleEntry struct {
	Weekday  time.Weekday // 0=Sunday, 1=Monday, etc.
	Start    string       // "10:45"
	End      string       // "12:10"
	Title    string
	Location string
	Tags     []string // e.g. ["Lecture", "Math"]
}

// Tag colors for schedule (Gruvbox-based)
var TagColors = map[string]string{
	// Type tags
	"Lecture": "#b8bb26", // green
	"Seminar": "#8ec07c", // aqua
	"Lab":     "#d3869b", // purple
	// Subject tags
	"Math":    "#fb4934", // red
	"DM":      "#fe8019", // orange
	"ATP":     "#fabd2f", // yellow
	"FL":      "#83a598", // blue
	"Practic": "#d3869b", // purple
}
