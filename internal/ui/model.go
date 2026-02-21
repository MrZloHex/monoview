package ui

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"monoview/pkg/concentrator"
)

const (
	pingInterval     = 2 * time.Minute
	achtungSyncEvery = 1 * time.Minute
)


// HubMsg wraps a concentrator message arriving through the inbox channel.
type HubMsg concentrator.Message

// Model is the main application model
type Model struct {
	ActiveSheet Sheet
	Width       int
	Height      int
	LastUpdate  time.Time

	// Concentrator client (runs in background goroutine)
	Hub *concentrator.Client

	// Calendar
	SelectedDate        time.Time
	SelectedEvent       int  // index into events for SelectedDate (when CalendarFocusEvents)
	CalendarFocusEvents bool // false = ↑/↓ move day (by week), Enter = focus events; true = ↑/↓ select event
	Events              []Event
	Deadlines     []Event // from GET:DEADLINES (upcoming deadlines box)
	Schedule      []ScheduleEntry

	// Diary
	DiaryEntries  []DiaryEntry
	SelectedEntry int

	// Home
	HomeDevices    []HomeDevice
	SelectedDevice int

	// System
	Nodes        []SystemNode
	Logs         []LogEntry
	SelectedNode int

	// ACHTUNG (timers & alarms, shown on Home sheet)
	AchtungJobs        []AchtungJob
	SelectedAchtungJob int
	AchtungTimerMenu     bool   // true = adding timer (all fields in right panel)
	AchtungTimerDuration string // e.g. "5m"
	AchtungTimerName     string // optional, Enter for auto
	AchtungTimerFocusField int  // 0=duration, 1=name
	AchtungAlarmMenu     bool   // true = adding alarm (all fields in right panel)
	AchtungAlarmDate     string // YYYY-MM-DD
	AchtungAlarmTime     string // HH:MM
	AchtungAlarmName     string // optional
	AchtungAlarmFocusField int  // 0=date, 1=time, 2=name
	HomeFocusAchtung    bool   // on Home: true = focus timers panel (j/k, enter, t, a, d)
	LastAchtungSync     time.Time

	// Fire alert popup (ALL:FIRE:TIMER/ALARM from ACHTUNG)
	FireAlert FireAlert

	// Calendar: viewing selected event details in right panel (Enter on event)
	EventViewMenu bool

	// Add event flow (Calendar sheet): popup with all fields; EventAddFocusField = which field gets input
	EventAddMenu        bool   // true = add-event form active
	EventAddFocusField  int    // 0=title, 1=date, 2=time, 3=location, 4=notes, 5=visible_from
	EventAddTitle       string
	EventAddDate        string // YYYY-MM-DD
	EventAddTime        string // HH:MM or HH:MM:SS
	EventAddLocation    string
	EventAddNotes       string
	EventAddVisibleFrom string // optional YYYY-MM-DD; omit = default (7 days before deadline)

	// Traffic indicators (timestamps of last rx/tx for arrow display)
	LastRx time.Time
	LastTx time.Time
}

// TickMsg is sent every second
type TickMsg time.Time

func tick() tea.Cmd {
	return tea.Tick(time.Second, func(t time.Time) tea.Msg {
		return TickMsg(t)
	})
}

// NewModel creates the initial model with sample data
func NewModel() Model {
	now := time.Now()
	return Model{
		ActiveSheet:  SheetCalendar,
		LastUpdate:   now,
		SelectedDate: now,

		// Events and Schedule are filled from GOVERNOR (GET:EVENTS, GET:SCHEDULE:<weekday>)
		Events:   nil,
		Schedule: nil,

		DiaryEntries: []DiaryEntry{
			{Date: now, Content: "Started working on MonoView TUI...", Mood: "focused"},
			{Date: now.Add(-24 * time.Hour), Content: "Fixed the WebSocket connection issues.", Mood: "productive"},
			{Date: now.Add(-48 * time.Hour), Content: "Rainy day. Read documentation.", Mood: "calm"},
		},
		SelectedEntry: 0,

		HomeDevices: []HomeDevice{
			{
				Name: "Desk Lamp", Node: "VERTEX", Topic: "LAMP",
				Kind: "toggle", Status: "unknown",
			},
			{
				Name: "LED Light", Node: "VERTEX", Topic: "LED",
				Kind: "toggle", Status: "unknown",
			},
			{
				Name: "LED Mode", Node: "VERTEX", Topic: "LED",
				Kind: "cycle", Status: "solid",
				Modes: []string{"solid", "fade", "blink"},
			},
			{
				Name: "Brightness", Node: "VERTEX", Topic: "LED",
				Kind: "value", Property: "BRIGHT",
				Val: 128, Min: 0, Max: 255, Step: 15,
			},
		},
		SelectedDevice: 0,

		Nodes: []SystemNode{
			{Name: "VERTEX", PingNoun: "PINT", Status: "offline", Uptime: "—"},
			{Name: "ACHTUNG", PingNoun: "PING", Status: "offline", Uptime: "—"},
			{Name: "GOVERNOR", PingNoun: "PING", Status: "offline", Uptime: "—"},
			{Name: "UKAZ", PingNoun: "PING", Status: "offline", Uptime: "—"},
		},
		SelectedNode: 0,
	}
}

// waitForHub blocks on the concentrator inbox and delivers the next
// message as a HubMsg into the Bubble Tea event loop.
func waitForHub(inbox <-chan concentrator.Message) tea.Cmd {
	return func() tea.Msg {
		msg, ok := <-inbox
		if !ok {
			return nil
		}
		return HubMsg(msg)
	}
}

func (m Model) Init() tea.Cmd {
	cmds := []tea.Cmd{tick()}
	if m.Hub != nil {
		if inbox := m.Hub.Inbox(); inbox != nil {
			cmds = append(cmds, waitForHub(inbox))
		}
		m.queryDeviceStates()
		m.requestGovernorSchedule()
		m.requestGovernorEvents()
		m.requestGovernorDeadlines()
	}
	return tea.Batch(cmds...)
}

