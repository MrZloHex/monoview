package app

import (
	"strings"
	"time"
	"unicode"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/MrZloHex/monolink"
	"github.com/MrZloHex/monolink/marshal"
	"monoview/internal/types"
)

const (
	pingInterval     = 2 * time.Minute
	achtungSyncEvery = 1 * time.Minute

	tickIntervalFast = 1 * time.Second  // when ACHTUNG countdowns need per-second updates
	tickIntervalIdle = 15 * time.Second // when idle: fewer wakeups, less CPU/redraws
)

// HubMsg wraps a concentrator message arriving through the inbox channel.
type HubMsg monolink.Message

// Model is the main application model
type Model struct {
	ActiveSheet types.Sheet
	Width       int
	Height      int
	LastUpdate  time.Time

	// Concentrator client (runs in background goroutine)
	Hub *monolink.Client

	// Calendar
	SelectedDate        time.Time
	SelectedEvent       int  // index into events for SelectedDate (when CalendarFocusEvents)
	CalendarFocusEvents bool // false = ↑/↓ move day (by week), Enter = focus events; true = ↑/↓ select event
	Events              []types.Event
	Deadlines           []types.Event // from GET:DEADLINES (upcoming deadlines box)
	eventsLoading       []types.Event // GET:EVENTS pages gathered so far
	eventsPages         int
	Schedule            []types.ScheduleEntry

	// Diary
	DiaryEntries  []types.DiaryEntry
	SelectedEntry int

	// Home
	HomeDevices    []types.HomeDevice
	SelectedDevice int

	// System
	Nodes               []types.SystemNode
	Logs                []types.LogEntry
	SelectedNode        int
	SystemFocusLogs     bool   // true = j/k scroll logs; Tab toggles
	LogScrollOffset     int    // 0 = newest at top; scroll up (k) increases to see older
	SystemCommandInput  bool   // true = typing custom message to bus (:)
	SystemCommandBuffer string // TO:VERB:NOUN[:args...]

	// ACHTUNG (timers & alarms, shown on Home sheet)
	AchtungJobs            []types.AchtungJob
	achtungLoading         []types.AchtungJob // GET:LIST pages gathered so far
	achtungPages           int
	SelectedAchtungJob     int
	AchtungTimerMenu       bool   // true = adding timer (all fields in right panel)
	AchtungTimerDuration   string // e.g. "5m"
	AchtungTimerName       string // optional, Enter for auto
	AchtungTimerFocusField int    // 0=duration, 1=name
	AchtungAlarmMenu       bool   // true = adding alarm (all fields in right panel)
	AchtungAlarmDate       string // YYYY-MM-DD
	AchtungAlarmTime       string // HH:MM
	AchtungAlarmName       string // optional
	AchtungAlarmFocusField int    // 0=date, 1=time, 2=name
	AchtungEveryMenu       bool   // true = adding a repeating interval job
	AchtungEveryInterval   string // e.g. "30m"
	AchtungEveryName       string // optional, Enter for auto
	AchtungEveryFocusField int    // 0=interval, 1=name
	AchtungDailyMenu       bool   // true = adding a wall-clock daily job
	AchtungDailyTime       string // HH:MM, local
	AchtungDailyName       string // optional
	AchtungDailyFocusField int    // 0=time, 1=name
	AchtungMorningMenu     bool   // true = setting the morning agenda print
	AchtungMorningTime     string // HH:MM, local
	HomeFocusAchtung       bool   // on Home: true = focus ACHTUNG panel (j/k, enter, t, a, d)
	HomeFocusUkaz          bool   // on Home: when false and !Achtung = VERTEX; when true = UKAZ
	AchtungViewMenu        bool   // Enter on job shows details in right panel
	LastAchtungSync        time.Time

	// Fire alert popup (ALL:FIRE:TIMER/ALARM from ACHTUNG)
	FireAlert types.FireAlert

	// Calendar: viewing selected event details in right panel (Enter on event)
	EventViewMenu bool

	// Add event flow (Calendar sheet): popup with all fields; EventAddFocusField = which field gets input
	EventAddMenu        bool // true = add-event form active
	EventAddFocusField  int  // 0=title, 1=date, 2=time, 3=location, 4=notes, 5=visible_from
	EventAddTitle       string
	EventAddDate        string // YYYY-MM-DD
	EventAddTime        string // HH:MM or HH:MM:SS
	EventAddLocation    string
	EventAddNotes       string
	EventAddVisibleFrom string // optional YYYY-MM-DD; omit = default (7 days before deadline)

	// Traffic indicators (timestamps of last rx/tx for arrow display)
	LastRx time.Time
	LastTx time.Time

	// People (MARSHAL): who is signed in at this panel, and the PEOPLE sheet
	Session          marshal.Session // zero when nobody is signed in
	SignedInAt       time.Time       // when Session was signed in: a change to people needs it recent
	Grants           []string        // the signed-in person's; checked before sending
	KeyPath          string          // this panel's key, sealed with its person's passphrase
	LastTicket       time.Time       // when the ticket at the hub was last renewed
	MarshalEnrolling bool            // MARSHAL awaits its first person
	People           []string
	PeopleGrants     map[string][]string
	PeopleKeys       map[string][]marshal.KeyInfo
	PeopleSessions   []marshal.SessionInfo
	Invitation       invitation // the last code NEW:INVITE gave
	PeopleSelected   int
	PeopleForm       peopleForm
	PeopleConfirm    string // person awaiting [y] to be removed
	PeopleStatus     string // the outcome of the last thing done on the sheet
	PeopleNote       string // why the list is short, when it is
	deniedLogged     map[string]bool

	// Messages (SYNAPSE): conversations with the other people of the bubble
	Synapse synapseState

	tickGen int // the one live tick chain; see TickMsg
}

