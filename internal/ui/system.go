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

	overhead := 5 + 2 + 2 + strings.Count(nodeRow, "\n") + 1 + 3 + 1 + 2
	maxLogs := m.Height - overhead
	if maxLogs < 3 {
		maxLogs = 3
	}

	logs := m.Logs
	if len(logs) > maxLogs {
		logs = logs[:maxLogs]
	}

	var logLines []string
	for _, l := range logs {
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
	width := 24

	var statusLine string
	switch n.Status {
	case "online":
		statusLine = Online.Render("● ONLINE")
	case "offline":
		statusLine = Offline.Render("● OFFLINE")
	default:
		statusLine = Warning.Render("● UNKNOWN")
	}

	var pingStr string
	if n.Status == "online" && n.PingMs > 0 {
		pingStr = fmt.Sprintf("%dms", n.PingMs)
	} else {
		pingStr = "—"
	}

	var lines []string
	lines = append(lines, PadLine(" "+Title.Render(n.Name), width-2))
	lines = append(lines, PadLine(" "+statusLine, width-2))
	lines = append(lines, "")
	lines = append(lines, PadLine(fmt.Sprintf(" %s %s", Label.Render("PING:"), Value.Render(pingStr)), width-2))
	lines = append(lines, PadLine(fmt.Sprintf(" %s %s", Label.Render("UP:  "), Value.Render(n.Uptime)), width-2))

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
