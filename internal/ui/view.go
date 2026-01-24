package ui

import (
	"fmt"
	"strings"

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

	b.WriteString("\n")
	b.WriteString(m.renderFooter())

	return b.String()
}

func (m Model) renderHeader() string {
	logo := `
  ╔╦╗╔═╗╔╗╔╔═╗╦  ╦╦╔═╗╦ ╦
  ║║║║ ║║║║║ ║╚╗╔╝║║╣ ║║║
  ╩ ╩╚═╝╝╚╝╚═╝ ╚╝ ╩╚═╝╚╩╝`

	logoStyled := lipgloss.NewStyle().Foreground(GruvOrange).Render(logo)

	clock := m.LastUpdate.Format("15:04:05")
	date := m.LastUpdate.Format("Mon, 02 Jan 2006")

	var timeLines []string
	timeLines = append(timeLines, PadLine(" "+Value.Render(clock)+" ", 20))
	timeLines = append(timeLines, PadLine(" "+Label.Render(date)+" ", 20))

	timeBox := NewBox(22).Render(strings.Join(timeLines, "\n"))

	gap := m.Width - lipgloss.Width(logo) - lipgloss.Width(timeBox) - 4
	if gap < 0 {
		gap = 2
	}

	return lipgloss.JoinHorizontal(
		lipgloss.Top,
		logoStyled,
		strings.Repeat(" ", gap),
		timeBox,
	)
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
		help = "[↑/k] prev  [↓/j] next  [enter] toggle  [1-4] sheets  [q] quit"
	case SheetSystem:
		help = "[↑/k] prev  [↓/j] next  [1-4] sheets  [q] quit"
	}

	return Help.Render("  " + help)
}
