package ui

import (
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
	Nodes            []SystemNode
	Logs             []LogEntry
	SelectedNode     int
	SystemFocusLogs  bool   // true = j/k scroll logs; Tab toggles
	LogScrollOffset  int    // 0 = newest at top; scroll up (k) increases to see older

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
	AchtungViewMenu     bool   // Enter on job shows details in right panel
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
			m.SystemFocusLogs = false
		case "tab":
			if m.ActiveSheet == SheetHome {
				m.HomeFocusAchtung = !m.HomeFocusAchtung
			} else if m.ActiveSheet == SheetSystem {
				m.SystemFocusLogs = !m.SystemFocusLogs
			}
		case "shift+tab":
			if m.ActiveSheet == SheetHome {
				m.HomeFocusAchtung = !m.HomeFocusAchtung
			} else if m.ActiveSheet == SheetSystem {
				m.SystemFocusLogs = !m.SystemFocusLogs
			}
		case "esc":
			if m.ActiveSheet == SheetCalendar {
				if m.EventViewMenu {
					m.EventViewMenu = false
				} else if m.CalendarFocusEvents {
					m.CalendarFocusEvents = false
				}
			} else if m.ActiveSheet == SheetHome && m.AchtungViewMenu {
				m.AchtungViewMenu = false
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
				if m.ActiveSheet == SheetSystem && m.SystemFocusLogs {
					m.scrollLogsDown()
				} else {
					m.navigateDown()
				}
			case "k", "up":
				if m.ActiveSheet == SheetSystem && m.SystemFocusLogs {
					m.scrollLogsUp()
				} else {
					m.navigateUp()
				}
			case "left", "h":
				m.navigateLeft()
			case "right", "l":
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
	// When scrolled up, keep viewport stable: new log prepended shifts indices
	if m.LogScrollOffset > 0 {
		m.LogScrollOffset++
		if m.LogScrollOffset >= len(m.Logs) {
			m.LogScrollOffset = len(m.Logs) - 1
		}
	}

	m.handleNodeResponse(msg)
	m.handleGovernorResponse(msg)
	m.handleDeviceResponse(msg)
	m.handleAchtungResponse(msg)
	m.handleFireAlert(msg)
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
		if !m.SystemFocusLogs {
			m.systemNodeGridDown()
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
		if !m.SystemFocusLogs {
			m.systemNodeGridUp()
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
		if !m.SystemFocusLogs {
			m.systemNodeGridLeft()
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
		if !m.SystemFocusLogs {
			m.systemNodeGridRight()
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

