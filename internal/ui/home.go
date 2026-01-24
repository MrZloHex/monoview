package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

func (m Model) renderHome() string {
	var b strings.Builder

	b.WriteString(Title.Render("▌HOME AUTOMATION") + "\n\n")

	// Group by room (maintain order)
	roomOrder := []string{"Living Room", "Hallway", "Entrance", "Bedroom", "Kitchen", "Garage"}
	rooms := make(map[string][]int)
	for i, d := range m.HomeDevices {
		rooms[d.Room] = append(rooms[d.Room], i)
	}

	var panels []string
	for _, room := range roomOrder {
		if indices, ok := rooms[room]; ok {
			panel := m.renderRoomPanel(room, indices)
			panels = append(panels, panel)
		}
	}

	// Arrange in grid (3 per row)
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

func (m Model) renderRoomPanel(room string, indices []int) string {
	width := 26

	var lines []string

	lines = append(lines, PadLine(" "+Accent.Render(room), width-2))
	lines = append(lines, PadLine(" "+Dim.Render(strings.Repeat("─", width-4)), width-2))

	for _, i := range indices {
		d := m.HomeDevices[i]
		status := getDeviceIcon(d.Status)
		name := d.Name
		if len(name) > 18 {
			name = name[:18]
		}

		line := fmt.Sprintf(" %s %-18s", status, name)
		if i == m.SelectedDevice {
			line = Selected.Render(line)
		}
		lines = append(lines, PadLine(line, width-2))
	}

	content := strings.Join(lines, "\n")
	return NewBox(width).Render(content) + " "
}

func getDeviceIcon(status string) string {
	switch status {
	case "on":
		return Online.Render("●")
	case "off":
		return Offline.Render("○")
	case "locked":
		return Accent.Render("◆")
	case "unlocked":
		return Warning.Render("◇")
	case "closed":
		return Accent.Render("■")
	case "open":
		return Warning.Render("□")
	default:
		return Label.Render("?")
	}
}
