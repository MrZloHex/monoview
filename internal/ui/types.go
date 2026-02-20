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
// Protocol reference (VERTEX):
//
//	TOGGLE:LAMP             -> OK:LAMP
//	ON:LED / OFF:LED        -> OK:LED
//	SET:LED:MODE:BLINK      -> OK:LED
//	SET:LED:BRIGHT:128      -> OK:LED
//	GET:LAMP:STATE          -> OK:LAMP:STATE:ON
//
// Kind determines interaction:
//
//	"toggle" – Enter sends TOGGLE:<Topic>
//	"cycle"  – Enter sends SET:<Topic>:MODE:<next> or OFF:<Topic>
//	"value"  – Left/Right adjusts, sends SET:<Topic>:BRIGHT:<val>
type HomeDevice struct {
	Name    string
	Node    string   // target node (VERTEX, LUCH, ACHTUNG)
	Topic   string   // protocol topic/noun (LAMP, LED, BUZZ)
	Kind    string   // "toggle", "cycle", "value"
	Status  string   // current state: "on"/"off"/"unknown" for toggle; mode for cycle
	Pending bool

	// cycle
	Modes []string // ordered modes, e.g. ["off","blink","fade","solid"]

	// value
	Property string // SET sub-property (e.g. "BRIGHT")
	Val      int
	Min      int
	Max      int
	Step     int
}

// SystemNode represents a real system node reachable through the concentrator.
type SystemNode struct {
	Name      string
	PingNoun  string    // noun for PING command ("PINT" for VERTEX, "PING" for ACHTUNG)
	Status    string    // "online", "offline", "unknown"
	Uptime    string    // human-readable uptime from GET:UPTIME
	LastSeen  time.Time // last time we got a PONG
	PingMs    int64     // last round-trip in ms
	PingSent  time.Time // when we last sent PING (for RTT calc)
}

// FireAlert is shown when ACHTUNG broadcasts ALL:FIRE:TIMER/ALARM:name.
type FireAlert struct {
	Show    bool
	JobKind string // "TIMER" or "ALARM"
	JobName string
}

// AchtungJob is a timer or alarm on the ACHTUNG node.
// Remaining is updated every tick from EndTime when set; Due is shown as-is.
type AchtungJob struct {
	Kind      string     // "TIMER" or "ALARM"
	Name      string
	Remaining string     // human-readable countdown, updated from EndTime
	Due       string     // human-readable due date/time from server
	EndTime   *time.Time // when set, Remaining is computed each tick until this time
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
