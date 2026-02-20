package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

func (m Model) renderDiary() string {
	var b strings.Builder

	b.WriteString(Title.Render("▌DIARY ENTRIES") + "\n\n")

	width := 52

	var entries []string
	for i, e := range m.DiaryEntries {
		dateStr := e.Date.Format("02 Jan")
		moodIcon := getMoodIcon(e.Mood)
		preview := e.Content
		if len(preview) > 38 {
			preview = preview[:38] + "..."
		}

		var lines []string
		line1 := fmt.Sprintf(" %s  %s  %s",
			Label.Render(dateStr),
			moodIcon,
			Label.Render(e.Mood))
		lines = append(lines, PadLine(line1, width-2))
		lines = append(lines, PadLine(" "+Value.Render(preview), width-2))

		content := strings.Join(lines, "\n")

		box := NewBox(width)
		if i == m.SelectedEntry {
			box = box.WithBorderColor(GruvYellow)
		}

		entries = append(entries, box.Render(content))
	}

	b.WriteString(lipgloss.JoinVertical(lipgloss.Left, entries...))

	// Preview panel
	if m.SelectedEntry < len(m.DiaryEntries) {
		b.WriteString("\n\n")
		selected := m.DiaryEntries[m.SelectedEntry]

		previewWidth := 62
		var lines []string
		lines = append(lines, PadLine(" "+Title.Render(selected.Date.Format("Monday, 02 January 2006")), previewWidth-2))
		lines = append(lines, "")
		lines = append(lines, PadLine(" "+Value.Render(selected.Content), previewWidth-2))

		content := strings.Join(lines, "\n")
		preview := NewBox(previewWidth).Render(content)
		b.WriteString(preview)
	}

	return indentLines(b.String(), "  ")
}

func getMoodIcon(mood string) string {
	switch mood {
	case "focused":
		return lipgloss.NewStyle().Foreground(GruvBlue).Render("◆")
	case "productive":
		return lipgloss.NewStyle().Foreground(GruvGreen).Render("◆")
	case "calm":
		return lipgloss.NewStyle().Foreground(GruvAqua).Render("◆")
	case "tired":
		return lipgloss.NewStyle().Foreground(GruvYellow).Render("◆")
	case "stressed":
		return lipgloss.NewStyle().Foreground(GruvRed).Render("◆")
	default:
		return lipgloss.NewStyle().Foreground(GruvGray).Render("◆")
	}
}
