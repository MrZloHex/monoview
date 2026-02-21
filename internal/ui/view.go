package ui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
)

func (m Model) View() string {
	if m.Width == 0 {
		return "Loading..."
	}

	var b strings.Builder

	b.WriteString(m.renderHeader())
	b.WriteString("\n")
	b.WriteString(m.renderTabs())
	b.WriteString("\n\n")

	switch m.ActiveSheet {
	case SheetCalendar:
		if m.EventAddMenu {
			b.WriteString(m.renderCalendarWithAddForm())
		} else {
			b.WriteString(m.renderCalendar())
		}
	case SheetDiary:
		b.WriteString(m.renderDiary())
	case SheetHome:
		showAchtungFormInline := !m.AchtungTimerMenu && !m.AchtungAlarmMenu
		b.WriteString(m.renderHome(showAchtungFormInline))
	case SheetSystem:
		b.WriteString(m.renderSystem())
	}

	content := b.String()
	footer := m.renderFooter()

	contentLines := strings.Count(content, "\n") + 1
	footerLines := 1
	padding := m.Height - contentLines - footerLines
	if padding < 0 {
		padding = 0
	}

	fullView := content + strings.Repeat("\n", padding) + footer

	// Right panel: Calendar (add-event, event details) or Home (timer/alarm forms).
	if m.ActiveSheet == SheetCalendar && (m.EventAddMenu || m.EventViewMenu) {
		return m.renderWithRightPanel(fullView)
	}
	if m.ActiveSheet == SheetHome && (m.AchtungTimerMenu || m.AchtungAlarmMenu) {
		return m.renderWithRightPanel(fullView)
	}

	// Fire alert still uses a centered popup (takes over the frame).
	if m.FireAlert.Show {
		return lipgloss.Place(m.Width, m.Height, lipgloss.Center, lipgloss.Center, m.renderFireAlertPopup())
	}

	return fullView
}

func (m Model) renderFireAlertPopup() string {
	const width = 44
	kind := m.FireAlert.JobKind
	name := m.FireAlert.JobName
	title := kind + " fired!"
	body := Accent.Render(name)
	action := "[ Enter ] Turn off buzzer"
	inner := strings.Join([]string{
		"",
		Title.Render("  "+title) + " ",
		"",
		"  " + body,
		"",
		Label.Render("  "+action) + " ",
		"",
	}, "\n")
	box := NewBox(width).WithBorderColor(GruvYellow).WithTitle(" ALARM ")
	return box.Render(inner)
}

func (m Model) renderEventAddForm() string {
	return m.renderEventAddFormInner(0)
}

// renderEventAddFormInner builds the form box. If minHeight > 0, pads inner content so the box spans minHeight lines (anchored right, full length).
func (m Model) renderEventAddFormInner(minHeight int) string {
	const width = 64
	labels := []string{"Title", "Date (YYYY-MM-DD)", "Time (HH:MM)", "Location", "Notes", "Visible from (opt)"}
	values := []string{m.EventAddTitle, m.EventAddDate, m.EventAddTime, m.EventAddLocation, m.EventAddNotes, m.EventAddVisibleFrom}
	focus := m.EventAddFocusField
	if focus < 0 || focus > 5 {
		focus = 0
	}
	var lines []string
	lines = append(lines, "")
	lines = append(lines, Title.Render("  New event")+" ")
	lines = append(lines, "")
	for i := 0; i < 6; i++ {
		line := Label.Render("  "+labels[i]+": ") + Value.Render(values[i])
		if i == focus {
			line += Dim.Render("▌")
		}
		lines = append(lines, line)
	}
	lines = append(lines, "")
	lines = append(lines, Label.Render("  [Esc] cancel  [Tab] next  [Enter] submit")+" ")
	lines = append(lines, "")
	inner := strings.Join(lines, "\n")
	// Pad so the box extends full height (minHeight - 2 for top/bottom border)
	if minHeight > 2 {
		innerLines := strings.Split(inner, "\n")
		needLines := minHeight - 2
		for len(innerLines) < needLines {
			innerLines = append(innerLines, "")
		}
		if len(innerLines) > needLines {
			innerLines = innerLines[:needLines]
		}
		inner = strings.Join(innerLines, "\n")
	}
	box := NewBox(width).WithBorderColor(GruvAqua).WithTitle(" ADD EVENT ")
	return box.Render(inner)
}

