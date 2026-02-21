package ui

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"monoview/pkg/concentrator"
)

// Home devices (VERTEX) and ACHTUNG timers/alarms.

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

func (m *Model) requestAchtungList() {
	m.HubSend("ACHTUNG", "GET", "LIST")
	m.LastAchtungSync = time.Now()
}

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

func (m *Model) updateAchtungRemaining() {
	for i := range m.AchtungJobs {
		m.updateJobRemaining(&m.AchtungJobs[i])
	}
}

func (m *Model) handleAchtungResponse(msg concentrator.Message) {
	if strings.ToUpper(msg.From) != "ACHTUNG" || strings.ToUpper(msg.Verb) != "OK" {
		return
	}
	noun := strings.ToUpper(msg.Noun)
	args := msg.Args

	switch noun {
	case "LIST":
		var jobs []AchtungJob
		for i := 0; i+1 < len(args); i += 2 {
			jobs = append(jobs, AchtungJob{
				Kind:      strings.ToUpper(args[i]),
				Name:      args[i+1],
				Remaining: "—",
				Due:       "—",
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
		for _, j := range m.AchtungJobs {
			m.HubSend("ACHTUNG", "GET", "JOB", j.Name)
		}
	case "JOB":
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

func formatAchtungAlarmDateTime(date, timeStr string) string {
	date = strings.ReplaceAll(date, "-", ".")
	timeStr = strings.ReplaceAll(timeStr, ":", ".")
	return date + ":" + timeStr
}

func parseAlarmDateTime(s string) (date, timeStr string) {
	s = strings.TrimSpace(s)
	now := time.Now().In(time.Local)

	for _, layout := range []string{"15:04", "15:04:05"} {
		t, err := time.ParseInLocation(layout, s, time.Local)
		if err != nil {
			continue
		}
		todayAt := time.Date(now.Year(), now.Month(), now.Day(), t.Hour(), t.Minute(), t.Second(), 0, time.Local)
		if !now.Before(todayAt) {
			tomorrow := todayAt.Add(24 * time.Hour)
			return tomorrow.Format("2006-01-02"), tomorrow.Format("15:04")
		}
		return todayAt.Format("2006-01-02"), todayAt.Format("15:04")
	}

	for _, layout := range []string{"2006-01-02 15:04", "2006-01-02T15:04", "2006-01-02 15:04:05", "02.01.2006 15:04"} {
		if t, err := time.ParseInLocation(layout, s, time.Local); err == nil {
			return t.Format("2006-01-02"), t.Format("15:04")
		}
	}
	return "", ""
}

func parseDuration(s string) int64 {
	if d, err := time.ParseDuration(s); err == nil && d > 0 {
		return int64(d.Seconds())
	}
	if n, err := strconv.ParseInt(s, 10, 64); err == nil && n > 0 {
		return n
	}
	return -1
}

func (m *Model) handleAchtungKeys(msg tea.KeyMsg) bool {
	key := msg.String()

	if m.ActiveSheet == SheetHome {
		if key == "t" || key == "a" {
			m.HomeFocusAchtung = true
			m.AchtungViewMenu = false
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
	}

	if m.HomeFocusAchtung {
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
			if m.AchtungViewMenu {
				m.AchtungViewMenu = false
				return true
			}
			if len(m.AchtungJobs) > 0 && m.SelectedAchtungJob < len(m.AchtungJobs) {
				m.AchtungViewMenu = true
			}
			return true
		case "d", "backspace":
			if m.AchtungViewMenu {
				m.AchtungViewMenu = false
			}
			m.achtungStopSelectedJob()
			return true
		case "t":
			m.AchtungTimerMenu = true
			m.AchtungTimerFocusField = 0
			m.AchtungTimerDuration = ""
			m.AchtungTimerName = ""
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
	}
	return false
}

func (m *Model) achtungStopSelectedJob() {
	if m.SelectedAchtungJob >= len(m.AchtungJobs) {
		return
	}
	name := m.AchtungJobs[m.SelectedAchtungJob].Name
	m.HubSend("ACHTUNG", "STOP", strings.ToUpper(m.AchtungJobs[m.SelectedAchtungJob].Kind), name)
	m.requestAchtungList()
	if m.SelectedAchtungJob >= len(m.AchtungJobs)-1 {
		m.SelectedAchtungJob--
	}
	if m.SelectedAchtungJob < 0 {
		m.SelectedAchtungJob = 0
	}
}

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

	if len(upperArgs) >= 2 {
		m.applyGetResponse(from, topic, upperArgs)
		return
	}

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

func (dev *HomeDevice) nextMode() string {
	if len(dev.Modes) == 0 {
		return dev.Status
	}
	for i, mode := range dev.Modes {
		if mode == dev.Status {
			return dev.Modes[(i+1)%len(dev.Modes)]
		}
	}
	return dev.Modes[0]
}