// queryDeviceStates sends GET requests for every home device to sync
// with the real hardware state on startup.
//
//	VERTEX:GET:LAMP:STATE:MONOVIEW
//	VERTEX:GET:LED:STATE:MONOVIEW
//	VERTEX:GET:LED:MODE:MONOVIEW
//	VERTEX:GET:LED:BRIGHT:MONOVIEW
func (m *Model) queryDeviceStates() {
	seen := map[string]bool{}
	for _, dev := range m.HomeDevices {
		switch dev.Kind {
		case "toggle":
			key := dev.Node + ":" + dev.Topic + ":STATE"
			if !seen[key] {
				seen[key] = true
				m.HubSend(dev.Node, "GET", dev.Topic, "STATE")
			}
		case "cycle":
			key := dev.Node + ":" + dev.Topic + ":MODE"
			if !seen[key] {
				seen[key] = true
				m.HubSend(dev.Node, "GET", dev.Topic, "MODE")
			}
		case "value":
			key := dev.Node + ":" + dev.Topic + ":" + dev.Property
			if !seen[key] {
				seen[key] = true
				m.HubSend(dev.Node, "GET", dev.Topic, dev.Property)
			}
		}
	}
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if m.FireAlert.Show {
			switch msg.String() {
			case "enter", " ", "q", "esc":
				m.dismissFireAlert()
				return m, nil
			}
		}
		if m.handleAchtungFormKeys(msg) {
			return m, nil
		}
		if m.handleAchtungKeys(msg) {
			return m, nil
		}
		if m.handleEventAddKeys(msg) {
			return m, nil
		}
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit

		// Sheet navigation
		case "1":
			m.ActiveSheet = SheetCalendar
			m.CalendarFocusEvents = false
			m.EventViewMenu = false
			if m.EventAddMenu {
				m.eventAddReset()
			}
		case "2":
			m.ActiveSheet = SheetDiary
		case "3":
			m.ActiveSheet = SheetHome
			m.requestAchtungList()
		case "4":
			m.ActiveSheet = SheetSystem
		case "tab":
			if m.ActiveSheet == SheetHome {
				m.HomeFocusAchtung = !m.HomeFocusAchtung
			} else {
				m.ActiveSheet = (m.ActiveSheet + 1) % 4
				if m.ActiveSheet == SheetCalendar {
					m.CalendarFocusEvents = false
					m.EventViewMenu = false
				}
				if m.ActiveSheet == SheetHome {
					m.requestAchtungList()
				}
			}
		case "shift+tab":
			if m.ActiveSheet == SheetHome {
				m.HomeFocusAchtung = !m.HomeFocusAchtung
			} else {
				m.ActiveSheet = (m.ActiveSheet + 3) % 4
				if m.ActiveSheet == SheetCalendar {
					m.CalendarFocusEvents = false
					m.EventViewMenu = false
				}
				if m.ActiveSheet == SheetHome {
					m.requestAchtungList()
				}
			}
		case "esc":
			if m.ActiveSheet == SheetCalendar {
				if m.EventViewMenu {
					m.EventViewMenu = false
				} else if m.CalendarFocusEvents {
					m.CalendarFocusEvents = false
				}
			}

		// Calendar: [a] or [n] add new event (opens form)
		case "a", "n":
			if m.ActiveSheet == SheetCalendar && !m.EventAddMenu && m.Hub != nil {
				m.EventAddMenu = true
				m.EventViewMenu = false
				m.EventAddFocusField = 0
				m.EventAddDate = m.SelectedDate.Format("2006-01-02")
			}
		}

		// Navigation within sheets (no-op when in add-event form)
		if m.ActiveSheet != SheetCalendar || !m.EventAddMenu {
			switch msg.String() {
			case "j", "down":
				m.navigateDown()
			case "k", "up":
				m.navigateUp()
			case "h", "left":
				m.navigateLeft()
			case "l", "right":
				m.navigateRight()
			case "enter", " ":
				// Calendar: Enter on selected day switches to event selection; Enter on event shows details
				if m.ActiveSheet == SheetCalendar && !m.CalendarFocusEvents {
					m.CalendarFocusEvents = true
					dayEvents := m.eventsForSelectedDate()
					if m.SelectedEvent >= len(dayEvents) {
						m.SelectedEvent = len(dayEvents) - 1
					}
					if m.SelectedEvent < 0 {
						m.SelectedEvent = 0
					}
				} else if m.ActiveSheet == SheetCalendar && m.CalendarFocusEvents {
					dayEvents := m.eventsForSelectedDate()
					if len(dayEvents) > 0 && m.SelectedEvent >= 0 && m.SelectedEvent < len(dayEvents) {
						m.EventViewMenu = true
					} else {
						m.toggleAction()
						m.pingSelectedNode()
					}
				} else {
					m.toggleAction()
					m.pingSelectedNode()
				}
			case "d", "backspace":
				if m.ActiveSheet == SheetCalendar {
					if m.EventViewMenu {
						m.deleteSelectedEvent()
						m.EventViewMenu = false
					} else if m.CalendarFocusEvents {
						m.deleteSelectedEvent()
					}
				}
			}
		}

	case tea.WindowSizeMsg:
		m.Width = msg.Width
		m.Height = msg.Height

	case TickMsg:
		m.LastUpdate = time.Time(msg)
		m.pollNodes()
		m.updateAchtungRemaining()
		if m.Hub != nil && m.Hub.Connected() && time.Since(m.LastAchtungSync) >= achtungSyncEvery {
			m.requestAchtungList()
			m.LastAchtungSync = time.Now()
		}
		return m, tick()

	case HubMsg:
		m.handleHub(concentrator.Message(msg))
		m.updateAchtungRemaining()
		var cmd tea.Cmd
		if m.Hub != nil {
			if inbox := m.Hub.Inbox(); inbox != nil {
				cmd = waitForHub(inbox)
			}
		}
		return m, tea.Batch(tick(), cmd)
	}

	return m, tick()
}

// handleHub processes an incoming concentrator message and updates model state.
func (m *Model) handleHub(msg concentrator.Message) {
	now := time.Now()
	m.LastRx = now

	m.Logs = append([]LogEntry{{
		Time:    now,
		Level:   "MSG",
		Source:  msg.From,
		Message: msg.Raw,
	}}, m.Logs...)

	const maxLogs = 50
	if len(m.Logs) > maxLogs {
		m.Logs = m.Logs[:maxLogs]
	}

	m.handleNodeResponse(msg)
	m.handleGovernorResponse(msg)
	m.handleDeviceResponse(msg)
	m.handleAchtungResponse(msg)
	m.handleFireAlert(msg)
}

// requestGovernorSchedule requests GET:SCHEDULE:<weekday> for Mon–Sat from GOVERNOR.
func (m *Model) requestGovernorSchedule() {
	weekdays := []string{"Mon", "Tue", "Wed", "Thu", "Fri", "Sat"}
	for _, wd := range weekdays {
		m.HubSend("GOVERNOR", "GET", "SCHEDULE", wd)
	}
}

