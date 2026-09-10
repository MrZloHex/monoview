package app

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"monoview/internal/types"
	"monoview/internal/ui"
)

func (m Model) renderHome(showAchtungFormInline bool) string {
	const boxWidth = 50
	const achtungWidth = 62 // wider: a job row carries kind, name, countdown and due
	const leftBoxHeight = 8

	// Two columns: the VERTEX and UKAZ boxes stack on the left, and
	// ACHTUNG sits top-right spanning both of them. Jobs are the thing
	// most worth seeing at a glance -- timers, alarms and the morning
	// print -- and stacked third at eight lines it could show about four.
	vertexContent := padToLinesWithSpacing(m.renderVertexDevicesContent(), leftBoxHeight)
	ukazContent := padToLinesWithSpacing(m.renderUkazDevicesContent(), leftBoxHeight)

	focusVertex := !m.HomeFocusAchtung && !m.HomeFocusUkaz
	focusUkaz := !m.HomeFocusAchtung && m.HomeFocusUkaz

	vertexBox := ui.NewBox(boxWidth).WithTitle("VERTEX  devices").WithDimTitle(!focusVertex)
	ukazBox := ui.NewBox(boxWidth).WithTitle("UKAZ  print").WithDimTitle(!focusUkaz)

	left := lipgloss.JoinVertical(lipgloss.Left,
		vertexBox.Render(vertexContent),
		"",
		ukazBox.Render(ukazContent),
	)

	// Match the left column's total height: two boxes of leftBoxHeight
	// content plus their borders, and the blank row between them, less
	// this box's own two border rows.
	achtungInner := (leftBoxHeight+2)*2 + 1 - 2
	achtungContent := padToLinesWithSpacing(
		m.renderAchtungContent(showAchtungFormInline), achtungInner)

	achtungBox := ui.NewBox(achtungWidth).WithTitle("ACHTUNG  jobs").WithDimTitle(!m.HomeFocusAchtung)

	// The right column shows the jobs, or -- when one is open -- the form
	// or job detail in their place. They used to render as a third column
	// further right, which at any ordinary terminal width squeezed the
	// jobs box down to a truncated sliver.
	right := achtungBox.Render(achtungContent)
	switch {
	case m.achtungFormOpen():
		right = m.renderAchtungFormBox(achtungInner + 2)
	case m.AchtungViewMenu && m.SelectedAchtungJob < len(m.AchtungJobs):
		right = m.renderAchtungJobDetailView(
			m.AchtungJobs[m.SelectedAchtungJob], achtungInner+2)
	}

	content := lipgloss.JoinHorizontal(lipgloss.Top, left, "  ", right)
	return ui.IndentLines(content, "  ")
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
	return m.renderDevicesForNode("VERTEX")
}

func (m Model) renderUkazDevicesContent() string {
	return m.renderDevicesForNode("UKAZ")
}

func (m Model) renderDevicesForNode(node string) string {
	var lines []string
	for i, d := range m.HomeDevices {
		if strings.ToUpper(d.Node) != node {
			continue
		}
		line := m.renderDeviceLine(d, i == m.SelectedDevice)
		lines = append(lines, line)
	}
	if len(lines) == 0 {
		return ui.Dim.Render("  No devices")
	}
	return strings.Join(lines, "\n")
}

func (m Model) renderAchtungContent(showFormInline bool) string {
	if len(m.AchtungJobs) == 0 {
		lines := []string{ui.Dim.Render("  No jobs."), ""}
		lines = append(lines, achtungKeyHints()...)
		return strings.Join(lines, "\n")
	}
	var lines []string
	for i, j := range m.AchtungJobs {
		active := i == m.SelectedAchtungJob
		kindStyle := ui.Label
		if j.Kind == "ALARM" {
			kindStyle = ui.Accent
		}
		line := fmt.Sprintf("%s %s  %s  %s",
			kindStyle.Render(j.Kind+":"),
			ui.Value.Render(j.Name),
			ui.Label.Render("left:"),
			ui.Value.Render(j.Remaining))
		if j.Due != "" && j.Due != "—" {
			line += "  " + ui.Label.Render("due:") + " " + ui.Value.Render(j.Due)
		}
		if active {
			line = "▌ " + line
		} else {
			line = "  " + line
		}
		lines = append(lines, line)
	}
	lines = append(lines, "")
	lines = append(lines, achtungKeyHints()...)
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

func (m Model) renderDeviceLine(d types.HomeDevice, selected bool) string {
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
	case "action":
		return prefix + m.renderActionDevice(d, selected, w-2)
	default:
		line := fmt.Sprintf("%s %s", ui.Label.Render(d.Name), d.Status)
		if selected {
			return prefix + ui.Selected.Render(ui.PadLine(line, w-2))
		}
		return prefix + ui.PadLine(line, w-2)
	}
}

