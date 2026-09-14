package app

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/MrZloHex/monolink"
	"monoview/internal/types"
)

// Home devices (VERTEX) and ACHTUNG timers/alarms.

// propOf is the property of vertex a device shows and sets, as uart2ws
// registers it: LAMP.STATE, LED.MODE, LED.BRIGHT. "" for an action.
func propOf(dev *types.HomeDevice) string {
	switch dev.Kind {
	case "toggle":
		return dev.Topic + ".STATE"
	case "cycle":
		return dev.Topic + ".MODE"
	case "value":
		return dev.Topic + "." + dev.Property
	}
	return ""
}

func (m *Model) queryDeviceStates() {
	seen := map[string]bool{}
	for i := range m.HomeDevices {
		dev := &m.HomeDevices[i]
		if p := propOf(dev); p != "" && !seen[dev.Node+":"+p] {
			seen[dev.Node+":"+p] = true
			m.HubSend(dev.Node, "GET", p)
		}
	}
}

func (m *Model) requestAchtungList() {
	m.achtungLoading, m.achtungPages = nil, 0
	m.HubSend("ACHTUNG", "GET", "LIST")
	m.LastAchtungSync = time.Now()
}

// maxAchtungPages bounds how many GET:LIST pages one refresh asks for.
const maxAchtungPages = 64

