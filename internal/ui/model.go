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
	SelectedDate time.Time
	Events       []Event
	Schedule     []ScheduleEntry

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
	AchtungTimerMenu    bool   // true = show duration then name flow
	AchtungTimerCustom  bool   // true = in text input (duration or name step)
	AchtungTimerInput   string // current line: duration (step 1) or name (step 2)
	AchtungTimerDuration string // chosen duration (e.g. "5m"); when set we're in name step
	// Alarm flow: step 0 = type, step 1 = date/time, step 2 = name
	AchtungAlarmMenu    bool   // true = adding alarm
	AchtungAlarmStep    int    // 0 = pick type, 1 = pick/enter date&time, 2 = enter name
	AchtungAlarmType    string // "oneshot" (others later if protocol supports)
	AchtungAlarmDate    string // YYYY-MM-DD
	AchtungAlarmTime    string // HH:MM
	AchtungAlarmInput   string // custom datetime (step 1) or name (step 2)
	AchtungAlarmCustom  bool   // true = typing custom date&time in step 1
	HomeFocusAchtung    bool   // on Home: true = focus timers panel (j/k, enter, t, a, d)
	LastAchtungSync     time.Time

	// Fire alert popup (ALL:FIRE:TIMER/ALARM from ACHTUNG)
	FireAlert FireAlert

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

		Events: []Event{
			{Date: now, Title: "Team standup", Category: "work"},
			{Date: now.Add(2 * time.Hour), Title: "Code review", Category: "work"},
			{Date: now.Add(24 * time.Hour), Title: "Doctor appointment", Category: "personal"},
			{Date: now.Add(48 * time.Hour), Title: "Project deadline", Category: "deadline"},
			{Date: now.Add(72 * time.Hour), Title: "Server maintenance", Category: "system"},
		},

		Schedule: []ScheduleEntry{
			// Monday
			{Weekday: time.Monday, Start: "10:45", End: "12:10", Title: "ТФКП", Location: "Б.Хим", Tags: []string{"Lecture", "Math"}},
			{Weekday: time.Monday, Start: "17:05", End: "18:30", Title: "Машинное обучение", Location: "Б.Хим", Tags: []string{"Lecture", "ATP"}},
			{Weekday: time.Monday, Start: "18:35", End: "20:00", Title: "Машинное обучение", Location: "Б.Хим", Tags: []string{"Seminar", "ATP"}},
			// Tuesday
			{Weekday: time.Tuesday, Start: "10:45", End: "12:10", Title: "Мат. статистика", Location: "ГК 230", Tags: []string{"Seminar", "DM"}},
			{Weekday: time.Tuesday, Start: "13:55", End: "15:20", Title: "Китайский язык", Location: "НК", Tags: []string{"Seminar", "FL"}},
			{Weekday: time.Tuesday, Start: "15:30", End: "16:55", Title: "Комп. Сети", Location: "UNK", Tags: []string{"Seminar", "ATP"}},
			{Weekday: time.Tuesday, Start: "17:05", End: "18:30", Title: "ФИЯТ", Location: "КПМ 802", Tags: []string{"Seminar", "ATP"}},
			{Weekday: time.Tuesday, Start: "18:35", End: "20:00", Title: "Комп. Сети", Location: "UNK", Tags: []string{"Lecture", "ATP"}},
			// Wednesday
			{Weekday: time.Wednesday, Start: "12:20", End: "13:45", Title: "Функ. анализ", Location: "ГК 415", Tags: []string{"Seminar", "Math"}},
			{Weekday: time.Wednesday, Start: "13:55", End: "15:20", Title: "ТФКП", Location: "ГК 522", Tags: []string{"Seminar", "Math"}},
			{Weekday: time.Wednesday, Start: "17:05", End: "18:30", Title: "Unity", Location: "UNK", Tags: []string{"Lecture", "ATP"}},
			{Weekday: time.Wednesday, Start: "18:35", End: "20:00", Title: "Unity", Location: "UNK", Tags: []string{"Seminar", "ATP"}},
			// Thursday
			{Weekday: time.Thursday, Start: "09:00", End: "10:25", Title: "ФИЯТ", Location: "Б.Хим", Tags: []string{"Lecture", "ATP"}},
			{Weekday: time.Thursday, Start: "13:55", End: "15:20", Title: "Китайский язык", Location: "НК", Tags: []string{"Seminar", "FL"}},
			{Weekday: time.Thursday, Start: "17:05", End: "18:30", Title: "ШМП", Location: "UNK", Tags: []string{"Lecture", "Practic"}},
			{Weekday: time.Thursday, Start: "18:35", End: "20:00", Title: "ШМП", Location: "UNK", Tags: []string{"Seminar", "Practic"}},
			// Friday
			{Weekday: time.Friday, Start: "10:45", End: "12:10", Title: "Функ. анализ", Location: "КПМ 115", Tags: []string{"Lecture", "Math"}},
			{Weekday: time.Friday, Start: "15:30", End: "16:55", Title: "Мат. статистика", Location: "КПМ 115", Tags: []string{"Lecture", "DM"}},
			// Saturday
			{Weekday: time.Saturday, Start: "17:05", End: "18:30", Title: "Практикум матстат", Location: "ГК 113", Tags: []string{"Lecture", "DM"}},
		},

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
			{Name: "VERTEX", PingNoun: "PINT", Status: "unknown", Uptime: "—"},
			{Name: "ACHTUNG", PingNoun: "PING", Status: "unknown", Uptime: "—"},
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
		if m.handleAchtungAlarmInput(msg) {
			return m, nil
		}
		if m.handleAchtungTimerCustomInput(msg) {
			return m, nil
		}
		if m.handleAchtungKeys(msg) {
			return m, nil
		}
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit

		// Sheet navigation
		case "1":
			m.ActiveSheet = SheetCalendar
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
				if m.ActiveSheet == SheetHome {
					m.requestAchtungList()
				}
			}
		case "shift+tab":
			if m.ActiveSheet == SheetHome {
				m.HomeFocusAchtung = !m.HomeFocusAchtung
			} else {
				m.ActiveSheet = (m.ActiveSheet + 3) % 4
				if m.ActiveSheet == SheetHome {
					m.requestAchtungList()
				}
			}

		// Navigation within sheets
		case "j", "down":
			m.navigateDown()
		case "k", "up":
			m.navigateUp()
		case "h", "left":
			m.navigateLeft()
		case "l", "right":
			m.navigateRight()
		case "enter", " ":
			m.toggleAction()
			m.pingSelectedNode()
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
	m.handleDeviceResponse(msg)
	m.handleAchtungResponse(msg)
	m.handleFireAlert(msg)
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
	m.AchtungTimerCustom = false
	m.AchtungTimerInput = ""
	m.AchtungTimerDuration = ""
}