func (m Model) renderActionDevice(d types.HomeDevice, selected bool, w int) string {
	icon := ui.Label.Render("▶")
	line := fmt.Sprintf(" %s %-20s %s", icon, d.Name, ui.Dim.Render("[Enter] trigger"))
	if selected {
		return ui.Selected.Render(ui.PadLine(line, w))
	}
	return ui.PadLine(line, w)
}

func (m Model) renderToggleDevice(d types.HomeDevice, selected bool, w int) string {
	icon := getDeviceIcon(d.Status)
	line := fmt.Sprintf(" %s %-16s %s", icon, d.Name, ui.Label.Render("["+d.Topic+"]"))
	if selected {
		return ui.Selected.Render(ui.PadLine(line, w))
	}
	return ui.PadLine(line, w)
}

func (m Model) renderCycleDevice(d types.HomeDevice, selected bool, w int) string {
	icon := getCycleIcon(d.Status)
	mode := strings.ToUpper(d.Status)
	modeStyled := cycleStatusStyle(d.Status).Render(mode)
	line := fmt.Sprintf(" %s %-12s %s", icon, d.Name, modeStyled)
	if selected {
		return ui.Selected.Render(ui.PadLine(line, w))
	}
	return ui.PadLine(line, w)
}

func (m Model) renderValueDevice(d types.HomeDevice, selected bool, w int) string {
	pct := float64(d.Val-d.Min) / float64(d.Max-d.Min) * 100
	barWidth := w - 18
	if barWidth < 6 {
		barWidth = 6
	}
	bar := ui.RenderBar(pct, barWidth)
	valStr := fmt.Sprintf("%3d", d.Val)

	line1 := fmt.Sprintf(" ◈ %-12s %s", d.Name, ui.Label.Render(d.Property))
	line2 := fmt.Sprintf("   %s %s", bar, ui.Value.Render(valStr))

	if selected {
		return ui.Selected.Render(ui.PadLine(line1, w)) + "\n" + ui.Selected.Render(ui.PadLine(line2, w))
	}
	return ui.PadLine(line1, w) + "\n" + ui.PadLine(line2, w)
}

func getDeviceIcon(status string) string {
	switch status {
	case "on":
		return ui.Online.Render("●")
	case "off":
		return ui.Offline.Render("○")
	case "unknown":
		return ui.Warning.Render("?")
	default:
		return ui.Label.Render("·")
	}
}

func getCycleIcon(status string) string {
	switch status {
	case "off":
		return ui.Offline.Render("○")
	case "blink":
		return lipgloss.NewStyle().Foreground(ui.GruvYellow).Render("◎")
	case "fade":
		return lipgloss.NewStyle().Foreground(ui.GruvPurple).Render("◉")
	case "solid":
		return ui.Online.Render("●")
	default:
		return ui.Label.Render("·")
	}
}

func cycleStatusStyle(status string) lipgloss.Style {
	switch status {
	case "off":
		return ui.Offline
	case "blink":
		return lipgloss.NewStyle().Foreground(ui.GruvYellow).Bold(true)
	case "fade":
		return lipgloss.NewStyle().Foreground(ui.GruvPurple).Bold(true)
	case "solid":
		return ui.Online
	default:
		return ui.Label
	}
}

// achtungKeyHints splits the key list over two rows. As one line it ran
// past the box and got clipped mid-word, which hid [m] -- the one that
// sets the morning print.
func achtungKeyHints() []string {
	return []string{
		ui.Dim.Render("  [t] timer   [a] alarm   [e] every   [D] daily"),
		ui.Dim.Render("  [m] morning agenda print          [d] delete"),
	}
}