// renderEventDetailView shows the selected event's info in a right-side box (full height of plain).
func (m Model) renderEventDetailView(e Event, minHeight int) string {
	const width = 64
	var lines []string
	if e.ID == "" {
		lines = append(lines, "")
		lines = append(lines, Label.Render("  No event selected"))
		lines = append(lines, "")
		lines = append(lines, Dim.Render("  [Esc] close"))
	} else {
		lines = append(lines, "")
		lines = append(lines, Title.Render("  Event details")+" ")
		lines = append(lines, "")
		lines = append(lines, Label.Render("  Title: ")+Value.Render(e.Title))
		lines = append(lines, Label.Render("  Date: ")+Value.Render(e.Date.Format("2006-01-02")))
		lines = append(lines, Label.Render("  Time: ")+Value.Render(e.Date.Format("15:04")))
		lines = append(lines, Label.Render("  Category: ")+getCategoryIcon(e.Category)+" "+Value.Render(e.Category))
		lines = append(lines, Label.Render("  Location: ")+Value.Render(e.Location))
		lines = append(lines, Label.Render("  Notes: ")+Value.Render(e.Notes))
		lines = append(lines, "")
		lines = append(lines, Dim.Render("  [d] delete  [Esc] close"))
	}
	inner := strings.Join(lines, "\n")
	if minHeight > 2 {
		innerLines := strings.Split(inner, "\n")
		needLines := minHeight - 2
		for len(innerLines) < needLines {
			innerLines = append(innerLines, "")
		}
		if len(innerLines) > needLines {
			innerLines = innerLines[:needLines]
		}
		inner = strings.Join(innerLines, "\n")
	}
	box := NewBox(width).WithBorderColor(GruvAqua).WithTitle(" EVENT ")
	return box.Render(inner)
}

// renderAchtungFormBox renders the timer or alarm creation form (all fields at once, like events).
func (m Model) renderAchtungFormBox(minHeight int) string {
	const width = 64
	var lines []string
	lines = append(lines, "")
	if m.AchtungTimerMenu {
		lines = append(lines, Title.Render("  New timer")+" ")
		lines = append(lines, "")
		for i, label := range []string{"Duration (e.g. 5m, 1h)", "Name (optional)"} {
			val := m.AchtungTimerDuration
			if i == 1 {
				val = m.AchtungTimerName
			}
			line := Label.Render("  "+label+": ") + Value.Render(val)
			if i == m.AchtungTimerFocusField {
				line += Dim.Render("▌")
			}
			lines = append(lines, line)
		}
		lines = append(lines, "")
		lines = append(lines, Dim.Render("  [Tab] next  [Enter] submit  [Esc] cancel"))
	} else if m.AchtungAlarmMenu {
		lines = append(lines, Title.Render("  New alarm")+" ")
		lines = append(lines, "")
		for i, label := range []string{"Date (YYYY-MM-DD)", "Time (HH:MM)", "Name (optional)"} {
			var val string
			switch i {
			case 0:
				val = m.AchtungAlarmDate
			case 1:
				val = m.AchtungAlarmTime
			case 2:
				val = m.AchtungAlarmName
			}
			line := Label.Render("  "+label+": ") + Value.Render(val)
			if i == m.AchtungAlarmFocusField {
				line += Dim.Render("▌")
			}
			lines = append(lines, line)
		}
		lines = append(lines, "")
		lines = append(lines, Dim.Render("  [Tab] next  [Enter] submit  [Esc] cancel"))
	}
	inner := strings.Join(lines, "\n")
	if minHeight > 2 {
		innerLines := strings.Split(inner, "\n")
		needLines := minHeight - 2
		for len(innerLines) < needLines {
			innerLines = append(innerLines, "")
		}
		if len(innerLines) > needLines {
			innerLines = innerLines[:needLines]
		}
		inner = strings.Join(innerLines, "\n")
	}
	title := " ADD TIMER "
	if m.AchtungAlarmMenu {
		title = " ADD ALARM "
	}
	box := NewBox(width).WithBorderColor(GruvAqua).WithTitle(title)
	return box.Render(inner)
}