// TickMsg is sent periodically (interval varies: fast when ACHTUNG countdowns, idle otherwise).
// Gen names the chain it belongs to; ticks from a superseded chain are dropped, so exactly
// one chain is ever running.
type TickMsg struct {
	Time time.Time
	Gen  int
}

// needsFastTick returns true when we need per-second ticks (e.g. ACHTUNG countdown display).
func (m *Model) needsFastTick() bool {
	for i := range m.AchtungJobs {
		if m.AchtungJobs[i].EndTime != nil {
			return true
		}
	}
	return false
}

func nextTickInterval(m *Model) time.Duration {
	if m.needsFastTick() {
		return tickIntervalFast
	}
	return tickIntervalIdle
}

func tickWithInterval(interval time.Duration, gen int) tea.Cmd {
	return tea.Tick(interval, func(t time.Time) tea.Msg {
		return TickMsg{Time: t, Gen: gen}
	})
}

// scheduleNextCmds returns the next periodic tick. Hub traffic is delivered via Program.Send
// from a single goroutine in main (see cmd/monoview) — do not tea.Batch a blocking inbox read
// with Tick: Bubble Tea runs Batch sub-commands under wg.Wait; waitForHub never completes on
// idle ticks, so each tick leaked a stuck execBatchMsg goroutine and an extra <-inbox waiter.
//
// Only the TickMsg handler may continue the chain. Scheduling from anywhere else forks a
// second chain that never ends — call restartTick instead.
func (m *Model) scheduleNextCmds() tea.Cmd {
	return tickWithInterval(nextTickInterval(m), m.tickGen)
}

// restartTick abandons the running chain and starts a fresh one at the current interval.
func (m *Model) restartTick() tea.Cmd {
	m.tickGen++
	return m.scheduleNextCmds()
}