// requestGovernorEvents requests GET:EVENTS from GOVERNOR (list all events).
func (m *Model) requestGovernorEvents() {
	m.HubSend("GOVERNOR", "GET", "EVENTS")
}

// requestGovernorDeadlines requests GET:DEADLINES from GOVERNOR (fills upcoming deadlines box).
// Optional scope: day, month, year (e.g. GET:DEADLINES:month).
func (m *Model) requestGovernorDeadlines() {
	m.HubSend("GOVERNOR", "GET", "DEADLINES")
}

// handleGovernorResponse processes OK:SCHEDULE and OK:EVENTS from GOVERNOR.
func (m *Model) handleGovernorResponse(msg concentrator.Message) {
	if strings.ToUpper(msg.From) != "GOVERNOR" || strings.ToUpper(msg.Verb) != "OK" {
		return
	}
	noun := strings.ToUpper(msg.Noun)
	switch noun {
	case "SCHEDULE":
		m.handleGovernorSchedule(msg)
	case "EVENTS":
		m.handleGovernorEvents(msg)
	case "EVENT":
		m.handleGovernorEventCreated(msg)
	case "DEADLINES":
		m.handleGovernorDeadlines(msg)
	}
}

// handleGovernorSchedule processes OK:SCHEDULE. Slot format (wire): Weekday|Start|End|Title|Location|Tags (colons in values are dots).
func (m *Model) handleGovernorSchedule(msg concentrator.Message) {
	entries := parseGovernorScheduleSlots(msg.Args)
	if len(entries) == 0 {
		return
	}
	wd := entries[0].Weekday
	var rest []ScheduleEntry
	for _, e := range m.Schedule {
		if e.Weekday != wd {
			rest = append(rest, e)
		}
	}
	m.Schedule = append(rest, entries...)
	sortSchedule(m.Schedule)
}

// handleGovernorEvents processes OK:EVENTS. Event format (wire): id|title|at|location|notes, at = YYYY.MM.DD.HH.MM (dots).
func (m *Model) handleGovernorEvents(msg concentrator.Message) {
	events := parseGovernorEvents(msg.Args)
	m.Events = events
	// Clamp selected event index after list refresh
	dayEvents := m.eventsForSelectedDate()
	if m.SelectedEvent >= len(dayEvents) {
		m.SelectedEvent = len(dayEvents) - 1
	}
	if m.SelectedEvent < 0 {
		m.SelectedEvent = 0
	}
	m.requestGovernorDeadlines()
}

// handleGovernorEventCreated processes OK:EVENT:<id> (response to NEW:EVENT). Appends the event we just added and closes the form.
func (m *Model) handleGovernorEventCreated(msg concentrator.Message) {
	if !m.EventAddMenu || len(msg.Args) < 1 {
		return
	}
	id := msg.Args[0]
	t, err := parseGovernorEventTime(strings.ReplaceAll(m.EventAddDate, "-", ":") + ":" + m.EventAddTime)
	if err != nil {
		m.eventAddReset()
		m.requestGovernorEvents()
		m.requestGovernorDeadlines()
		return
	}
	m.Events = append(m.Events, Event{
		ID:       id,
		Date:     t,
		Title:    m.EventAddTitle,
		Category: "personal",
		Location: m.EventAddLocation,
		Notes:    m.EventAddNotes,
	})
	sortEvents(m.Events)
	m.eventAddReset()
	m.requestGovernorDeadlines()
}

func (m *Model) eventAddReset() {
	m.EventAddMenu = false
	m.EventAddFocusField = 0
	m.EventAddTitle = ""
	m.EventAddDate = ""
	m.EventAddTime = ""
	m.EventAddLocation = ""
	m.EventAddNotes = ""
	m.EventAddVisibleFrom = ""
}

// eventAddFocusedValue returns a pointer to the string field that has focus (for editing).
func (m *Model) eventAddFocusedValue() *string {
	switch m.EventAddFocusField {
	case 0:
		return &m.EventAddTitle
	case 1:
		return &m.EventAddDate
	case 2:
		return &m.EventAddTime
	case 3:
		return &m.EventAddLocation
	case 4:
		return &m.EventAddNotes
	case 5:
		return &m.EventAddVisibleFrom
	default:
		return &m.EventAddTitle
	}
}

// handleEventAddKeys handles key input for the add-event form (Calendar sheet). Returns true if key was consumed.
func (m *Model) handleEventAddKeys(msg tea.KeyMsg) bool {
	if !m.EventAddMenu || m.ActiveSheet != SheetCalendar {
		return false
	}
	key := msg.String()
	switch key {
	case "esc":
		m.eventAddReset()
		return true
	case "tab":
		m.EventAddFocusField = (m.EventAddFocusField + 1) % 6
		return true
	case "shift+tab":
		m.EventAddFocusField = (m.EventAddFocusField + 5) % 6
		return true
	case "enter":
		if m.EventAddFocusField == 5 {
			// Last field: submit if valid
			if m.eventAddValidateAndSubmit() {
				return true
			}
		}
		m.EventAddFocusField = (m.EventAddFocusField + 1) % 6
		return true
	case "backspace":
		s := m.eventAddFocusedValue()
		runes := []rune(*s)
		if len(runes) > 0 {
			*s = string(runes[:len(runes)-1])
		}
		return true
	case " ":
		*m.eventAddFocusedValue() += " "
		return true
	}
	if msg.Type == tea.KeyRunes && len(msg.Runes) > 0 {
		*m.eventAddFocusedValue() += string(msg.Runes)
		return true
	}
	return true
}

// eventAddValidateAndSubmit checks required fields (title, date, time) and sends NEW:EVENT. Returns true if key was consumed.
func (m *Model) eventAddValidateAndSubmit() bool {
	if strings.TrimSpace(m.EventAddTitle) == "" {
		return true
	}
	if _, err := time.Parse("2006-01-02", m.EventAddDate); err != nil {
		return true
	}
	if m.EventAddTime != "" {
		if _, err := time.Parse("15:04", m.EventAddTime); err != nil {
			if _, err2 := time.Parse("15:04:05", m.EventAddTime); err2 != nil {
				return true
			}
		}
	} else {
		return true
	}
	m.eventAddSubmit()
	return true
}