func (m *Model) achtungAlarmReset() {
	m.AchtungAlarmMenu = false
	m.AchtungAlarmStep = 0
	m.AchtungAlarmType = ""
	m.AchtungAlarmDate = ""
	m.AchtungAlarmTime = ""
	m.AchtungAlarmInput = ""
	m.AchtungAlarmCustom = false
}

// handleAchtungAlarmInput handles key input for alarm flow (date/time step then name step).
// Step 0 (type selection) is handled in handleAchtungKeys so menu keys like "1" are not eaten.
func (m *Model) handleAchtungAlarmInput(msg tea.KeyMsg) bool {
	if !m.AchtungAlarmMenu {
		return false
	}
	// Step 0 (type) and step 1 preset menu have no text input; let handleAchtungKeys handle 1, 2, c, esc.
	if m.AchtungAlarmStep == 0 || (m.AchtungAlarmStep == 1 && !m.AchtungAlarmCustom) {
		return false
	}
	key := msg.String()
	switch key {
	case "enter":
		if m.AchtungAlarmStep == 2 {
			if m.AchtungAlarmDate != "" && m.AchtungAlarmTime != "" {
				name := strings.TrimSpace(m.AchtungAlarmInput)
				if name == "" {
					name = fmt.Sprintf("alarm_%d", time.Now().Unix())
				}
				datetime := formatAchtungAlarmDateTime(m.AchtungAlarmDate, m.AchtungAlarmTime)
				m.HubSend("ACHTUNG", "NEW", "ALARM", name, datetime)
				m.requestAchtungList()
			}
			m.achtungAlarmReset()
		} else if m.AchtungAlarmCustom {
			date, timeStr := parseAlarmDateTime(strings.TrimSpace(m.AchtungAlarmInput))
			if date != "" && timeStr != "" {
				m.AchtungAlarmDate = date
				m.AchtungAlarmTime = timeStr
				m.AchtungAlarmStep = 2
				m.AchtungAlarmInput = ""
				m.AchtungAlarmCustom = false
			}
		}
		return true
	case "esc":
		m.achtungAlarmReset()
		return true
	case "backspace":
		runes := []rune(m.AchtungAlarmInput)
		if len(runes) > 0 {
			m.AchtungAlarmInput = string(runes[:len(runes)-1])
		}
		return true
	}
	if msg.Type == tea.KeyRunes && len(msg.Runes) > 0 {
		m.AchtungAlarmInput += string(msg.Runes)
		return true
	}
	return true
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

// handleAchtungTimerCustomInput handles key input for duration step (custom) and name step.
// Returns true if the key was consumed.
func (m *Model) handleAchtungTimerCustomInput(msg tea.KeyMsg) bool {
	if !m.AchtungTimerCustom {
		return false
	}
	key := msg.String()
	switch key {
	case "enter":
		if m.AchtungTimerDuration != "" {
			// Name step: submit timer
			name := strings.TrimSpace(m.AchtungTimerInput)
			if name == "" {
				name = fmt.Sprintf("t_%s_%d", m.AchtungTimerDuration, time.Now().Unix())
			}
			m.HubSend("ACHTUNG", "NEW", "TIMER", name, m.AchtungTimerDuration)
			m.requestAchtungList()
			m.achtungTimerReset()
		} else {
			// Duration step: validate and go to name step
			dur := strings.TrimSpace(m.AchtungTimerInput)
			if parseDuration(dur) >= 0 {
				m.AchtungTimerDuration = dur
				m.AchtungTimerInput = ""
			}
		}
		return true
	case "esc":
		m.achtungTimerReset()
		return true
	case "backspace":
		runes := []rune(m.AchtungTimerInput)
		if len(runes) > 0 {
			m.AchtungTimerInput = string(runes[:len(runes)-1])
		}
		return true
	}
	if msg.Type == tea.KeyRunes && len(msg.Runes) > 0 {
		m.AchtungTimerInput += string(msg.Runes)
		return true
	}
	return true
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

// handleAchtungKeys handles keys when ACHTUNG is selected or timer menu is open.
// Returns true if the key was consumed.
func (m *Model) handleAchtungKeys(msg tea.KeyMsg) bool {
	key := msg.String()

	// Step 1: duration selection (preset or [c] custom)
	if m.AchtungTimerMenu && !m.AchtungTimerCustom && m.AchtungTimerDuration == "" {
		if key == "q" || key == "ctrl+c" {
			return false
		}
		if key == "esc" {
			m.achtungTimerReset()
			return true
		}
		if key == "c" {
			m.AchtungTimerCustom = true
			m.AchtungTimerInput = ""
			return true
		}
		presets := map[string]string{"1": "1m", "2": "5m", "3": "10m", "4": "30m", "5": "1h"}
		if dur, ok := presets[key]; ok {
			m.AchtungTimerDuration = dur
			m.AchtungTimerCustom = true
			m.AchtungTimerInput = ""
			return true
		}
		return true
	}

	// Alarm step 0: select type
	if m.AchtungAlarmMenu && m.AchtungAlarmStep == 0 {
		if key == "q" || key == "ctrl+c" {
			return false
		}
		if key == "esc" {
			m.achtungAlarmReset()
			return true
		}
		if key == "1" {
			m.AchtungAlarmType = "oneshot"
			m.AchtungAlarmStep = 1
			return true
		}
		return true
	}

	// Alarm step 1: preset date/time
	if m.AchtungAlarmMenu && m.AchtungAlarmStep == 1 && !m.AchtungAlarmCustom {
		if key == "q" || key == "ctrl+c" {
			return false
		}
		if key == "esc" {
			m.achtungAlarmReset()
			return true
		}
		if key == "c" {
			m.AchtungAlarmCustom = true
			m.AchtungAlarmInput = ""
			return true
		}
		now := time.Now()
		if key == "1" {
			m.AchtungAlarmDate = now.Format("2006-01-02")
			m.AchtungAlarmTime = "20:00"
			m.AchtungAlarmStep = 2
			m.AchtungAlarmInput = ""
			return true
		}
		if key == "2" {
			tomorrow := now.Add(24 * time.Hour)
			m.AchtungAlarmDate = tomorrow.Format("2006-01-02")
			m.AchtungAlarmTime = "08:00"
			m.AchtungAlarmStep = 2
			m.AchtungAlarmInput = ""
			return true
		}
		return true
	}

	// On Home: when focus is ACHTUNG (or key t/a to focus and act), handle ACHTUNG keys
	if m.ActiveSheet == SheetHome {
		if key == "t" || key == "a" {
			m.HomeFocusAchtung = true
			if key == "t" {
				m.AchtungTimerMenu = true
				return true
			}
			if key == "a" {
				m.AchtungAlarmMenu = true
				m.AchtungAlarmStep = 0
				m.AchtungAlarmType = ""
				m.AchtungAlarmCustom = false
				m.AchtungAlarmInput = ""
				m.AchtungAlarmDate = ""
				m.AchtungAlarmTime = ""
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
		m.AchtungAlarmStep = 0
		m.AchtungAlarmType = ""
		m.AchtungAlarmCustom = false
		m.AchtungAlarmInput = ""
		m.AchtungAlarmDate = ""
		m.AchtungAlarmTime = ""
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

func (m *Model) navigateDown() {
	switch m.ActiveSheet {
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
		m.SelectedDate = m.SelectedDate.Add(-24 * time.Hour)
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
		m.SelectedDate = m.SelectedDate.Add(24 * time.Hour)
	case SheetHome:
		m.adjustValue(m.homeStep())
	case SheetSystem:
		if m.SelectedNode < len(m.Nodes)-1 {
			m.SelectedNode++
		}
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