// achtungListed is whether this refresh has a job of that name already.
func (m *Model) achtungListed(name string) bool {
	for _, j := range m.achtungLoading {
		if j.Name == name {
			return true
		}
	}
	return false
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
	// ALARM, EVERY and DAILY all report an absolute next-fire time.
	//
	// The canonical form is achtung's serializeTimeLocal: a single
	// colon-free, zero-padded token, "2026.09.10.07.05". The older
	// colon-bearing layouts are kept because a frame from an achtung that
	// has not been updated still parses -- back then the colon split the
	// due across two fields, so `due` arrived as the date alone and this
	// never matched anything, which is why alarm countdowns were blank.
	if kind == "ALARM" || kind == "EVERY" || kind == "DAILY" {
		for _, layout := range []string{
			"2006.01.02.15.04", "2006.01.02.15.04.05",
			"2006.1.2.15.4",
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

func (m *Model) updateJobRemaining(job *types.AchtungJob) {
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

func (m *Model) handleAchtungResponse(msg monolink.Message) {
	if strings.ToUpper(msg.From) != "ACHTUNG" || strings.ToUpper(msg.Verb) != "OK" {
		return
	}
	noun := strings.ToUpper(msg.Noun)
	args := msg.Args

	switch noun {
	case "LIST":
		// ACHTUNG gives as many jobs as a frame holds, by name, and those
		// after a name when asked. A page that adds nothing new — empty, or
		// an older achtung's whole list again — is the end.
		added := 0
		for i := 0; i+1 < len(args); i += 2 {
			if m.achtungListed(args[i+1]) {
				continue
			}
			m.achtungLoading = append(m.achtungLoading, types.AchtungJob{
				Kind:      strings.ToUpper(args[i]),
				Name:      args[i+1],
				Remaining: "—",
				Due:       "—",
			})
			added++
		}
		m.achtungPages++
		if added > 0 && m.achtungPages < maxAchtungPages {
			m.HubSend("ACHTUNG", "GET", "LIST", args[len(args)-1])
			return
		}
		jobs := m.achtungLoading
		m.achtungLoading, m.achtungPages = nil, 0
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

// achtungFormOpen reports whether any ACHTUNG creation form is showing.
// Four kinds now, so the disjunction lives in one place.
func (m Model) achtungFormOpen() bool {
	return m.AchtungTimerMenu || m.AchtungAlarmMenu ||
		m.AchtungEveryMenu || m.AchtungDailyMenu || m.AchtungMorningMenu
}

// MorningJobName is the ACHTUNG job whose ALL:FIRE makes UKAZ print the
// agenda. It has to match CONFIG_MORNING_JOB in the ukaz firmware -- a
// daily job by any other name fires and nothing prints, with no error
// anywhere to explain why. Hence the dedicated [m] form rather than
// leaving it to whatever gets typed into the generic daily one.
const MorningJobName = "morning"

func (m *Model) achtungMorningReset() {
	m.AchtungMorningMenu = false
	m.AchtungMorningTime = ""
}

// achtungMorningExisting returns the time of the current morning job, as
// HH:MM, or "" if there is not one.
func (m Model) achtungMorningExisting() string {
	for _, j := range m.AchtungJobs {
		if j.Name != MorningJobName || strings.ToUpper(j.Kind) != "DAILY" {
			continue
		}
		if j.EndTime != nil {
			return j.EndTime.Format("15:04")
		}
	}
	return ""
}

func (m *Model) achtungMorningSubmit() {
	t, err := time.Parse("15:04", strings.TrimSpace(m.AchtungMorningTime))
	if err != nil {
		return
	}
	// ACHTUNG replaces a job of the same name, so this edits as well as
	// creates. H.M because the wire has no colons inside a field.
	m.HubSend("ACHTUNG", "NEW", "DAILY", MorningJobName,
		fmt.Sprintf("%d.%d", t.Hour(), t.Minute()))
	m.requestAchtungList()
	m.achtungMorningReset()
}

func (m *Model) achtungMorningOpen() {
	m.AchtungMorningMenu = true
	if cur := m.achtungMorningExisting(); cur != "" {
		m.AchtungMorningTime = cur
	} else {
		m.AchtungMorningTime = "07:00"
	}
}

func (m *Model) achtungEveryReset() {
	m.AchtungEveryMenu = false
	m.AchtungEveryInterval = ""
	m.AchtungEveryName = ""
	m.AchtungEveryFocusField = 0
}

func (m *Model) achtungDailyReset() {
	m.AchtungDailyMenu = false
	m.AchtungDailyTime = ""
	m.AchtungDailyName = ""
	m.AchtungDailyFocusField = 0
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
	if m.AchtungEveryMenu {
		switch m.AchtungEveryFocusField {
		case 0:
			return &m.AchtungEveryInterval
		case 1:
			return &m.AchtungEveryName
		}
	}
	if m.AchtungDailyMenu {
		switch m.AchtungDailyFocusField {
		case 0:
			return &m.AchtungDailyTime
		case 1:
			return &m.AchtungDailyName
		}
	}
	return &m.AchtungTimerName
}

// achtungTwoFieldFormKeys drives the EVERY and DAILY forms, which share a
// shape: two fields, Enter on the last one submits.
func (m *Model) achtungTwoFieldFormKeys(msg tea.KeyMsg, focus *int, reset func(), submit func()) bool {
	switch msg.String() {
	case "esc":
		reset()
		return true
	case "tab", "shift+tab":
		*focus = (*focus + 1) % 2
		return true
	case "enter":
		if *focus == 1 {
			submit()
			return true
		}
		*focus = 1
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

func (m *Model) achtungEverySubmit() {
	iv := strings.TrimSpace(m.AchtungEveryInterval)
	if parseDuration(iv) < 0 {
		return
	}
	name := strings.TrimSpace(m.AchtungEveryName)
	if name == "" {
		name = fmt.Sprintf("e_%s_%d", iv, time.Now().Unix())
	}
	m.HubSend("ACHTUNG", "NEW", "EVERY", name, iv)
	m.requestAchtungList()
	m.achtungEveryReset()
}

func (m *Model) achtungDailySubmit() {
	t, err := time.Parse("15:04", strings.TrimSpace(m.AchtungDailyTime))
	if err != nil {
		return
	}
	name := strings.TrimSpace(m.AchtungDailyName)
	if name == "" {
		name = fmt.Sprintf("d_%02d%02d", t.Hour(), t.Minute())
	}
	// achtung wants H.M -- the wire has no colons inside a field.
	m.HubSend("ACHTUNG", "NEW", "DAILY", name, fmt.Sprintf("%d.%d", t.Hour(), t.Minute()))
	m.requestAchtungList()
	m.achtungDailyReset()
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
	if m.AchtungEveryMenu {
		return m.achtungTwoFieldFormKeys(msg, &m.AchtungEveryFocusField,
			m.achtungEveryReset, m.achtungEverySubmit)
	}
	if m.AchtungDailyMenu {
		return m.achtungTwoFieldFormKeys(msg, &m.AchtungDailyFocusField,
			m.achtungDailyReset, m.achtungDailySubmit)
	}
	if m.AchtungMorningMenu {
		switch msg.String() {
		case "esc":
			m.achtungMorningReset()
			return true
		case "enter":
			m.achtungMorningSubmit()
			return true
		case "backspace":
			runes := []rune(m.AchtungMorningTime)
			if len(runes) > 0 {
				m.AchtungMorningTime = string(runes[:len(runes)-1])
			}
			return true
		}
		if msg.Type == tea.KeyRunes && len(msg.Runes) > 0 {
			m.AchtungMorningTime += string(msg.Runes)
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
	// Date and time are two arguments, as ACHTUNG reads them. v1 put the
	// same bytes on the wire when they were one "date:time" argument; v2
	// escapes a colon inside an argument, so only this form survives it.
	date, tm := formatAchtungAlarmDateTime(m.AchtungAlarmDate, m.AchtungAlarmTime)
	m.HubSend("ACHTUNG", "NEW", "ALARM", name, date, tm)
	m.requestAchtungList()
	m.achtungAlarmReset()
}

func formatAchtungAlarmDateTime(date, timeStr string) (string, string) {
	return strings.ReplaceAll(date, "-", "."), strings.ReplaceAll(timeStr, ":", ".")
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

	if m.ActiveSheet == types.SheetHome {
		if key == "m" {
			m.HomeFocusAchtung = true
			m.AchtungViewMenu = false
			m.achtungMorningOpen()
			return true
		}
		if key == "e" || key == "D" {
			m.HomeFocusAchtung = true
			m.AchtungViewMenu = false
			if key == "e" {
				m.AchtungEveryMenu = true
				m.AchtungEveryFocusField = 0
				m.AchtungEveryInterval = ""
				m.AchtungEveryName = ""
				return true
			}
			m.AchtungDailyMenu = true
			m.AchtungDailyFocusField = 0
			m.AchtungDailyTime = "07:00"
			m.AchtungDailyName = ""
			return true
		}
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
		case "e":
			m.AchtungEveryMenu = true
			m.AchtungEveryFocusField = 0
			m.AchtungEveryInterval = ""
			m.AchtungEveryName = ""
			return true
		case "D":
			m.AchtungDailyMenu = true
			m.AchtungDailyFocusField = 0
			m.AchtungDailyTime = "07:00"
			m.AchtungDailyName = ""
			return true
		case "m":
			m.achtungMorningOpen()
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

// handleDeviceResponse follows vertex's properties: the answer to a GET or a
// SET, and every PUB when one changes — whoever changed it.
func (m *Model) handleDeviceResponse(msg monolink.Message) {
	if msg.Verb != monolink.VerbOK && msg.Verb != monolink.VerbPub {
		return
	}
	from := strings.ToUpper(msg.From)
	if a, err := monolink.ParseAddress(msg.From); err == nil {
		from = strings.ToUpper(a.Node)
	}
	prop := strings.ToUpper(msg.Noun)
	for i := range m.HomeDevices {
		dev := &m.HomeDevices[i]
		if strings.ToUpper(dev.Node) != from {
			continue
		}
		if dev.Kind == "action" {
			if msg.Verb == monolink.VerbOK && strings.ToUpper(dev.Topic) == prop {
				dev.Pending = false
			}
			continue
		}
		if propOf(dev) != prop || len(msg.Args) == 0 {
			continue
		}
		switch dev.Kind {
		case "toggle", "cycle":
			dev.Status = strings.ToLower(msg.Args[0])
		case "value":
			if v, err := strconv.Atoi(msg.Args[0]); err == nil {
				dev.Val = v
			}
		}
		dev.Pending = false
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

// homeFocusNext cycles focus: VERTEX -> UKAZ -> ACHTUNG -> VERTEX.
func (m *Model) homeFocusNext() {
	if m.HomeFocusAchtung {
		m.HomeFocusAchtung = false
		m.HomeFocusUkaz = false
		m.SelectedDevice = m.firstDeviceIndexForNode("VERTEX")
	} else if m.HomeFocusUkaz {
		m.HomeFocusAchtung = true
		m.HomeFocusUkaz = false
	} else {
		m.HomeFocusUkaz = true
		m.SelectedDevice = m.firstDeviceIndexForNode("UKAZ")
	}
}

// homeFocusPrev cycles focus backwards: ACHTUNG -> UKAZ -> VERTEX -> ACHTUNG.
func (m *Model) homeFocusPrev() {
	if m.HomeFocusAchtung {
		m.HomeFocusUkaz = true
		m.HomeFocusAchtung = false
		m.SelectedDevice = m.firstDeviceIndexForNode("UKAZ")
	} else if m.HomeFocusUkaz {
		m.HomeFocusUkaz = false
		m.SelectedDevice = m.firstDeviceIndexForNode("VERTEX")
	} else {
		m.HomeFocusAchtung = true
		m.HomeFocusUkaz = false
	}
}

func (m *Model) firstDeviceIndexForNode(node string) int {
	for i, d := range m.HomeDevices {
		if strings.ToUpper(d.Node) == node {
			return i
		}
	}
	return 0
}

// homeDeviceIndicesForFocus returns device indices for the currently focused panel.
func (m *Model) homeDeviceIndicesForFocus() []int {
	node := "VERTEX"
	if m.HomeFocusUkaz {
		node = "UKAZ"
	}
	var out []int
	for i, d := range m.HomeDevices {
		if strings.ToUpper(d.Node) == node {
			out = append(out, i)
		}
	}
	return out
}

func (m *Model) toggleAction() {
	if m.ActiveSheet != types.SheetHome || m.SelectedDevice >= len(m.HomeDevices) {
		return
	}
	dev := &m.HomeDevices[m.SelectedDevice]
	dev.Pending = true

	switch dev.Kind {
	case "toggle":
		v := "ON"
		if dev.Status == "on" {
			v = "OFF"
		}
		m.HubSend(dev.Node, "SET", propOf(dev), v)

	case "cycle":
		m.HubSend(dev.Node, "SET", propOf(dev), strings.ToUpper(nextModeForDevice(dev)))

	case "value":
		m.HubSend(dev.Node, "SET", propOf(dev), strconv.Itoa(dev.Val))

	case "action":
		verb := dev.Property
		if verb == "" {
			verb = "PRINT"
		}
		m.HubSend(dev.Node, verb, dev.Topic)
	}
}

func (m *Model) adjustValue(delta int) {
	if m.ActiveSheet != types.SheetHome || m.SelectedDevice >= len(m.HomeDevices) {
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
	m.HubSend(dev.Node, "SET", propOf(dev), strconv.Itoa(dev.Val))
}

func nextModeForDevice(dev *types.HomeDevice) string {
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
