package app

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"monoview/internal/ui"
)

func (m Model) renderDiary() string {
	var b strings.Builder

	b.WriteString(ui.Title.Render("▌DIARY ENTRIES") + "\n\n")

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
			ui.Label.Render(dateStr),
			moodIcon,
			ui.Label.Render(e.Mood))
		lines = append(lines, ui.PadLine(line1, width-2))
		lines = append(lines, ui.PadLine(" "+ui.Value.Render(preview), width-2))

		content := strings.Join(lines, "\n")

		box := ui.NewBox(width)
		if i == m.SelectedEntry {
			box = box.WithBorderColor(ui.GruvYellow)
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
		lines = append(lines, ui.PadLine(" "+ui.Title.Render(selected.Date.Format("Monday, 02 January 2006")), previewWidth-2))
		lines = append(lines, "")
		lines = append(lines, ui.PadLine(" "+ui.Value.Render(selected.Content), previewWidth-2))

		content := strings.Join(lines, "\n")
		preview := ui.NewBox(previewWidth).Render(content)
		b.WriteString(preview)
	}

	return ui.IndentLines(b.String(), "  ")
}

func getMoodIcon(mood string) string {
	switch mood {
	case "focused":
		return lipgloss.NewStyle().Foreground(ui.GruvBlue).Render("◆")
	case "productive":
		return lipgloss.NewStyle().Foreground(ui.GruvGreen).Render("◆")
	case "calm":
		return lipgloss.NewStyle().Foreground(ui.GruvAqua).Render("◆")
	case "tired":
		return lipgloss.NewStyle().Foreground(ui.GruvYellow).Render("◆")
	case "stressed":
		return lipgloss.NewStyle().Foreground(ui.GruvRed).Render("◆")
	default:
		return lipgloss.NewStyle().Foreground(ui.GruvGray).Render("◆")
	}
}
