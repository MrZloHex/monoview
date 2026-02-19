package ui

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"monoview/pkg/concentrator"
)

const pingInterval = 2 * time.Minute


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
			{Name: "VERTEX", Status: "unknown", Uptime: "—"},
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
			m.pingSelectedNode()
		}

	case tea.WindowSizeMsg:
		m.Width = msg.Width
		m.Height = msg.Height

	case TickMsg:
		m.LastUpdate = time.Time(msg)
		m.pollNodes()
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

	m.handleNodeResponse(msg)
	m.handleDeviceResponse(msg)
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
	m.HubSend(node.Name, "PING", "PINT")
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
				ms, err := strconv.ParseInt(msg.Args[0], 10, 64)
				if err == nil {
					node.Uptime = formatUptime(ms)
				}
			}
		}
		return
	}
}

func formatUptime(ms int64) string {
	d := time.Duration(ms) * time.Millisecond
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