// renderCalendarWithAddForm is used when building content; the right-edge layout is done in renderWithRightPanel.
func (m Model) renderCalendarWithAddForm() string {
	return m.renderCalendar()
}

const addEventFormWidth = 64

// renderWithRightPanel draws the full view with a right-side panel (add form or event details).
// The panel starts where the plain (content) starts (below header/tabs) and spans full height.
func (m Model) renderWithRightPanel(fullView string) string {
	allLines := strings.Split(fullView, "\n")
	for len(allLines) < m.Height {
		allLines = append(allLines, "")
	}
	if len(allLines) > m.Height {
		allLines = allLines[:m.Height]
	}
	aboveContent := m.renderHeader() + "\n" + m.renderTabs() + "\n\n"
	aboveLines := strings.Count(aboveContent, "\n")
	contentHeight := m.Height - aboveLines
	if contentHeight < 1 {
		contentHeight = 1
	}
	leftWidth := m.Width - addEventFormWidth - 2
	if leftWidth < 8 {
		leftWidth = 8
	}
	var rightContent string
	if m.EventAddMenu {
		rightContent = m.renderEventAddFormInner(contentHeight)
	} else if m.EventViewMenu {
		dayEvents := m.eventsForSelectedDate()
		if len(dayEvents) > 0 && m.SelectedEvent >= 0 && m.SelectedEvent < len(dayEvents) {
			rightContent = m.renderEventDetailView(dayEvents[m.SelectedEvent], contentHeight)
		} else {
			rightContent = m.renderEventDetailView(Event{}, contentHeight)
		}
	} else if m.ActiveSheet == SheetHome && (m.AchtungTimerMenu || m.AchtungAlarmMenu) {
		rightContent = m.renderAchtungFormBox(contentHeight)
	}
	formLines := strings.Split(rightContent, "\n")
	for len(formLines) < contentHeight {
		formLines = append(formLines, "")
	}
	if len(formLines) > contentHeight {
		formLines = formLines[:contentHeight]
	}
	var out strings.Builder
	for i := 0; i < m.Height; i++ {
		line := allLines[i]
		if i < aboveLines {
			// Header/tabs: full width, no form on right
			if lipgloss.Width(line) < m.Width {
				line = PadLine(line, m.Width)
			} else if lipgloss.Width(line) > m.Width {
				line = truncateString(line, m.Width)
			}
			out.WriteString(line)
		} else {
			// Plain: left column truncated + form on right
			if lipgloss.Width(line) > leftWidth {
				line = truncateString(line, leftWidth)
			} else {
				line = PadLine(line, leftWidth)
			}
			out.WriteString(line)
			out.WriteString("  ")
			out.WriteString(formLines[i-aboveLines])
		}
		if i < m.Height-1 {
			out.WriteString("\n")
		}
	}
	return out.String()
}

func (m Model) renderHeader() string {
	logo := `
░█▄▒▄█░▄▀▄░█▄░█░▄▀▄░█▒█░█▒██▀░█░░▒█
░█▒▀▒█░▀▄▀░█▒▀█░▀▄▀░▀▄▀░█░█▄▄░▀▄▀▄▀`


	logoStyled := lipgloss.NewStyle().Foreground(GruvOrange).Render(logo)

	clock := m.LastUpdate.Format("15:04:05")
	date := m.LastUpdate.Format("Mon, 02 Jan 2006")

	var timeLines []string
	timeLines = append(timeLines, PadLine(" "+Value.Render(clock), 20))
	timeLines = append(timeLines, PadLine(" "+Label.Render(date), 20))
	timeBox := NewBox(22).Render(strings.Join(timeLines, "\n"))

	hubBox := m.renderHubStatus()

	rightPanel := lipgloss.JoinHorizontal(lipgloss.Top, hubBox, " ", timeBox)

	gap := m.Width - lipgloss.Width(logo) - lipgloss.Width(rightPanel) - 4
	if gap < 0 {
		gap = 2
	}

	return lipgloss.JoinHorizontal(
		lipgloss.Top,
		logoStyled,
		strings.Repeat(" ", gap),
		rightPanel,
	)
}

