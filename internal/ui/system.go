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

func (m Model) renderAchtungPanel() string {
	var b strings.Builder
	achtungSelected := m.HomeFocusAchtung
	if achtungSelected {
		b.WriteString(NodeHeaderSelected.Render("▌ACHTUNG  · timers & alarms") + "\n")
	} else {
		b.WriteString(Title.Render("▌ACHTUNG") + " " + Dim.Render("· timers & alarms") + "\n")
	}
	b.WriteString(Dim.Render(strings.Repeat("─", 28)) + "\n\n")

	if m.AchtungTimerDuration != "" {
		b.WriteString(Label.Render("  Name (Enter for auto): "))
		b.WriteString(Value.Render(m.AchtungTimerInput))
		b.WriteString(Dim.Render("▌"))
		b.WriteString("\n")
		b.WriteString(Dim.Render("  [Enter] add  [Esc] cancel"))
		b.WriteString("\n\n")
	} else if m.AchtungTimerCustom {
		b.WriteString(Label.Render("  Duration (e.g. 5m or 2m30s): "))
		b.WriteString(Value.Render(m.AchtungTimerInput))
		b.WriteString(Dim.Render("▌"))
		b.WriteString("\n")
		b.WriteString(Dim.Render("  [Enter] next  [Esc] cancel"))
		b.WriteString("\n\n")
	} else if m.AchtungTimerMenu {
		b.WriteString(Label.Render("  Duration: "))
		b.WriteString(Accent.Render("[1] 1m "))
		b.WriteString(Accent.Render("[2] 5m "))
		b.WriteString(Accent.Render("[3] 10m "))
		b.WriteString(Accent.Render("[4] 30m "))
		b.WriteString(Accent.Render("[5] 1h "))
		b.WriteString(Accent.Render(" [c] custom "))
		b.WriteString(Dim.Render("  [Esc] cancel"))
		b.WriteString("\n\n")
	} else if m.AchtungAlarmStep == 2 {
		b.WriteString(Label.Render("  Alarm name (Enter for auto): "))
		b.WriteString(Value.Render(m.AchtungAlarmInput))
		b.WriteString(Dim.Render("▌"))
		b.WriteString("\n")
		b.WriteString(Dim.Render("  " + m.AchtungAlarmDate + " " + m.AchtungAlarmTime + "  [Enter] add  [Esc] cancel"))
		b.WriteString("\n\n")
	} else if m.AchtungAlarmMenu && m.AchtungAlarmCustom {
		b.WriteString(Label.Render("  Time (HH:MM, past today = tomorrow): "))
		b.WriteString(Value.Render(m.AchtungAlarmInput))
		b.WriteString(Dim.Render("▌"))
		b.WriteString("\n")
		b.WriteString(Dim.Render("  [Enter] next  [Esc] cancel"))
		b.WriteString("\n\n")
	} else if m.AchtungAlarmMenu && m.AchtungAlarmStep == 0 {
		b.WriteString(Label.Render("  Type: "))
		b.WriteString(Accent.Render("[1] One-shot "))
		b.WriteString(Dim.Render("(date & time)  [Esc] cancel"))
		b.WriteString("\n\n")
	} else if m.AchtungAlarmMenu && m.AchtungAlarmStep == 1 {
		b.WriteString(Label.Render("  When: "))
		b.WriteString(Accent.Render("[1] today 20:00 "))
		b.WriteString(Accent.Render("[2] tomorrow 08:00 "))
		b.WriteString(Accent.Render(" [c] custom "))
		b.WriteString(Dim.Render("  [Esc] cancel"))
		b.WriteString("\n\n")
	}

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
