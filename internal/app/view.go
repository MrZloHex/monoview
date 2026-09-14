package app

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"

	"monoview/internal/types"
	"monoview/internal/ui"
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
	case types.SheetCalendar:
		if m.EventAddMenu {
			b.WriteString(m.renderCalendarWithAddForm())
		} else {
			b.WriteString(m.renderCalendar())
		}
	case types.SheetDiary:
		b.WriteString(m.renderDiary())
	case types.SheetHome:
		showAchtungFormInline := !m.achtungFormOpen()
		b.WriteString(m.renderHome(showAchtungFormInline))
	case types.SheetSystem:
		b.WriteString(m.renderSystem())
	case types.SheetPeople:
		b.WriteString(m.renderPeople())
	case types.SheetSynapse:
		b.WriteString(m.renderSynapse())
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

	// Right panel: Calendar (add-event, event details) or Home (timer/alarm forms, job details).
	if m.ActiveSheet == types.SheetCalendar && (m.EventAddMenu || m.EventViewMenu) {
		return m.renderWithRightPanel(fullView)
	}
	// Home renders its ACHTUNG form/detail inside its own right column, so
	// it does not need the extra panel.

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
	body := ui.Accent.Render(name)
	action := "[ Enter ] Turn off buzzer"
	inner := strings.Join([]string{
		"",
		ui.Title.Render("  "+title) + " ",
		"",
		"  " + body,
		"",
		ui.Label.Render("  "+action) + " ",
		"",
	}, "\n")
	box := ui.NewBox(width).WithBorderColor(ui.GruvYellow).WithTitle(" ALARM ")
	return box.Render(inner)
}

// renderCalendarWithAddForm is used when building content; the right-edge layout is done in renderWithRightPanel.
func (m Model) renderCalendarWithAddForm() string {
	return m.renderCalendar()
}

const addEventFormWidth = 64

// plainHeight returns the number of lines available for sheet content (below header/tabs, above footer).
func (m Model) plainHeight() int {
	above := m.renderHeader() + "\n" + m.renderTabs() + "\n\n"
	return m.Height - strings.Count(above, "\n") - 1
}

// renderWithRightPanel draws the full view with a right-side panel (add form or event details).
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
			rightContent = m.renderEventDetailView(types.Event{}, contentHeight)
		}
	} else if m.ActiveSheet == types.SheetHome {
		if m.achtungFormOpen() {
			rightContent = m.renderAchtungFormBox(contentHeight)
		} else if m.AchtungViewMenu && m.SelectedAchtungJob < len(m.AchtungJobs) {
			rightContent = m.renderAchtungJobDetailView(m.AchtungJobs[m.SelectedAchtungJob], contentHeight)
		} else {
			rightContent = m.renderAchtungJobDetailView(types.AchtungJob{}, contentHeight)
		}
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
			if lipgloss.Width(line) < m.Width {
				line = ui.PadLine(line, m.Width)
			} else if lipgloss.Width(line) > m.Width {
				line = ui.TruncateString(line, m.Width)
			}
			out.WriteString(line)
		} else {
			if lipgloss.Width(line) > leftWidth {
				line = ui.TruncateString(line, leftWidth)
			} else {
				line = ui.PadLine(line, leftWidth)
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
	logo := ` _______  _____  __   _  _____         _____ _______ _     _
 |  |  | |     | | \  | |     | |        |      |    |_____|
 |  |  | |_____| |  \_| |_____| |_____ __|__    |    |     |`

	logoStyled := lipgloss.NewStyle().Foreground(ui.GruvOrange).Render(logo)

	clock := m.LastUpdate.Format("15:04:05")
	date := m.LastUpdate.Format("Mon, 02 Jan 2006")

	var timeLines []string
	timeLines = append(timeLines, ui.PadLine(" "+ui.Value.Render(clock), 20))
	timeLines = append(timeLines, ui.PadLine(" "+ui.Label.Render(date), 20))
	timeBox := ui.NewBox(22).Render(strings.Join(timeLines, "\n"))

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
		dot = ui.Online.Render("●")
	} else {
		dot = ui.Offline.Render("●")
	}

	var statusLabel string
	if m.Hub != nil && m.Hub.Connected() {
		statusLabel = ui.Online.Render("ONLINE")
	} else {
		statusLabel = ui.Offline.Render("OFFLINE")
	}

	rxArrow := ui.Dim.Render("▼")
	txArrow := ui.Dim.Render("▲")
	if !m.LastRx.IsZero() && now.Sub(m.LastRx) < trafficWindow {
		rxArrow = lipgloss.NewStyle().Foreground(ui.GruvAqua).Bold(true).Render("▼")
	}
	if !m.LastTx.IsZero() && now.Sub(m.LastTx) < trafficWindow {
		txArrow = lipgloss.NewStyle().Foreground(ui.GruvOrange).Bold(true).Render("▲")
	}

	var lines []string
	lines = append(lines, ui.PadLine(" "+dot+" "+statusLabel, width-2))
	lines = append(lines, ui.PadLine(" "+txArrow+" "+rxArrow+" "+ui.Label.Render("HUB"), width-2))

	content := strings.Join(lines, "\n")
	return ui.NewBox(width).Render(content)
}