// eventAddSubmit sends NEW:EVENT per governor protocol:
// NEW:EVENT:<title>:<date>:<time>[:location][:notes][:visible_from] -> OK:EVENT:<id>
// Date YYYY.MM.DD, time HH.MM or HH.MM.SS. visible_from optional (omit = 7 days before).
func (m *Model) eventAddSubmit() {
	dateWire := strings.ReplaceAll(m.EventAddDate, "-", ".")
	timeWire := strings.ReplaceAll(m.EventAddTime, ":", ".")
	args := []string{m.EventAddTitle, dateWire, timeWire}
	args = append(args, m.EventAddLocation)
	args = append(args, m.EventAddNotes)
	args = append(args, strings.ReplaceAll(m.EventAddVisibleFrom, "-", "."))
	m.HubSend("GOVERNOR", "NEW", "EVENT", args...)
	// Keep form open until OK:EVENT:<id> in handleGovernorEventCreated (then eventAddReset)
}

// handleGovernorDeadlines processes OK:DEADLINES. Same event wire format as OK:EVENTS.
func (m *Model) handleGovernorDeadlines(msg concentrator.Message) {
	m.Deadlines = parseGovernorEvents(msg.Args)
	sortEvents(m.Deadlines)
}

// parseGovernorEvents parses OK:EVENTS args into Event slice. One arg per event: id|title|at|location|notes|visible_from (visible_from optional).
func parseGovernorEvents(args []string) []Event {
	var out []Event
	for _, arg := range args {
		parts := strings.Split(arg, "|")
		if len(parts) < 3 {
			continue
		}
		id := parts[0]
		title := parts[1]
		atStr := strings.ReplaceAll(parts[2], ".", ":") // YYYY.MM.DD.HH.MM -> YYYY:MM:DD:HH:MM for parsing
		location := ""
		if len(parts) > 3 {
			location = parts[3]
		}
		notes := ""
		if len(parts) > 4 {
			notes = parts[4]
		}
		t, err := parseGovernorEventTime(atStr)
		if err != nil {
			continue
		}
		category := "personal"
		if strings.Contains(strings.ToLower(notes), "deadline") {
			category = "deadline"
		} else if strings.Contains(strings.ToLower(notes), "work") {
			category = "work"
		} else if strings.Contains(strings.ToLower(notes), "system") {
			category = "system"
		}
		out = append(out, Event{
			ID:       id,
			Date:     t,
			Title:    title,
			Category: category,
			Location: location,
			Notes:    notes,
		})
	}
	// Sort by date
	sortEvents(out)
	return out
}

// parseGovernorEventTime parses at string: dots or colons YYYY.MM.DD.HH.MM or YYYY.MM.DD.HH.MM.SS
func parseGovernorEventTime(s string) (time.Time, error) {
	s = strings.TrimSpace(s)
	// Normalize to one separator for parsing
	s = strings.ReplaceAll(s, ".", ":")
	parts := strings.Split(s, ":")
	if len(parts) < 5 {
		return time.Time{}, fmt.Errorf("need at least YYYY:MM:DD:HH:MM")
	}
	var y, mo, d, h, min, sec int
	fmt.Sscanf(parts[0], "%d", &y)
	fmt.Sscanf(parts[1], "%d", &mo)
	fmt.Sscanf(parts[2], "%d", &d)
	fmt.Sscanf(parts[3], "%d", &h)
	fmt.Sscanf(parts[4], "%d", &min)
	if len(parts) >= 6 {
		fmt.Sscanf(parts[5], "%d", &sec)
	}
	return time.Date(y, time.Month(mo), d, h, min, sec, 0, time.Local), nil
}

func sortEvents(events []Event) {
	for i := 0; i < len(events); i++ {
		for j := i + 1; j < len(events); j++ {
			if events[i].Date.After(events[j].Date) {
				events[i], events[j] = events[j], events[i]
			}
		}
	}
}

// parseGovernorScheduleSlots parses wire slot args into ScheduleEntry slice.
// One arg per slot: Weekday|Start|End|Title|Location|Tags (colons → dots on wire).
func parseGovernorScheduleSlots(args []string) []ScheduleEntry {
	var entries []ScheduleEntry
	for _, arg := range args {
		parts := strings.Split(arg, "|")
		if len(parts) < 6 {
			continue
		}
		wd := parseWeekday(parts[0])
		start := strings.ReplaceAll(parts[1], ".", ":")
		end := strings.ReplaceAll(parts[2], ".", ":")
		title := parts[3]
		location := parts[4]
		tags := strings.Split(parts[5], ";")
		for i, t := range tags {
			tags[i] = strings.TrimSpace(t)
		}
		entries = append(entries, ScheduleEntry{
			Weekday:  wd,
			Start:    start,
			End:      end,
			Title:    title,
			Location: location,
			Tags:     tags,
		})
	}
	return entries
}

func parseWeekday(s string) time.Weekday {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "mon":
		return time.Monday
	case "tue":
		return time.Tuesday
	case "wed":
		return time.Wednesday
	case "thu":
		return time.Thursday
	case "fri":
		return time.Friday
	case "sat":
		return time.Saturday
	case "sun":
		return time.Sunday
	default:
		return time.Sunday
	}
}

func sortSchedule(s []ScheduleEntry) {
	weekdayOrder := func(w time.Weekday) int {
		if w == time.Sunday {
			return 7
		}
		return int(w)
	}
	startMins := func(e ScheduleEntry) int {
		var h, m int
		fmt.Sscanf(e.Start, "%d:%d", &h, &m)
		return h*60 + m
	}
	for i := 0; i < len(s); i++ {
		for j := i + 1; j < len(s); j++ {
			wi, wj := weekdayOrder(s[i].Weekday), weekdayOrder(s[j].Weekday)
			if wi > wj || (wi == wj && startMins(s[i]) > startMins(s[j])) {
				s[i], s[j] = s[j], s[i]
			}
		}
	}
}

// handleFireAlert shows the popup when ACHTUNG sends ALL:FIRE:TIMER/ALARM:name.
func (m *Model) handleFireAlert(msg concentrator.Message) {
	if strings.ToUpper(msg.From) != "ACHTUNG" || strings.ToUpper(msg.Verb) != "FIRE" {
		return
	}
	noun := strings.ToUpper(msg.Noun)
	if noun != "TIMER" && noun != "ALARM" {
		return
	}
	if len(msg.Args) < 1 {
		return
	}
	m.FireAlert = FireAlert{Show: true, JobKind: noun, JobName: msg.Args[0]}
}

// dismissFireAlert turns off the buzzer and closes the popup.
func (m *Model) dismissFireAlert() {
	m.HubSend("VERTEX", "OFF", "BUZZ")
	m.FireAlert.Show = false
}