func (m Model) renderHubStatus() string {
	const width = 12
	now := time.Now()
	trafficWindow := 500 * time.Millisecond

	var dot string
	if m.Hub != nil && m.Hub.Connected() {
		dot = Online.Render("●")
	} else {
		dot = Offline.Render("●")
	}

	var statusLabel string
	if m.Hub != nil && m.Hub.Connected() {
		statusLabel = Online.Render("ONLINE")
	} else {
		statusLabel = Offline.Render("OFFLINE")
	}

	rxArrow := Dim.Render("▼")
	txArrow := Dim.Render("▲")
	if !m.LastRx.IsZero() && now.Sub(m.LastRx) < trafficWindow {
		rxArrow = lipgloss.NewStyle().Foreground(GruvAqua).Bold(true).Render("▼")
	}
	if !m.LastTx.IsZero() && now.Sub(m.LastTx) < trafficWindow {
		txArrow = lipgloss.NewStyle().Foreground(GruvOrange).Bold(true).Render("▲")
	}

	var lines []string
	lines = append(lines, PadLine(" "+dot+" "+statusLabel, width-2))
	lines = append(lines, PadLine(" "+txArrow+" "+rxArrow+" "+Label.Render("HUB"), width-2))

	content := strings.Join(lines, "\n")
	return NewBox(width).Render(content)
}

func (m Model) renderTabs() string {
	var tabs []string

	for i, name := range SheetNames {
		if Sheet(i) == m.ActiveSheet {
			tabs = append(tabs, TabActive.Render(name))
		} else {
			tabs = append(tabs, TabInactive.Render(name))
		}
	}

	tabBar := lipgloss.JoinHorizontal(lipgloss.Top, tabs...)
	line := Dim.Render(strings.Repeat("─", m.Width-2))

	return fmt.Sprintf("  %s\n  %s", tabBar, line)
}

func (m Model) renderFooter() string {
	var help string
	switch m.ActiveSheet {
	case SheetCalendar:
		if m.EventAddMenu {
			help = "[Tab] next field  [Shift+Tab] prev  [Enter] submit  [Esc] cancel  [a/n] add event"
		} else if m.EventViewMenu {
			help = "[d] delete event  [Esc] close  [a/n] add  [1-4] sheets  [q] quit"
		} else if m.CalendarFocusEvents {
			help = "[↑/↓] select event  [Enter] view  [d] delete  [Esc] back  [a/n] add  [1-4] sheets  [q] quit"
		} else {
			help = "[↑/↓] week  [←/→] day  [Enter] select day → events  [a/n] add  [1-4] sheets  [q] quit"
		}
	case SheetDiary:
		help = "[↑/k] prev  [↓/j] next  [1-4] sheets  [q] quit"
	case SheetHome:
		if m.AchtungTimerMenu || m.AchtungAlarmMenu {
			help = "[Tab] next field  [Enter] submit  [Esc] cancel  [t] timer  [a] alarm  [q] quit"
		} else if m.HomeFocusAchtung {
			help = "[tab] devices  [↑/k ↓/j] job  [t] timer  [a] alarm  [d] delete  [1-4] sheets  [q] quit"
		} else {
			help = "[tab] timers  [↑/k ↓/j] device  [enter] toggle  [←/h →/l] adjust  [1-4] sheets  [q] quit"
		}
	case SheetSystem:
		help = "[↑/k ↓/j] or [←/h →/l] select node  [enter] ping  [1-4] sheets  [q] quit"
	}

	return Help.Render("  " + help)
}
