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

	nodesHeader := Title.Render("▌NODES") + "\n\n"
	nodesSection := nodesHeader + nodesBlock

	// Right: logs
	logsHeader := Title.Render("▌RECENT LOGS") + "\n\n"
	logs := m.Logs
	maxLogs := 50
	if len(logs) > maxLogs {
		logs = logs[:maxLogs]
	}
	var logLines []string
	logWidth := 50
	if m.Width > 0 {
		logWidth = m.Width/2 - 10
		if logWidth < 30 {
			logWidth = 30
		}
	}
	for _, l := range logs {
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


func (m Model) renderAchtungPanel(showFormInline bool) string {
	var b strings.Builder
	achtungSelected := m.HomeFocusAchtung
	if achtungSelected {
		b.WriteString(NodeHeaderSelected.Render("▌ACHTUNG  · timers & alarms") + "\n")
	} else {
		b.WriteString(Title.Render("▌ACHTUNG") + " " + Dim.Render("· timers & alarms") + "\n")
	}
	b.WriteString(Dim.Render(strings.Repeat("─", 28)) + "\n\n")

	// Form is shown in right panel when AchtungTimerMenu or AchtungAlarmMenu; no inline form here.

	if len(m.AchtungJobs) == 0 {
		b.WriteString(Dim.Render("  No timers or alarms. [t] New timer  [a] New alarm"))
		b.WriteString("\n")
	} else {
		for i, j := range m.AchtungJobs {
			active := i == m.SelectedAchtungJob
			kindStyle := Label
			if j.Kind == "ALARM" {
				kindStyle = Accent
			}
			line := fmt.Sprintf("%s %s  %s  %s",
				kindStyle.Render(j.Kind+":"),
				Value.Render(j.Name),
				Label.Render("left:"),
				Value.Render(j.Remaining))
			if j.Due != "" && j.Due != "—" {
				line += "  " + Label.Render("due:") + " " + Value.Render(j.Due)
			}
			if active {
				line = "▌ " + line
			} else {
				line = "  " + line
			}
			b.WriteString(line + "\n")
		}
		b.WriteString(Dim.Render("\n  [t] New timer  [a] New alarm  [d] Delete selected"))
		b.WriteString("\n")
	}

	return b.String()
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