// selectedNodeIsAchtung returns true when the selected system node is ACHTUNG.
func (m *Model) selectedNodeIsAchtung() bool {
	if m.ActiveSheet != SheetSystem || m.SelectedNode >= len(m.Nodes) {
		return false
	}
	return strings.ToUpper(m.Nodes[m.SelectedNode].Name) == "ACHTUNG"
}

// requestAchtungList sends GET:LIST to ACHTUNG to refresh the jobs list.
func (m *Model) requestAchtungList() {
	m.HubSend("ACHTUNG", "GET", "LIST")
	m.LastAchtungSync = time.Now()
}

// parseAchtungEndTime parses remaining (timer) or due (alarm) into an absolute time.
// Returns nil if unparseable (Remaining will be shown from server as-is).
func parseAchtungEndTime(kind, remaining, due string) *time.Time {
	now := time.Now()
	kind = strings.ToUpper(kind)
	if kind == "TIMER" {
		var d time.Duration
		if v, err := time.ParseDuration(remaining); err == nil {
			d = v
		} else if sec, err := strconv.ParseInt(remaining, 10, 64); err == nil {
			d = time.Duration(sec) * time.Second
		} else {
			return nil
		}
		t := now.Add(d)
		return &t
	}
	if kind == "ALARM" {
		// ACHTUNG uses YYYY.MM.DD:HH.MM; also support common variants
		for _, layout := range []string{
			"2006.01.02:15.04", "2006.01.02:15.04:05",
			"2006-01-02 15:04", "2006-01-02T15:04", "2006-01-02 15:04:05",
			"02.01.2006 15:04",
		} {
			if t, err := time.ParseInLocation(layout, due, time.Local); err == nil {
				return &t
			}
		}
		return nil
	}
	return nil
}

// updateJobRemaining sets job.Remaining from job.EndTime (countdown until then).
func (m *Model) updateJobRemaining(job *AchtungJob) {
	if job.EndTime == nil {
		return
	}
	now := m.LastUpdate
	if now.IsZero() {
		now = time.Now()
	}
	left := job.EndTime.Sub(now)
	if left <= 0 {
		job.Remaining = "0s"
		return
	}
	job.Remaining = formatDuration(left)
}

// updateAchtungRemaining refreshes Remaining for all jobs that have EndTime set.
func (m *Model) updateAchtungRemaining() {
	for i := range m.AchtungJobs {
		m.updateJobRemaining(&m.AchtungJobs[i])
	}
}

// handleAchtungResponse processes OK:LIST, OK:JOB, OK:TIMER, OK:ALARM from ACHTUNG.
func (m *Model) handleAchtungResponse(msg concentrator.Message) {
	if strings.ToUpper(msg.From) != "ACHTUNG" || strings.ToUpper(msg.Verb) != "OK" {
		return
	}
	noun := strings.ToUpper(msg.Noun)
	args := msg.Args

	switch noun {
	case "LIST":
		// OK:LIST[:<kind>:<name>:...] — args are alternating kind, name
		var jobs []AchtungJob
		for i := 0; i+1 < len(args); i += 2 {
			jobs = append(jobs, AchtungJob{
				Kind: strings.ToUpper(args[i]),
				Name: args[i+1],
				Remaining: "—",
				Due:   "—",
			})
		}
		m.AchtungJobs = jobs
		if m.SelectedAchtungJob >= len(m.AchtungJobs) {
			if len(m.AchtungJobs) > 0 {
				m.SelectedAchtungJob = len(m.AchtungJobs) - 1
			} else {
				m.SelectedAchtungJob = 0
			}
		}
		// Request details for each job
		for _, j := range m.AchtungJobs {
			m.HubSend("ACHTUNG", "GET", "JOB", j.Name)
		}
	case "JOB":
		// OK:JOB:<kind>:<name>:<remaining>:<due>
		if len(args) < 4 {
			return
		}
		kind, name, remaining, due := args[0], args[1], args[2], args[3]
		for i := range m.AchtungJobs {
			if m.AchtungJobs[i].Name == name {
				m.AchtungJobs[i].Kind = strings.ToUpper(kind)
				m.AchtungJobs[i].Due = due
				m.AchtungJobs[i].EndTime = parseAchtungEndTime(kind, remaining, due)
				if m.AchtungJobs[i].EndTime == nil {
					m.AchtungJobs[i].Remaining = remaining
				} else {
					m.updateJobRemaining(&m.AchtungJobs[i])
				}
				break
			}
		}
	case "TIMER":
		m.requestAchtungList()
		m.achtungTimerReset()
	case "ALARM":
		m.requestAchtungList()
		m.achtungAlarmReset()
	}
}

func (m *Model) achtungTimerReset() {
	m.AchtungTimerMenu = false
	m.AchtungTimerDuration = ""
	m.AchtungTimerName = ""
	m.AchtungTimerFocusField = 0
}

func (m *Model) achtungAlarmReset() {
	m.AchtungAlarmMenu = false
	m.AchtungAlarmDate = ""
	m.AchtungAlarmTime = ""
	m.AchtungAlarmName = ""
	m.AchtungAlarmFocusField = 0
}

// achtungFormFocusedValue returns the string field that has focus (for editing).
func (m *Model) achtungFormFocusedValue() *string {
	if m.AchtungTimerMenu {
		switch m.AchtungTimerFocusField {
		case 0:
			return &m.AchtungTimerDuration
		case 1:
			return &m.AchtungTimerName
		}
	}
	if m.AchtungAlarmMenu {
		switch m.AchtungAlarmFocusField {
		case 0:
			return &m.AchtungAlarmDate
		case 1:
			return &m.AchtungAlarmTime
		case 2:
			return &m.AchtungAlarmName
		}
	}
	return &m.AchtungTimerName
}

