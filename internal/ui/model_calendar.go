package ui

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"monoview/pkg/concentrator"
)

// Governor protocol handlers and event add flow.

func (m *Model) requestGovernorSchedule() {
	weekdays := []string{"Mon", "Tue", "Wed", "Thu", "Fri", "Sat"}
	for _, wd := range weekdays {
		m.HubSend("GOVERNOR", "GET", "SCHEDULE", wd)
	}
}

func (m *Model) requestGovernorEvents() {
	m.HubSend("GOVERNOR", "GET", "EVENTS")
}

func (m *Model) requestGovernorDeadlines() {
	m.HubSend("GOVERNOR", "GET", "DEADLINES")
}

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

func (m *Model) handleGovernorEvents(msg concentrator.Message) {
	events := parseGovernorEvents(msg.Args)
	m.Events = events
	dayEvents := m.eventsForSelectedDate()
	if m.SelectedEvent >= len(dayEvents) {
		m.SelectedEvent = len(dayEvents) - 1
	}
	if m.SelectedEvent < 0 {
		m.SelectedEvent = 0
	}
	m.requestGovernorDeadlines()
}

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

func (m *Model) eventAddSubmit() {
	dateWire := strings.ReplaceAll(m.EventAddDate, "-", ".")
	timeWire := strings.ReplaceAll(m.EventAddTime, ":", ".")
	args := []string{m.EventAddTitle, dateWire, timeWire}
	if m.EventAddLocation != "" {
		args = append(args, m.EventAddLocation)
	}
	if m.EventAddNotes != "" {
		args = append(args, m.EventAddNotes)
	}
	if m.EventAddVisibleFrom != "" {
		args = append(args, strings.ReplaceAll(m.EventAddVisibleFrom, "-", "."))
	}
	m.HubSend("GOVERNOR", "NEW", "EVENT", args...)
}

func (m *Model) handleGovernorDeadlines(msg concentrator.Message) {
	m.Deadlines = parseGovernorEvents(msg.Args)
	sortEvents(m.Deadlines)
}

func parseGovernorEvents(args []string) []Event {
	var out []Event
	for _, arg := range args {
		parts := strings.Split(arg, "|")
		if len(parts) < 3 {
			continue
		}
		id := parts[0]
		title := parts[1]
		atStr := strings.ReplaceAll(parts[2], ".", ":")
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
	sortEvents(out)
	return out
}

func parseGovernorEventTime(s string) (time.Time, error) {
	s = strings.TrimSpace(s)
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
