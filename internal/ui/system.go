package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

func (m Model) renderSystem() string {
	// Left: nodes in 2 columns (vertical layout). Right: logs.
	// Split nodes into two columns
	mid := (len(m.Nodes) + 1) / 2
	col1Nodes := m.Nodes[:mid]
	col2Nodes := m.Nodes[mid:]

	var col1Panels, col2Panels []string
	for i, n := range col1Nodes {
		panel := m.renderNodePanel(n, i == m.SelectedNode)
		col1Panels = append(col1Panels, panel)
	}
	for i, n := range col2Nodes {
		panel := m.renderNodePanel(n, mid+i == m.SelectedNode)
		col2Panels = append(col2Panels, panel)
	}

	col1 := lipgloss.JoinVertical(lipgloss.Left, col1Panels...)
	col2 := lipgloss.JoinVertical(lipgloss.Left, col2Panels...)
	nodesBlock := lipgloss.JoinHorizontal(lipgloss.Top, col1, "  ", col2)

	// Left: nodes. Selected = color, unselected = dim (Tab cycles nodes ↔ logs).
	nodesHeader := Dim.Render("▌NODES") + "\n\n"
	if !m.SystemFocusLogs {
		nodesHeader = Title.Render("▌NODES") + " " + Dim.Render("[←↑↓→] grid nodes  [Enter] ping") + "\n\n"
	}
	nodesSection := nodesHeader + nodesBlock

	// Right: logs. Selected = color, unselected = dim.
	logsHeaderLines := 3
	visibleLogLines := m.plainHeight() - logsHeaderLines
	if visibleLogLines < 5 {
		visibleLogLines = 5
	}
	logsHeader := Title.Render("▌RECENT LOGS") + "\n\n"
	if m.SystemFocusLogs {
		logsHeader = Title.Render("▌RECENT LOGS") + " " + Dim.Render("[↑↓] scroll") + "\n\n"
	} else {
		logsHeader = Dim.Render("▌RECENT LOGS") + "\n\n"
	}
	logs := m.Logs
	if len(logs) > 50 {
		logs = logs[:50]
	}
	// Viewport: slice by scroll offset
	offset := m.LogScrollOffset
	if offset < 0 {
		offset = 0
	}
	if offset >= len(logs) && len(logs) > 0 {
		offset = len(logs) - 1
	}
	end := offset + visibleLogLines
	if end > len(logs) {
		end = len(logs)
	}
	visibleLogs := logs[offset:end]
	var logLines []string
	logWidth := 50
	if m.Width > 0 {
		logWidth = m.Width/2 - 10
		if logWidth < 30 {
			logWidth = 30
		}
	}
	for _, l := range visibleLogs {
		timeStr := l.Time.Format("15:04:05")
		level := getLogLevelStyle(l.Level)
		source := Accent.Render(fmt.Sprintf("%-8s", l.Source))
		msg := l.Message
		if len(msg) > logWidth {
			msg = msg[:logWidth] + "..."
		}
		line := fmt.Sprintf("%s %s %s %s",
			Label.Render(timeStr),
			level,
			source,
			Value.Render(msg))
		logLines = append(logLines, line)
	}
	logsSection := logsHeader + strings.Join(logLines, "\n")

	// Join left (nodes) and right (logs)
	content := lipgloss.JoinHorizontal(lipgloss.Top, nodesSection, "    ", logsSection)
	return indentLines(content, "  ")
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