// NewModel creates the initial model with sample data
func NewModel() Model {
	now := time.Now()
	return Model{
		ActiveSheet:  types.SheetCalendar,
		LastUpdate:   now,
		SelectedDate: now,

		// Events and Schedule are filled from GOVERNOR (GET:EVENTS, GET:SCHEDULE:<weekday>)
		Events:   nil,
		Schedule: nil,

		DiaryEntries: []types.DiaryEntry{
			{Date: now, Content: "Started working on MonoView TUI...", Mood: "focused"},
			{Date: now.Add(-24 * time.Hour), Content: "Fixed the WebSocket connection issues.", Mood: "productive"},
			{Date: now.Add(-48 * time.Hour), Content: "Rainy day. Read documentation.", Mood: "calm"},
		},
		SelectedEntry: 0,

		HomeDevices: []types.HomeDevice{
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
			{
				Name: "Print Deadlines", Node: "UKAZ", Topic: "DEADLINES",
				Kind: "action", Property: "PRINT", Status: "—",
			},
			{
				Name: "Print Status", Node: "UKAZ", Topic: "STATUS",
				Kind: "action", Property: "PRINT", Status: "—",
			},
			{
				Name: "Print Agenda", Node: "UKAZ", Topic: "AGENDA",
				Kind: "action", Property: "PRINT", Status: "—",
			},
		},
		SelectedDevice: 0,

		Nodes: []types.SystemNode{
			{Name: "VERTEX", PingNoun: "PING", Status: "offline", Uptime: "—"},
			{Name: "ACHTUNG", PingNoun: "PING", Status: "offline", Uptime: "—"},
			{Name: "GOVERNOR", PingNoun: "PING", Status: "offline", Uptime: "—"},
			{Name: "UKAZ", PingNoun: "PING", Status: "offline", Uptime: "—"},
			{Name: "MARSHAL", PingNoun: "PING", Status: "offline", Uptime: "—"},
			{Name: "SYNAPSE", PingNoun: "PING", Status: "offline", Uptime: "—"},
		},
		SelectedNode: 0,
	}
}

