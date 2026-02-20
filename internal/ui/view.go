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
		b.WriteString(m.renderCalendar())
	case SheetDiary:
		b.WriteString(m.renderDiary())
	case SheetHome:
		b.WriteString(m.renderHome())
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

	if m.FireAlert.Show {
		fullView = m.overlayFireAlert(fullView)
	}

	return fullView
}

// overlayFireAlert draws the fire-alert popup centered over the view.
func (m Model) overlayFireAlert(fullView string) string {
	lines := strings.Split(fullView, "\n")
	// Ensure we have exactly Height lines
	for len(lines) < m.Height {
		lines = append(lines, "")
	}
	if len(lines) > m.Height {
		lines = lines[:m.Height]
	}

	popupContent := m.renderFireAlertPopup()
	popupLines := strings.Split(popupContent, "\n")
	popupHeight := len(popupLines)
	startRow := (m.Height - popupHeight) / 2
	if startRow < 0 {
		startRow = 0
	}

	for i, pl := range popupLines {
		row := startRow + i
		if row >= len(lines) {
			break
		}
		w := lipgloss.Width(pl)
		padLeft := (m.Width - w) / 2
		if padLeft < 0 {
			padLeft = 0
		}
		lines[row] = strings.Repeat(" ", padLeft) + pl
	}

	return strings.Join(lines, "\n")
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
		help = "[←/h] prev day  [→/l] next day  [1-4] sheets  [q] quit"
	case SheetDiary:
		help = "[↑/k] prev  [↓/j] next  [1-4] sheets  [q] quit"
	case SheetHome:
		if m.HomeFocusAchtung {
			help = "[tab] devices  [↑/k ↓/j] job  [t] timer  [a] alarm  [d] delete  [1-4] sheets  [q] quit"
		} else {
			help = "[tab] timers  [↑/k ↓/j] device  [enter] toggle  [←/h →/l] adjust  [1-4] sheets  [q] quit"
		}
	case SheetSystem:
		help = "[↑/k ↓/j] or [←/h →/l] select node  [enter] ping  [1-4] sheets  [q] quit"
	}

	return Help.Render("  " + help)
}
