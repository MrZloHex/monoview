package app

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"monoview/internal/types"
	"monoview/internal/ui"
)

func (m Model) renderSystem() string {
	// Left: nodes in 2 columns (vertical layout). Right: logs.
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

	nodesHeader := ui.Dim.Render("▌NODES") + "\n\n"
	if !m.SystemFocusLogs {
		nodesHeader = ui.Title.Render("▌NODES") + " " + ui.Dim.Render("[←↑↓→] grid nodes  [Enter] ping") + "\n\n"
	}
	nodesSection := nodesHeader + nodesBlock

	logsHeaderLines := 3
	visibleLogLines := m.plainHeight() - logsHeaderLines
	if visibleLogLines < 5 {
		visibleLogLines = 5
	}
	logsHeader := ui.Title.Render("▌RECENT LOGS") + "\n\n"
	if m.SystemFocusLogs {
		logsHeader = ui.Title.Render("▌RECENT LOGS") + " " + ui.Dim.Render("[↑↓] scroll") + "\n\n"
	} else {
		logsHeader = ui.Dim.Render("▌RECENT LOGS") + "\n\n"
	}
	logs := m.Logs
	if len(logs) > 50 {
		logs = logs[:50]
	}
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
		source := ui.Accent.Render(fmt.Sprintf("%-8s", l.Source))
		msg := l.Message
		if len(msg) > logWidth {
			msg = msg[:logWidth] + "..."
		}
		line := fmt.Sprintf("%s %s %s %s",
			ui.Label.Render(timeStr),
			level,
			source,
			ui.Value.Render(msg))
		logLines = append(logLines, line)
	}
	logsSection := logsHeader + strings.Join(logLines, "\n")

	content := lipgloss.JoinHorizontal(lipgloss.Top, nodesSection, "    ", logsSection)
	return ui.IndentLines(content, "  ")
}

func (m Model) renderNodePanel(n types.SystemNode, active bool) string {
	width := 24

	var statusLine string
	switch n.Status {
	case "online":
		statusLine = ui.Online.Render("● ONLINE")
	case "offline":
		statusLine = ui.Offline.Render("● OFFLINE")
	default:
		statusLine = ui.Warning.Render("● UNKNOWN")
	}

	var pingStr string
	if n.Status == "online" && n.PingMs > 0 {
		pingStr = fmt.Sprintf("%dms", n.PingMs)
	} else {
		pingStr = "—"
	}

	var lines []string
	lines = append(lines, ui.PadLine(" "+ui.Title.Render(n.Name), width-2))
	lines = append(lines, ui.PadLine(" "+statusLine, width-2))
	lines = append(lines, "")
	lines = append(lines, ui.PadLine(fmt.Sprintf(" %s %s", ui.Label.Render("PING:"), ui.Value.Render(pingStr)), width-2))
	lines = append(lines, ui.PadLine(fmt.Sprintf(" %s %s", ui.Label.Render("UP:  "), ui.Value.Render(n.Uptime)), width-2))

	content := strings.Join(lines, "\n")

	box := ui.NewBox(width)
	if active {
		box = box.WithBorderColor(ui.GruvYellow)
	}

	return box.Render(content) + " "
}

func getLogLevelStyle(level string) string {
	switch level {
	case "INFO":
		return lipgloss.NewStyle().Foreground(ui.GruvBlue).Width(5).Render(level)
	case "WARN":
		return lipgloss.NewStyle().Foreground(ui.GruvYellow).Width(5).Render(level)
	case "ERR":
		return lipgloss.NewStyle().Foreground(ui.GruvRed).Width(5).Render(level)
	default:
		return lipgloss.NewStyle().Foreground(ui.GruvGray).Width(5).Render(level)
	}
}