func (m Model) Init() tea.Cmd {
	var cmds []tea.Cmd
	if m.Hub != nil {
		m.queryDeviceStates()
		m.requestGovernorSchedule()
		m.requestGovernorEvents()
		m.requestGovernorDeadlines()
		cmds = append(cmds, (&m).refreshPeople())
	}
	return tea.Batch(append(cmds, (&m).scheduleNextCmds())...)
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		// While typing command, only command handler gets keys (disables all hotkeys)
		if m.SystemCommandInput {
			if m.handleSystemCommandKeys(msg) {
				return m, nil
			}
			// handleSystemCommandKeys returned false: allow q/ctrl+c to fall through to quit
		} else if m.FireAlert.Show {
			switch msg.String() {
			case "enter", " ", "q", "esc":
				m.dismissFireAlert()
				return m, nil
			}
		} else {
			if handled, cmd := m.handlePeopleKeys(msg); handled {
				return m, cmd
			}
			if handled, cmd := m.handleSynapseKeys(msg); handled {
				return m, cmd
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
			if m.handleSystemCommandKeys(msg) {
				return m, nil
			}
		}
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit

		// Sheet navigation
		case "1":
			m.ActiveSheet = types.SheetCalendar
			m.CalendarFocusEvents = false
			m.EventViewMenu = false
			m.SystemCommandInput = false
			if m.EventAddMenu {
				m.eventAddReset()
			}
		case "2":
			m.ActiveSheet = types.SheetDiary
			m.SystemCommandInput = false
		case "3":
			m.ActiveSheet = types.SheetHome
			m.SystemCommandInput = false
			m.requestAchtungList()
		case "4":
			m.ActiveSheet = types.SheetSystem
			m.SystemFocusLogs = false
			m.SystemCommandInput = false
		case "5":
			m.ActiveSheet = types.SheetPeople
			m.SystemCommandInput = false
			return m, m.refreshPeople()
		case "6":
			m.ActiveSheet = types.SheetSynapse
			m.SystemCommandInput = false
			return m, m.synapseRefresh()
		case ":":
			if m.ActiveSheet == types.SheetSystem && m.Hub != nil && !m.SystemCommandInput {
				m.SystemCommandInput = true
				m.SystemCommandBuffer = ""
			}
		case "tab":
			if m.ActiveSheet == types.SheetHome {
				m.homeFocusNext()
			} else if m.ActiveSheet == types.SheetSystem {
				m.SystemFocusLogs = !m.SystemFocusLogs
			}
		case "shift+tab":
			if m.ActiveSheet == types.SheetHome {
				m.homeFocusPrev()
			} else if m.ActiveSheet == types.SheetSystem {
				m.SystemFocusLogs = !m.SystemFocusLogs
			}
		case "esc":
			if m.SystemCommandInput {
				m.SystemCommandInput = false
				m.SystemCommandBuffer = ""
			} else if m.ActiveSheet == types.SheetCalendar {
				if m.EventViewMenu {
					m.EventViewMenu = false
				} else if m.CalendarFocusEvents {
					m.CalendarFocusEvents = false
				}
			} else if m.ActiveSheet == types.SheetHome && m.AchtungViewMenu {
				m.AchtungViewMenu = false
			}

		// Calendar: [a] or [n] add new event (opens form)
		case "a", "n":
			if m.ActiveSheet == types.SheetCalendar && !m.EventAddMenu && m.Hub != nil {
				m.EventAddMenu = true
				m.EventViewMenu = false
				m.EventAddFocusField = 0
				m.EventAddDate = m.SelectedDate.Format("2006-01-02")
			}
		}

		// Navigation within sheets (no-op when in add-event form)
		if m.ActiveSheet != types.SheetCalendar || !m.EventAddMenu {
			switch msg.String() {
			case "j", "down":
				if m.ActiveSheet == types.SheetSystem && m.SystemFocusLogs {
					m.scrollLogsDown()
				} else {
					m.navigateDown()
				}
			case "k", "up":
				if m.ActiveSheet == types.SheetSystem && m.SystemFocusLogs {
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
				if m.ActiveSheet == types.SheetCalendar && !m.CalendarFocusEvents {
					m.CalendarFocusEvents = true
					dayEvents := m.eventsForSelectedDate()
					if m.SelectedEvent >= len(dayEvents) {
						m.SelectedEvent = len(dayEvents) - 1
					}
					if m.SelectedEvent < 0 {
						m.SelectedEvent = 0
					}
				} else if m.ActiveSheet == types.SheetCalendar && m.CalendarFocusEvents {
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
				if m.ActiveSheet == types.SheetCalendar {
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
		if msg.Gen != m.tickGen {
			return m, nil // superseded chain: let it die
		}
		m.LastUpdate = msg.Time
		m.pollNodes()
		m.updateAchtungRemaining()
		if m.Hub != nil && m.Hub.Connected() && time.Since(m.LastAchtungSync) >= achtungSyncEvery {
			m.requestAchtungList()
			m.LastAchtungSync = time.Now()
		}
		return m, tea.Batch(m.scheduleNextCmds(), m.renewTicket())

	case sessionMsg, grantsMsg, peopleMsg, inviteMsg, enrollingMsg, peopleOpMsg, signedOutMsg, ticketMsg:
		return m, m.handlePeopleMsg(msg)

	case ReconnectedMsg:
		m.LastTicket = time.Time{}
		return m, m.renewTicket()

	case synapseListMsg, synapseConvMsg, synapseSentMsg:
		return m, m.handleSynapseMsg(msg)

	case HubMsg:
		hm := plain(monolink.Message(msg))
		wasFast := m.needsFastTick()
		m.handleHub(hm)
		m.updateAchtungRemaining()
		cmd := m.handleSynapsePub(hm)
		if !wasFast && m.needsFastTick() {
			// A countdown just appeared; don't sit out the rest of an idle interval.
			return m, tea.Batch(cmd, m.restartTick())
		}
		return m, cmd
	}

	return m, nil
}

// plain is msg with every control character taken out, before any of it is
// stored or drawn. Timer names, events, key labels and error details come
// from other nodes and other people; a terminal would act on an escape
// sequence among them — redraw the screen, retitle the window, write the
// clipboard — instead of showing it.
func plain(msg monolink.Message) monolink.Message {
	msg.Raw = stripControl(msg.Raw)
	msg.From = stripControl(msg.From)
	msg.Noun = stripControl(msg.Noun)
	args := make([]string, len(msg.Args))
	for i, a := range msg.Args {
		args[i] = stripControl(a)
	}
	msg.Args = args
	return msg
}

func stripControl(s string) string {
	return strings.Map(func(r rune) rune {
		if unicode.IsControl(r) || r == '\u2028' || r == '\u2029' ||
			(unicode.Is(unicode.Cf, r) && r != '\u200C' && r != '\u200D') {
			return -1
		}
		return r
	}, s)
}

// cleanAll is stripControl for a list.
//
// plain covers the inbox, but an answer to a Request never goes through it:
// it comes back to whoever asked. A node's ERR detail, a person's grants, a
// key's label, a message \u2014 all of it is another node's text, and reaches a
// terminal that would act on an escape sequence among it. So every answer
// passes through here or through stripControl before it is stored.
//
// Names are not cleaned but checked (marshal.ValidName): taking the control
// characters out of one would turn a name nobody has into a name someone
// does, and file a stranger's message under them.
func cleanAll(ss []string) []string {
	out := make([]string, len(ss))
	for i, s := range ss {
		out[i] = stripControl(s)
	}
	return out
}

// handleHub processes an incoming concentrator message and updates model state.
func (m *Model) handleHub(msg monolink.Message) {
	now := time.Now()
	m.LastRx = now

	m.Logs = append([]types.LogEntry{{
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
	m.handleMarshalPub(msg)
}

// HubSend is a convenience for sending a request through the concentrator
// from any place that has access to the Model (key handlers, etc.): in v2,
// all the enforcing hub carries, as the person signed in here — and only
// what their grants cover.
func (m *Model) HubSend(to, verb, noun string, args ...string) {
	if m.Hub == nil || !m.permitted(to, verb, noun) {
		return
	}
	m.Hub.SendMessage(monolink.Message{Version: monolink.V2, ID: m.Hub.NewID(), From: m.Hub.Address(),
		To: to, Verb: verb, Noun: noun, Args: args})
	m.LastTx = time.Now()
}

// loadSheets asks every node for what the sheets show — once someone is
// signed in, since nothing is sent before.
func (m *Model) loadSheets() {
	m.queryDeviceStates()
	m.requestGovernorSchedule()
	m.requestGovernorEvents()
	m.requestGovernorDeadlines()
	m.requestAchtungList()
}

// eventsForSelectedDate returns events on the selected date, sorted by time.
func (m *Model) eventsForSelectedDate() []types.Event {
	var out []types.Event
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
	case types.SheetCalendar:
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
	case types.SheetDiary:
		if m.SelectedEntry < len(m.DiaryEntries)-1 {
			m.SelectedEntry++
		}
	case types.SheetHome:
		if m.HomeFocusAchtung {
			if m.SelectedAchtungJob < len(m.AchtungJobs)-1 {
				m.SelectedAchtungJob++
			}
		} else {
			indices := m.homeDeviceIndicesForFocus()
			for i, idx := range indices {
				if idx == m.SelectedDevice && i < len(indices)-1 {
					m.SelectedDevice = indices[i+1]
					break
				}
			}
		}
	case types.SheetSystem:
		if !m.SystemFocusLogs {
			m.systemNodeGridDown()
		}
	}
}

func (m *Model) navigateUp() {
	switch m.ActiveSheet {
	case types.SheetCalendar:
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
	case types.SheetDiary:
		if m.SelectedEntry > 0 {
			m.SelectedEntry--
		}
	case types.SheetHome:
		if m.HomeFocusAchtung {
			if m.SelectedAchtungJob > 0 {
				m.SelectedAchtungJob--
			}
		} else {
			indices := m.homeDeviceIndicesForFocus()
			for i, idx := range indices {
				if idx == m.SelectedDevice && i > 0 {
					m.SelectedDevice = indices[i-1]
					break
				}
			}
		}
	case types.SheetSystem:
		if !m.SystemFocusLogs {
			m.systemNodeGridUp()
		}
	}
}

func (m *Model) navigateLeft() {
	switch m.ActiveSheet {
	case types.SheetCalendar:
		if !m.CalendarFocusEvents {
			m.SelectedDate = m.SelectedDate.Add(-24 * time.Hour)
			m.SelectedEvent = 0
		}
	case types.SheetHome:
		m.adjustValue(-m.homeStep())
	case types.SheetSystem:
		if !m.SystemFocusLogs {
			m.systemNodeGridLeft()
		}
	}
}

func (m *Model) navigateRight() {
	switch m.ActiveSheet {
	case types.SheetCalendar:
		if !m.CalendarFocusEvents {
			m.SelectedDate = m.SelectedDate.Add(24 * time.Hour)
			m.SelectedEvent = 0
		}
	case types.SheetHome:
		m.adjustValue(m.homeStep())
	case types.SheetSystem:
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