// handleAchtungFormKeys handles key input for timer/alarm forms (all fields at once, like events).
func (m *Model) handleAchtungFormKeys(msg tea.KeyMsg) bool {
	key := msg.String()
	if m.AchtungTimerMenu {
		switch key {
		case "esc":
			m.achtungTimerReset()
			return true
		case "tab":
			m.AchtungTimerFocusField = (m.AchtungTimerFocusField + 1) % 2
			return true
		case "shift+tab":
			m.AchtungTimerFocusField = (m.AchtungTimerFocusField + 1) % 2
			return true
		case "enter":
			if m.AchtungTimerFocusField == 1 {
				m.achtungTimerSubmit()
				return true
			}
			m.AchtungTimerFocusField = (m.AchtungTimerFocusField + 1) % 2
			return true
		case "backspace":
			s := m.achtungFormFocusedValue()
			runes := []rune(*s)
			if len(runes) > 0 {
				*s = string(runes[:len(runes)-1])
			}
			return true
		case " ":
			*m.achtungFormFocusedValue() += " "
			return true
		}
		if msg.Type == tea.KeyRunes && len(msg.Runes) > 0 {
			*m.achtungFormFocusedValue() += string(msg.Runes)
			return true
		}
		return false
	}
	if m.AchtungAlarmMenu {
		switch key {
		case "esc":
			m.achtungAlarmReset()
			return true
		case "tab":
			m.AchtungAlarmFocusField = (m.AchtungAlarmFocusField + 1) % 3
			return true
		case "shift+tab":
			m.AchtungAlarmFocusField = (m.AchtungAlarmFocusField + 2) % 3
			return true
		case "enter":
			if m.AchtungAlarmFocusField == 2 {
				m.achtungAlarmSubmit()
				return true
			}
			m.AchtungAlarmFocusField = (m.AchtungAlarmFocusField + 1) % 3
			return true
		case "backspace":
			s := m.achtungFormFocusedValue()
			runes := []rune(*s)
			if len(runes) > 0 {
				*s = string(runes[:len(runes)-1])
			}
			return true
		case " ":
			*m.achtungFormFocusedValue() += " "
			return true
		}
		if msg.Type == tea.KeyRunes && len(msg.Runes) > 0 {
			*m.achtungFormFocusedValue() += string(msg.Runes)
			return true
		}
		return false
	}
	return false
}

func (m *Model) achtungTimerSubmit() {
	dur := strings.TrimSpace(m.AchtungTimerDuration)
	if parseDuration(dur) < 0 {
		return
	}
	name := strings.TrimSpace(m.AchtungTimerName)
	if name == "" {
		name = fmt.Sprintf("t_%s_%d", dur, time.Now().Unix())
	}
	m.HubSend("ACHTUNG", "NEW", "TIMER", name, dur)
	m.requestAchtungList()
	m.achtungTimerReset()
}

func (m *Model) achtungAlarmSubmit() {
	if _, err := time.Parse("2006-01-02", m.AchtungAlarmDate); err != nil {
		return
	}
	if _, err := time.Parse("15:04", m.AchtungAlarmTime); err != nil {
		return
	}
	name := strings.TrimSpace(m.AchtungAlarmName)
	if name == "" {
		name = fmt.Sprintf("alarm_%d", time.Now().Unix())
	}
	datetime := formatAchtungAlarmDateTime(m.AchtungAlarmDate, m.AchtungAlarmTime)
	m.HubSend("ACHTUNG", "NEW", "ALARM", name, datetime)
	m.requestAchtungList()
	m.achtungAlarmReset()
}

// formatAchtungAlarmDateTime formats date (YYYY-MM-DD) and time (HH:MM) for ACHTUNG: YYYY.MM.DD:HH.MM
func formatAchtungAlarmDateTime(date, timeStr string) string {
	date = strings.ReplaceAll(date, "-", ".")
	timeStr = strings.ReplaceAll(timeStr, ":", ".")
	return date + ":" + timeStr
}

// parseAlarmDateTime parses time-only (HH:MM) or full datetime into date and time strings.
// If only time is given: use today if that time is still ahead, otherwise tomorrow.
func parseAlarmDateTime(s string) (date, timeStr string) {
	s = strings.TrimSpace(s)
	now := time.Now().In(time.Local)

	// Try time-only first: HH:MM or HH:MM:SS (e.g. 08:00, 14:30)
	for _, layout := range []string{"15:04", "15:04:05"} {
		t, err := time.ParseInLocation(layout, s, time.Local)
		if err != nil {
			continue
		}
		// Same day at that time
		todayAt := time.Date(now.Year(), now.Month(), now.Day(), t.Hour(), t.Minute(), t.Second(), 0, time.Local)
		if !now.Before(todayAt) {
			tomorrow := todayAt.Add(24 * time.Hour)
			return tomorrow.Format("2006-01-02"), tomorrow.Format("15:04")
		}
		return todayAt.Format("2006-01-02"), todayAt.Format("15:04")
	}

	// Full date+time
	for _, layout := range []string{"2006-01-02 15:04", "2006-01-02T15:04", "2006-01-02 15:04:05", "02.01.2006 15:04"} {
		if t, err := time.ParseInLocation(layout, s, time.Local); err == nil {
			return t.Format("2006-01-02"), t.Format("15:04")
		}
	}
	return "", ""
}

// parseDuration returns duration in seconds if s is valid (e.g. "5m", "90s"), else -1.
func parseDuration(s string) int64 {
	if d, err := time.ParseDuration(s); err == nil && d > 0 {
		return int64(d.Seconds())
	}
	if n, err := strconv.ParseInt(s, 10, 64); err == nil && n > 0 {
		return n
	}
	return -1
}

// handleAchtungKeys handles keys when ACHTUNG is selected (form input goes to handleAchtungFormKeys).
func (m *Model) handleAchtungKeys(msg tea.KeyMsg) bool {
	key := msg.String()

	// On Home: when focus is ACHTUNG (or key t/a to focus and act), handle ACHTUNG keys
	if m.ActiveSheet == SheetHome {
		if key == "t" || key == "a" {
			m.HomeFocusAchtung = true
			if key == "t" {
				m.AchtungTimerMenu = true
				m.AchtungTimerFocusField = 0
				m.AchtungTimerDuration = ""
				m.AchtungTimerName = ""
				return true
			}
			if key == "a" {
				m.AchtungAlarmMenu = true
				m.AchtungAlarmFocusField = 0
				now := time.Now()
				m.AchtungAlarmDate = now.Format("2006-01-02")
				m.AchtungAlarmTime = "20:00"
				m.AchtungAlarmName = ""
				return true
			}
		}
		if !m.HomeFocusAchtung {
			return false
		}
	} else {
		return false
	}

	switch key {
	case "j", "down":
		if m.SelectedAchtungJob < len(m.AchtungJobs)-1 {
			m.SelectedAchtungJob++
		}
		return true
	case "k", "up":
		if m.SelectedAchtungJob > 0 {
			m.SelectedAchtungJob--
		}
		return true
	case "enter", " ":
		m.achtungStopSelectedJob()
		return true
	case "d", "backspace":
		m.achtungStopSelectedJob()
		return true
	case "t":
		m.AchtungTimerMenu = true
		return true
	case "a":
		m.AchtungAlarmMenu = true
		m.AchtungAlarmFocusField = 0
		now := time.Now()
		m.AchtungAlarmDate = now.Format("2006-01-02")
		m.AchtungAlarmTime = "20:00"
		m.AchtungAlarmName = ""
		return true
	}
	return false
}