func (m Model) renderTabs() string {
	var tabs []string

	for i, name := range types.SheetNames {
		if types.Sheet(i) == types.SheetSynapse && m.Synapse.unread > 0 {
			name += fmt.Sprintf(" (%d)", m.Synapse.unread)
		}
		if types.Sheet(i) == m.ActiveSheet {
			tabs = append(tabs, ui.TabActive.Render(name))
		} else {
			tabs = append(tabs, ui.TabInactive.Render(name))
		}
	}

	tabBar := lipgloss.JoinHorizontal(lipgloss.Top, tabs...)
	line := ui.Dim.Render(strings.Repeat("─", m.Width))

	who := ui.Dim.Render("not signed in")
	if m.signedIn() {
		who = ui.Accent.Render("● " + m.Session.User)
	}
	gap := max(m.Width-lipgloss.Width(tabBar)-lipgloss.Width(who)-4, 1)

	return fmt.Sprintf("  %s%s%s\n%s", tabBar, strings.Repeat(" ", gap), who, line)
}

func (m Model) renderFooter() string {
	var help string
	switch m.ActiveSheet {
	case types.SheetSynapse:
		switch {
		case m.Synapse.writing:
			help = "[Enter] send  [Esc] stop writing"
		case m.Synapse.peer != "":
			help = "[i] write  [u] earlier  [↑/↓] person  [Enter] open  [Esc] close  [r] refresh  [1-6] sheets  [q] quit"
		default:
			help = "[↑/↓] person  [Enter] open  [i] write  [r] refresh  [1-6] sheets  [q] quit"
		}
	case types.SheetCalendar:
		if m.EventAddMenu {
			help = "[Tab] next field  [Shift+Tab] prev  [Enter] submit  [Esc] cancel  [a/n] add event"
		} else if m.EventViewMenu {
			help = "[d] delete event  [Esc] close  [a/n] add  [1-6] sheets  [q] quit"
		} else if m.CalendarFocusEvents {
			help = "[↑/↓] select event  [Enter] view  [d] delete  [Esc] back  [a/n] add  [1-6] sheets  [q] quit"
		} else {
			help = "[↑/↓] week  [←/→] day  [Enter] select day → events  [a/n] add  [1-6] sheets  [q] quit"
		}
	case types.SheetDiary:
		help = "[↑/k] prev  [↓/j] next  [1-6] sheets  [q] quit"
	case types.SheetHome:
		if m.achtungFormOpen() {
			help = "[Tab] next field  [Enter] submit  [Esc] cancel  [q] quit"
		} else if m.AchtungViewMenu {
			help = "[d] stop  [Esc] close  [1-6] sheets  [q] quit"
		} else if m.HomeFocusAchtung {
			help = "[tab] VERTEX/UKAZ  [↑/k ↓/j] job  [Enter] details  [t] timer  [a] alarm  [e] every  [D] daily  [m] morning  [d] stop  [1-6] sheets  [q] quit"
		} else if m.HomeFocusUkaz {
			help = "[tab] VERTEX/ACHTUNG  [↑/k ↓/j] UKAZ  [enter] trigger  [1-6] sheets  [q] quit"
		} else {
			help = "[tab] UKAZ/ACHTUNG  [↑/k ↓/j] device  [enter] toggle  [←/h →/l] adjust  [1-6] sheets  [q] quit"
		}
	case types.SheetSystem:
		if m.SystemCommandInput {
			help = ": " + m.SystemCommandBuffer + "▌  [Enter] send  [Esc] cancel"
		} else if m.SystemFocusLogs {
			help = "[Tab] nodes  [:] command  [1-6] sheets  [q] quit"
		} else {
			help = "[Tab] logs  [:] command  [1-6] sheets  [q] quit"
		}
	case types.SheetPeople:
		switch {
		case m.PeopleForm.kind != formNone:
			help = "[Tab] next field  [Enter] submit  [Esc] cancel"
		case m.PeopleConfirm != "":
			help = "[y] remove " + m.PeopleConfirm + "  [any other key] keep"
		case m.signedIn():
			help = "[↑/↓] person  [n] invite  [g] grant  [x] revoke  [K] remove a key  [D] remove  [s] sign in  [e] enrol  [i] invitation  [o] sign out  [r] refresh  [1-6] sheets  [q] quit"
		default:
			help = "[s] sign in  [e] first person  [r] refresh  [1-6] sheets  [q] quit"
		}
	}

	return ui.Help.Render("  " + help)
}
