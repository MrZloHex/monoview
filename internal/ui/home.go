package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

func (m Model) renderHome() string {
	var b strings.Builder

	b.WriteString(Title.Render("▌HOME AUTOMATION") + "\n\n")

	nodeOrder := m.deviceNodes()
	nodes := make(map[string][]int)
	for i, d := range m.HomeDevices {
		nodes[d.Node] = append(nodes[d.Node], i)
	}

	var panels []string
	for _, node := range nodeOrder {
		if indices, ok := nodes[node]; ok {
			panel := m.renderNodeDevicePanel(node, indices)
			panels = append(panels, panel)
		}
	}

	var rows []string
	for i := 0; i < len(panels); i += 3 {
		end := i + 3
		if end > len(panels) {
			end = len(panels)
		}
		row := lipgloss.JoinHorizontal(lipgloss.Top, panels[i:end]...)
		rows = append(rows, row)
	}

	content := lipgloss.JoinVertical(lipgloss.Left, rows...)
	b.WriteString(content)

	return indentLines(b.String(), "  ")
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

func (m Model) renderNodeDevicePanel(node string, indices []int) string {
	width := 34

	var lines []string

	lines = append(lines, PadLine(" "+Accent.Render(node), width-2))
	lines = append(lines, PadLine(" "+Dim.Render(strings.Repeat("─", width-4)), width-2))

	for _, i := range indices {
		d := m.HomeDevices[i]
		selected := i == m.SelectedDevice

		switch d.Kind {
		case "toggle":
			lines = append(lines, m.renderToggleDevice(d, selected, width-2))
		case "cycle":
			lines = append(lines, m.renderCycleDevice(d, selected, width-2))
		case "value":
			lines = append(lines, m.renderValueDevice(d, selected, width-2))
		}
	}

	content := strings.Join(lines, "\n")
	return NewBox(width).Render(content) + " "
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