func (m *Model) achtungStopSelectedJob() {
	if m.SelectedAchtungJob >= len(m.AchtungJobs) {
		return
	}
	j := &m.AchtungJobs[m.SelectedAchtungJob]
	m.HubSend("ACHTUNG", "STOP", j.Kind, j.Name)
}


// pollNodes sends PING and GET:UPTIME to each node on a regular interval.
func (m *Model) pollNodes() {
	now := m.LastUpdate
	for i := range m.Nodes {
		node := &m.Nodes[i]

		// Mark offline if no PONG for 3x ping interval
		if !node.LastSeen.IsZero() && now.Sub(node.LastSeen) > pingInterval*3 {
			node.Status = "offline"
			node.PingMs = 0
		}

		if node.PingSent.IsZero() || now.Sub(node.PingSent) >= pingInterval {
			m.pingNode(node)
		}
	}
}

func (m *Model) pingNode(node *SystemNode) {
	node.PingSent = time.Now()
	m.HubSend(node.Name, "PING", node.PingNoun)
	m.HubSend(node.Name, "GET", "UPTIME")
}

// pingSelectedNode triggers an immediate ping when Enter is pressed on a node
// in the System sheet.
func (m *Model) pingSelectedNode() {
	if m.ActiveSheet != SheetSystem {
		return
	}
	if m.SelectedNode >= len(m.Nodes) {
		return
	}
	m.pingNode(&m.Nodes[m.SelectedNode])
}

// handleNodeResponse processes PONG and OK:UPTIME responses.
func (m *Model) handleNodeResponse(msg concentrator.Message) {
	verb := strings.ToUpper(msg.Verb)
	from := strings.ToUpper(msg.From)

	for i := range m.Nodes {
		node := &m.Nodes[i]
		if node.Name != from {
			continue
		}

		switch verb {
		case "PONG":
			now := time.Now()
			node.Status = "online"
			node.LastSeen = now
			if !node.PingSent.IsZero() {
				node.PingMs = now.Sub(node.PingSent).Milliseconds()
			}
			if from == "GOVERNOR" {
				m.requestGovernorEvents()
				m.requestGovernorDeadlines()
			}

		case "OK":
			topic := strings.ToUpper(msg.Noun)
			if topic == "UPTIME" && len(msg.Args) >= 1 {
				node.Uptime = parseUptime(msg.Args[0])
			}
		}
		return
	}
}

// parseUptime handles both millisecond integers (VERTEX) and Go duration
// strings like "3h2m15s" (ACHTUNG).
func parseUptime(raw string) string {
	if ms, err := strconv.ParseInt(raw, 10, 64); err == nil {
		return formatDuration(time.Duration(ms) * time.Millisecond)
	}
	if d, err := time.ParseDuration(raw); err == nil {
		return formatDuration(d)
	}
	return raw
}

func formatDuration(d time.Duration) string {
	days := int(d.Hours()) / 24
	hours := int(d.Hours()) % 24
	mins := int(d.Minutes()) % 60
	secs := int(d.Seconds()) % 60

	if days > 0 {
		return fmt.Sprintf("%dd %dh %dm", days, hours, mins)
	}
	if hours > 0 {
		return fmt.Sprintf("%dh %dm %ds", hours, mins, secs)
	}
	if mins > 0 {
		return fmt.Sprintf("%dm %ds", mins, secs)
	}
	return fmt.Sprintf("%ds", secs)
}

// handleDeviceResponse matches incoming OK/ERR responses to pending devices.
//
// Response format: MONOVIEW:OK:<TOPIC>[:detail...]:VERTEX
//
//	OK:LAMP                -> toggle confirmed
//	OK:LED                 -> cycle/value confirmed
//	OK:LAMP:STATE:ON       -> GET response
//	OK:LED:MODE:BLINK      -> GET response
//	OK:LED:BRIGHT:128      -> GET response
func (m *Model) handleDeviceResponse(msg concentrator.Message) {
	verb := strings.ToUpper(msg.Verb)
	from := strings.ToUpper(msg.From)

	if verb != "OK" {
		return
	}

	topic := strings.ToUpper(msg.Noun)
	args := msg.Args
	upperArgs := make([]string, len(args))
	for i, a := range args {
		upperArgs[i] = strings.ToUpper(a)
	}

	// GET responses carry extra detail: OK:LAMP:STATE:ON or OK:LED:MODE:BLINK
	if len(upperArgs) >= 2 {
		m.applyGetResponse(from, topic, upperArgs)
		return
	}

	// Simple OK:<TOPIC> — find the pending device
	for i := range m.HomeDevices {
		dev := &m.HomeDevices[i]
		if strings.ToUpper(dev.Topic) != topic || strings.ToUpper(dev.Node) != from {
			continue
		}
		if !dev.Pending {
			continue
		}
		dev.Pending = false

		switch dev.Kind {
		case "toggle":
			switch dev.Status {
			case "on":
				dev.Status = "off"
			case "off", "unknown":
				dev.Status = "on"
			}
		case "cycle":
			dev.Status = dev.nextMode()
		}
		return
	}
}

// applyGetResponse handles detailed OK responses like OK:LAMP:STATE:ON.
func (m *Model) applyGetResponse(from, topic string, args []string) {
	prop := args[0]
	val := args[1]

	for i := range m.HomeDevices {
		dev := &m.HomeDevices[i]
		if strings.ToUpper(dev.Topic) != topic || strings.ToUpper(dev.Node) != from {
			continue
		}

		switch dev.Kind {
		case "toggle":
			if prop == "STATE" {
				dev.Status = strings.ToLower(val)
				dev.Pending = false
			}
		case "cycle":
			if prop == "MODE" {
				dev.Status = strings.ToLower(val)
				dev.Pending = false
			}
		case "value":
			if prop == strings.ToUpper(dev.Property) {
				if v, err := fmt.Sscanf(val, "%d", &dev.Val); v == 1 && err == nil {
					dev.Pending = false
				}
			}
		}
	}
}

// nextMode returns the mode after the current status in the Modes cycle.
func (dev *HomeDevice) nextMode() string {
	if len(dev.Modes) == 0 {
		return dev.Status
	}
	for i, m := range dev.Modes {
		if m == dev.Status {
			return dev.Modes[(i+1)%len(dev.Modes)]
		}
	}
	return dev.Modes[0]
}

