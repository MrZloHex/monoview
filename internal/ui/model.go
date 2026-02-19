package ui

import (
	"fmt"
	"math/rand"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"monoview/pkg/concentrator"
)

func nounToStatus(noun string) string {
	switch noun {
	case "ON":
		return "on"
	case "OFF":
		return "off"
	default:
		return "on"
	}
}

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
				Name: "Desk Lamp", Node: "VERTEX", Verb: "LAMP",
				Kind: "toggle", Status: "unknown",
				Actions: map[string]string{
					"on": "OFF", "off": "ON", "unknown": "ON",
				},
			},
			{
				Name: "LED Light", Node: "VERTEX", Verb: "LED",
				Kind: "cycle", Status: "off",
				Actions: map[string]string{
					"off": "BLINK", "blink": "FADE", "fade": "SOLID", "solid": "OFF",
				},
			},
			{
				Name: "Brightness", Node: "VERTEX", Verb: "LED",
				Kind: "value", Status: "—", Noun: "BRIGHT",
				Val: 128, Min: 0, Max: 255, Step: 15,
			},
		},
		SelectedDevice: 0,

		Nodes: []SystemNode{
			{Name: "obelisk", Status: "online", CPU: 23.5, Memory: 45.2, Uptime: "14d 3h"},
			{Name: "vertex", Status: "online", CPU: 67.8, Memory: 72.1, Uptime: "7d 12h"},
			{Name: "nexus", Status: "offline", CPU: 0, Memory: 0, Uptime: "—"},
			{Name: "hal9000", Status: "online", CPU: 12.3, Memory: 38.9, Uptime: "30d 8h"},
		},
		Logs: []LogEntry{
			{Time: now, Level: "INFO", Source: "obelisk", Message: "System heartbeat OK"},
			{Time: now.Add(-30 * time.Second), Level: "WARN", Source: "vertex", Message: "High memory usage detected"},
			{Time: now.Add(-2 * time.Minute), Level: "INFO", Source: "hal9000", Message: "Backup completed"},
			{Time: now.Add(-5 * time.Minute), Level: "ERR", Source: "nexus", Message: "Connection lost"},
			{Time: now.Add(-10 * time.Minute), Level: "INFO", Source: "obelisk", Message: "Service restarted"},
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
	}
	return tea.Batch(cmds...)
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
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
		case "4":
			m.ActiveSheet = SheetSystem
		case "tab":
			m.ActiveSheet = (m.ActiveSheet + 1) % 4
		case "shift+tab":
			m.ActiveSheet = (m.ActiveSheet + 3) % 4

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
		}

	case tea.WindowSizeMsg:
		m.Width = msg.Width
		m.Height = msg.Height

	case TickMsg:
		m.LastUpdate = time.Time(msg)
		for i := range m.Nodes {
			if m.Nodes[i].Status == "online" {
				m.Nodes[i].CPU = Clamp(m.Nodes[i].CPU+rand.Float64()*6-3, 0, 100)
				m.Nodes[i].Memory = Clamp(m.Nodes[i].Memory+rand.Float64()*2-1, 0, 100)
			}
		}
		return m, tick()

	case HubMsg:
		m.handleHub(concentrator.Message(msg))
		var cmd tea.Cmd
		if m.Hub != nil {
			if inbox := m.Hub.Inbox(); inbox != nil {
				cmd = waitForHub(inbox)
			}
		}
		return m, cmd
	}

	return m, nil
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

	m.handleDeviceResponse(msg)
}

// handleDeviceResponse checks if the message is a device OK/status and
// updates the corresponding HomeDevice.
//
// Response patterns from the real system:
//
//	MONOVIEW:LAMP:OK:VERTEX   (FROM=VERTEX, VERB=LAMP, NOUN=OK)
//	MONOVIEW:LED:OK:VERTEX    (FROM=VERTEX, VERB=LED, NOUN=OK)
func (m *Model) handleDeviceResponse(msg concentrator.Message) {
	verb := strings.ToUpper(msg.Verb)
	noun := strings.ToUpper(msg.Noun)
	from := strings.ToUpper(msg.From)

	for i := range m.HomeDevices {
		dev := &m.HomeDevices[i]
		if strings.ToUpper(dev.Verb) != verb || strings.ToUpper(dev.Node) != from {
			continue
		}
		if !dev.Pending {
			continue
		}

		dev.Pending = false

		switch dev.Kind {
		case "toggle":
			if noun == "OK" {
				if next, ok := dev.Actions[dev.Status]; ok {
					dev.Status = strings.ToLower(next)
				}
			} else {
				dev.Status = nounToStatus(noun)
			}
		case "cycle":
			if noun == "OK" {
				if next, ok := dev.Actions[dev.Status]; ok {
					dev.Status = strings.ToLower(next)
				}
			} else {
				dev.Status = strings.ToLower(noun)
			}
		case "value":
			// OK confirms the value was accepted; nothing to change.
		}
		return
	}
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
		if m.SelectedDevice < len(m.HomeDevices)-1 {
			m.SelectedDevice++
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
		if m.SelectedDevice > 0 {
			m.SelectedDevice--
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
	}
}

func (m *Model) navigateRight() {
	switch m.ActiveSheet {
	case SheetCalendar:
		m.SelectedDate = m.SelectedDate.Add(24 * time.Hour)
	case SheetHome:
		m.adjustValue(m.homeStep())
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

func (m *Model) toggleAction() {
	if m.ActiveSheet != SheetHome || m.SelectedDevice >= len(m.HomeDevices) {
		return
	}
	dev := &m.HomeDevices[m.SelectedDevice]

	switch dev.Kind {
	case "toggle", "cycle":
		noun, ok := dev.Actions[dev.Status]
		if !ok {
			return
		}
		dev.Pending = true
		m.HubSend(dev.Node, dev.Verb, noun)
	case "value":
		dev.Pending = true
		m.HubSend(dev.Node, dev.Verb, dev.Noun, fmt.Sprintf("%d", dev.Val))
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
	m.HubSend(dev.Node, dev.Verb, dev.Noun, fmt.Sprintf("%d", dev.Val))
}
