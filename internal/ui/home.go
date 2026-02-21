package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

func (m Model) renderHome(showAchtungFormInline bool) string {
	boxWidth := 50
	uniformHeight := 8

	// Two uniform boxes stacked vertically (1 row spacing inside each)
	vertexContent := padToLinesWithSpacing(m.renderVertexDevicesContent(), uniformHeight)
	achtungContent := padToLinesWithSpacing(m.renderAchtungContent(showAchtungFormInline), uniformHeight)

	vertexBox := NewBox(boxWidth).WithTitle("VERTEX  devices").WithDimTitle(m.HomeFocusAchtung)
	achtungBox := NewBox(boxWidth).WithTitle("ACHTUNG  timers & alarms").WithDimTitle(!m.HomeFocusAchtung)

	vertexSection := vertexBox.Render(vertexContent)
	achtungSection := achtungBox.Render(achtungContent)

	content := lipgloss.JoinVertical(lipgloss.Left, vertexSection, "", achtungSection)
	return indentLines(content, "  ")
}

func padToLines(s string, n int) string {
	lines := strings.Split(s, "\n")
	for len(lines) < n {
		lines = append(lines, "")
	}
	if len(lines) > n {
		lines = lines[:n]
	}
	return strings.Join(lines, "\n")
}

// padToLinesWithSpacing pads content to n lines with 1 row spacing at top and bottom (for VERTEX/ACHTUNG boxes).
func padToLinesWithSpacing(s string, n int) string {
	lines := strings.Split(s, "\n")
	// Reserve first and last for spacing
	inner := n - 2
	if inner < 1 {
		inner = 1
	}
	for len(lines) < inner {
		lines = append(lines, "")
	}
	if len(lines) > inner {
		lines = lines[:inner]
	}
	return "\n" + strings.Join(lines, "\n") + "\n"
}

func (m Model) renderVertexDevicesContent() string {
	nodeOrder := m.deviceNodes()
	nodes := make(map[string][]int)
	for i, d := range m.HomeDevices {
		nodes[d.Node] = append(nodes[d.Node], i)
	}

	var lines []string
	for _, node := range nodeOrder {
		if indices, ok := nodes[node]; ok {
			for _, i := range indices {
				d := m.HomeDevices[i]
				line := m.renderDeviceLine(d, i == m.SelectedDevice)
				lines = append(lines, line)
			}
		}
	}

	if len(lines) == 0 {
		return Dim.Render("  No devices")
	}
	return strings.Join(lines, "\n")
}

func (m Model) renderAchtungContent(showFormInline bool) string {
	if len(m.AchtungJobs) == 0 {
		return Dim.Render("  No timers or alarms.\n  [t] New timer  [a] New alarm")
	}
	var lines []string
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
		lines = append(lines, line)
	}
	lines = append(lines, "")
	lines = append(lines, Dim.Render("  [t] timer  [a] alarm  [d] delete"))
	return strings.Join(lines, "\n")
}

func (m Model) deviceNodes() []string {
	seen := map[string]bool{}
	var order []string
	for _, d := range m.HomeDevices {
		if !seen[d.Node] {
			seen[d.Node] = true
			order = append(order, d.Node)
		}
	}
	return order
}

func (m Model) renderDeviceLine(d HomeDevice, selected bool) string {
	w := 46
	prefix := "  "
	if selected {
		prefix = "▌ "
	}
	switch d.Kind {
	case "toggle":
		return prefix + m.renderToggleDevice(d, selected, w-2)
	case "cycle":
		return prefix + m.renderCycleDevice(d, selected, w-2)
	case "value":
		s := m.renderValueDevice(d, selected, w-2)
		parts := strings.SplitN(s, "\n", 2)
		if len(parts) == 2 {
			return prefix + parts[0] + "\n  " + parts[1]
		}
		return prefix + s
	default:
		line := fmt.Sprintf("%s %s", Label.Render(d.Name), d.Status)
		if selected {
			return prefix + Selected.Render(PadLine(line, w-2))
		}
		return prefix + PadLine(line, w-2)
	}
}

func (m Model) renderToggleDevice(d HomeDevice, selected bool, w int) string {
	icon := getDeviceIcon(d.Status)
	line := fmt.Sprintf(" %s %-16s %s", icon, d.Name, Label.Render("["+d.Topic+"]"))
	if selected {
		return Selected.Render(PadLine(line, w))
	}
	return PadLine(line, w)
}

func (m Model) renderCycleDevice(d HomeDevice, selected bool, w int) string {
	icon := getCycleIcon(d.Status)
	mode := strings.ToUpper(d.Status)
	modeStyled := cycleStatusStyle(d.Status).Render(mode)
	line := fmt.Sprintf(" %s %-12s %s", icon, d.Name, modeStyled)
	if selected {
		return Selected.Render(PadLine(line, w))
	}
	return PadLine(line, w)
}

func (m Model) renderValueDevice(d HomeDevice, selected bool, w int) string {
	pct := float64(d.Val-d.Min) / float64(d.Max-d.Min) * 100
	barWidth := w - 18
	if barWidth < 6 {
		barWidth = 6
	}
	bar := RenderBar(pct, barWidth)
	valStr := fmt.Sprintf("%3d", d.Val)

	line1 := fmt.Sprintf(" ◈ %-12s %s", d.Name, Label.Render(d.Property))
	line2 := fmt.Sprintf("   %s %s", bar, Value.Render(valStr))

	if selected {
		return Selected.Render(PadLine(line1, w)) + "\n" + Selected.Render(PadLine(line2, w))
	}
	return PadLine(line1, w) + "\n" + PadLine(line2, w)
}

func getDeviceIcon(status string) string {
	switch status {
	case "on":
		return Online.Render("●")
	case "off":
		return Offline.Render("○")
	case "unknown":
		return Warning.Render("?")
	default:
		return Label.Render("·")
	}
}

func getCycleIcon(status string) string {
	switch status {
	case "off":
		return Offline.Render("○")
	case "blink":
		return lipgloss.NewStyle().Foreground(GruvYellow).Render("◎")
	case "fade":
		return lipgloss.NewStyle().Foreground(GruvPurple).Render("◉")
	case "solid":
		return Online.Render("●")
	default:
		return Label.Render("·")
	}
}

func cycleStatusStyle(status string) lipgloss.Style {
	switch status {
	case "off":
		return Offline
	case "blink":
		return lipgloss.NewStyle().Foreground(GruvYellow).Bold(true)
	case "fade":
		return lipgloss.NewStyle().Foreground(GruvPurple).Bold(true)
	case "solid":
		return Online
	default:
		return Label
	}
}