// HubSend is a convenience for sending a command through the concentrator
// from any place that has access to the Model (key handlers, etc.).
func (m *Model) HubSend(to, verb, noun string, args ...string) {
	if m.Hub != nil {
		m.Hub.Send(to, verb, noun, args...)
		m.LastTx = time.Now()
	}
}

// eventsForSelectedDate returns events on the selected date, sorted by time.
func (m *Model) eventsForSelectedDate() []Event {
	var out []Event
	for _, e := range m.Events {
		if e.Date.YearDay() == m.SelectedDate.YearDay() && e.Date.Year() == m.SelectedDate.Year() {
			out = append(out, e)
		}
	}
	// sort by time
	for i := 0; i < len(out); i++ {
		for j := i + 1; j < len(out); j++ {
			if out[i].Date.After(out[j].Date) {
				out[i], out[j] = out[j], out[i]
			}
		}
	}
	return out
}

func (m *Model) navigateDown() {
	switch m.ActiveSheet {
	case SheetCalendar:
		if m.EventAddMenu {
			break
		}
		if m.CalendarFocusEvents {
			dayEvents := m.eventsForSelectedDate()
			if m.SelectedEvent < len(dayEvents)-1 {
				m.SelectedEvent++
			}
		} else {
			m.SelectedDate = m.SelectedDate.Add(7 * 24 * time.Hour)
			m.SelectedEvent = 0
		}
	case SheetDiary:
		if m.SelectedEntry < len(m.DiaryEntries)-1 {
			m.SelectedEntry++
		}
	case SheetHome:
		if m.HomeFocusAchtung {
			if m.SelectedAchtungJob < len(m.AchtungJobs)-1 {
				m.SelectedAchtungJob++
			}
		} else {
			if m.SelectedDevice < len(m.HomeDevices)-1 {
				m.SelectedDevice++
			}
		}
	case SheetSystem:
		if m.SelectedNode < len(m.Nodes)-1 {
			m.SelectedNode++
		}
	}
}

func (m *Model) navigateUp() {
	switch m.ActiveSheet {
	case SheetCalendar:
		if m.EventAddMenu {
			break
		}
		if m.CalendarFocusEvents {
			if m.SelectedEvent > 0 {
				m.SelectedEvent--
			}
		} else {
			m.SelectedDate = m.SelectedDate.Add(-7 * 24 * time.Hour)
			m.SelectedEvent = 0
		}
	case SheetDiary:
		if m.SelectedEntry > 0 {
			m.SelectedEntry--
		}
	case SheetHome:
		if m.HomeFocusAchtung {
			if m.SelectedAchtungJob > 0 {
				m.SelectedAchtungJob--
			}
		} else {
			if m.SelectedDevice > 0 {
				m.SelectedDevice--
			}
		}
	case SheetSystem:
		if m.SelectedNode > 0 {
			m.SelectedNode--
		}
	}
}

func (m *Model) navigateLeft() {
	switch m.ActiveSheet {
	case SheetCalendar:
		if !m.CalendarFocusEvents {
			m.SelectedDate = m.SelectedDate.Add(-24 * time.Hour)
			m.SelectedEvent = 0
		}
	case SheetHome:
		m.adjustValue(-m.homeStep())
	case SheetSystem:
		if m.SelectedNode > 0 {
			m.SelectedNode--
		}
	}
}

func (m *Model) navigateRight() {
	switch m.ActiveSheet {
	case SheetCalendar:
		if !m.CalendarFocusEvents {
			m.SelectedDate = m.SelectedDate.Add(24 * time.Hour)
			m.SelectedEvent = 0
		}
	case SheetHome:
		m.adjustValue(m.homeStep())
	case SheetSystem:
		if m.SelectedNode < len(m.Nodes)-1 {
			m.SelectedNode++
		}
	}
}

func (m *Model) deleteSelectedEvent() {
	dayEvents := m.eventsForSelectedDate()
	if len(dayEvents) == 0 || m.SelectedEvent < 0 || m.SelectedEvent >= len(dayEvents) {
		return
	}
	id := dayEvents[m.SelectedEvent].ID
	if id == "" {
		return
	}
	m.HubSend("GOVERNOR", "STOP", "EVENT", id)
	m.requestGovernorEvents()
	m.requestGovernorDeadlines()
	if m.SelectedEvent >= len(dayEvents)-1 {
		m.SelectedEvent--
	}
	if m.SelectedEvent < 0 {
		m.SelectedEvent = 0
	}
}

func (m *Model) homeStep() int {
	if m.SelectedDevice < len(m.HomeDevices) {
		if s := m.HomeDevices[m.SelectedDevice].Step; s > 0 {
			return s
		}
	}
	return 1
}

// toggleAction sends the appropriate command for the selected device.
//
//	toggle → VERTEX:TOGGLE:LAMP:MONOVIEW
//	cycle  → VERTEX:SET:LED:MODE:BLINK:MONOVIEW  or  VERTEX:OFF:LED:MONOVIEW
//	value  → VERTEX:SET:LED:BRIGHT:128:MONOVIEW
func (m *Model) toggleAction() {
	if m.ActiveSheet != SheetHome || m.SelectedDevice >= len(m.HomeDevices) {
		return
	}
	dev := &m.HomeDevices[m.SelectedDevice]
	dev.Pending = true

	switch dev.Kind {
	case "toggle":
		if dev.Status == "on" {
			m.HubSend(dev.Node, "OFF", dev.Topic)
		} else {
			m.HubSend(dev.Node, "ON", dev.Topic)
		}

	case "cycle":
		next := dev.nextMode()
		m.HubSend(dev.Node, "SET", dev.Topic, "MODE", strings.ToUpper(next))

	case "value":
		m.HubSend(dev.Node, "SET", dev.Topic, dev.Property, fmt.Sprintf("%d", dev.Val))
	}
}

func (m *Model) adjustValue(delta int) {
	if m.ActiveSheet != SheetHome || m.SelectedDevice >= len(m.HomeDevices) {
		return
	}
	dev := &m.HomeDevices[m.SelectedDevice]
	if dev.Kind != "value" {
		return
	}
	dev.Val += delta
	if dev.Val < dev.Min {
		dev.Val = dev.Min
	}
	if dev.Val > dev.Max {
		dev.Val = dev.Max
	}
	dev.Pending = true
	m.HubSend(dev.Node, "SET", dev.Topic, dev.Property, fmt.Sprintf("%d", dev.Val))
}
