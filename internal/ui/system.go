package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

func (m Model) renderSystem() string {
	var b strings.Builder

	b.WriteString(Title.Render("▌SYSTEM NODES") + "\n\n")

	var nodePanels []string
	for i, n := range m.Nodes {
		panel := m.renderNodePanel(n, i == m.SelectedNode)
		nodePanels = append(nodePanels, panel)
	}

	nodeRow := lipgloss.JoinHorizontal(lipgloss.Top, nodePanels...)
	b.WriteString(nodeRow)
	b.WriteString("\n\n")

	b.WriteString(Title.Render("▌RECENT LOGS") + "\n\n")

	var logLines []string
	for _, l := range m.Logs {
		timeStr := l.Time.Format("15:04:05")
		level := getLogLevelStyle(l.Level)
		source := Accent.Render(fmt.Sprintf("%-8s", l.Source))
		msg := l.Message
		if len(msg) > 50 {
			msg = msg[:50] + "..."
		}

		line := fmt.Sprintf("%s %s %s %s",
			Label.Render(timeStr),
			level,
			source,
			Value.Render(msg))
		logLines = append(logLines, line)
	}

	b.WriteString(strings.Join(logLines, "\n"))

	return indentLines(b.String(), "  ")
}

func (m Model) renderNodePanel(n SystemNode, active bool) string {
	width := 22

	status := Online.Render("● ONLINE ")
	if n.Status == "offline" {
		status = Offline.Render("● OFFLINE")
	}

	cpu := RenderBar(n.CPU, 10)
	mem := RenderBar(n.Memory, 10)

	var lines []string
	lines = append(lines, PadLine(" "+Title.Render(strings.ToUpper(n.Name)), width-2))
	lines = append(lines, PadLine(" "+status, width-2))
	lines = append(lines, "")
	lines = append(lines, PadLine(" "+Label.Render("CPU"), width-2))
	lines = append(lines, PadLine(fmt.Sprintf(" %s %s", cpu, Label.Render(fmt.Sprintf("%5.1f%%", n.CPU))), width-2))
	lines = append(lines, PadLine(" "+Label.Render("MEM"), width-2))
	lines = append(lines, PadLine(fmt.Sprintf(" %s %s", mem, Label.Render(fmt.Sprintf("%5.1f%%", n.Memory))), width-2))
	lines = append(lines, "")
	lines = append(lines, PadLine(fmt.Sprintf(" %s %s", Label.Render("UP:"), Value.Render(n.Uptime)), width-2))

	content := strings.Join(lines, "\n")

	box := NewBox(width)
	if active {
		box = box.WithBorderColor(GruvYellow)
	}

	return box.Render(content) + " "
}

func getLogLevelStyle(level string) string {
	switch level {
	case "INFO":
		return lipgloss.NewStyle().Foreground(GruvBlue).Width(5).Render(level)
	case "WARN":
		return lipgloss.NewStyle().Foreground(GruvYellow).Width(5).Render(level)
	case "ERR":
		return lipgloss.NewStyle().Foreground(GruvRed).Width(5).Render(level)
	default:
		return lipgloss.NewStyle().Foreground(GruvGray).Width(5).Render(level)
	}
}
