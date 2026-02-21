package ui

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"monoview/pkg/concentrator"
)

// System nodes, ping/pong, fire alert.

func (m *Model) handleFireAlert(msg concentrator.Message) {
	if strings.ToUpper(msg.From) != "ACHTUNG" || strings.ToUpper(msg.Verb) != "FIRE" {
		return
	}
	noun := strings.ToUpper(msg.Noun)
	if noun != "TIMER" && noun != "ALARM" {
		return
	}
	if len(msg.Args) < 1 {
		return
	}
	m.FireAlert = FireAlert{Show: true, JobKind: noun, JobName: msg.Args[0]}
	m.requestAchtungList() // refresh list (fired job is removed)
}

func (m *Model) dismissFireAlert() {
	m.HubSend("VERTEX", "OFF", "BUZZ")
	m.FireAlert.Show = false
	m.requestAchtungList() // refresh list after dismiss
}

// systemNodeGridUp/Down/Left/Right move SelectedNode in a 2-column grid.
// Col 0 = indices 0..mid-1, Col 1 = mid..n-1. Arrows: up/down within column, left/right between columns.
func (m *Model) systemNodeGridUp() {
	if len(m.Nodes) == 0 {
		return
	}
	mid := (len(m.Nodes) + 1) / 2
	if m.SelectedNode < mid {
		// Col 0: up = index--
		if m.SelectedNode > 0 {
			m.SelectedNode--
		}
	} else {
		// Col 1: up = index--
		if m.SelectedNode > mid {
			m.SelectedNode--
		}
	}
}

func (m *Model) systemNodeGridDown() {
	if len(m.Nodes) == 0 {
		return
	}
	mid := (len(m.Nodes) + 1) / 2
	if m.SelectedNode < mid {
		// Col 0: down = index++
		if m.SelectedNode < mid-1 {
			m.SelectedNode++
		}
	} else {
		// Col 1: down = index++
		if m.SelectedNode < len(m.Nodes)-1 {
			m.SelectedNode++
		}
	}
}

func (m *Model) systemNodeGridLeft() {
	if len(m.Nodes) == 0 {
		return
	}
	mid := (len(m.Nodes) + 1) / 2
	if m.SelectedNode >= mid {
		// Col 1 -> Col 0. Same row, clamp to col0 max.
		row := m.SelectedNode - mid
		if row < mid {
			m.SelectedNode = row
		} else {
			m.SelectedNode = mid - 1
		}
	}
}

func (m *Model) systemNodeGridRight() {
	if len(m.Nodes) == 0 {
		return
	}
	mid := (len(m.Nodes) + 1) / 2
	if m.SelectedNode < mid {
		// Col 0 -> Col 1. Same row, clamp to col1 max.
		row := m.SelectedNode
		col1Len := len(m.Nodes) - mid
		if row < col1Len {
			m.SelectedNode = mid + row
		} else {
			m.SelectedNode = len(m.Nodes) - 1
		}
	}
}

func (m *Model) selectedNodeIsAchtung() bool {
	if m.ActiveSheet != SheetSystem || m.SelectedNode >= len(m.Nodes) {
		return false
	}
	return strings.ToUpper(m.Nodes[m.SelectedNode].Name) == "ACHTUNG"
}

func (m *Model) pollNodes() {
	now := m.LastUpdate
	for i := range m.Nodes {
		node := &m.Nodes[i]

		if !node.LastSeen.IsZero() && now.Sub(node.LastSeen) > pingInterval*3 {
			node.Status = "offline"
			node.PingMs = 0
		}

		if node.PingSent.IsZero() || now.Sub(node.PingSent) >= pingInterval {
			m.pingNode(node)
		}
	}
}

func (m *Model) pingNode(node *SystemNode) {
	node.PingSent = time.Now()
	m.HubSend(node.Name, "PING", node.PingNoun)
	m.HubSend(node.Name, "GET", "UPTIME")
}

func (m *Model) pingSelectedNode() {
	if m.ActiveSheet != SheetSystem {
		return
	}
	if m.SelectedNode >= len(m.Nodes) {
		return
	}
	m.pingNode(&m.Nodes[m.SelectedNode])
}

func (m *Model) handleNodeResponse(msg concentrator.Message) {
	verb := strings.ToUpper(msg.Verb)
	from := strings.ToUpper(msg.From)

	for i := range m.Nodes {
		node := &m.Nodes[i]
		if node.Name != from {
			continue
		}

		switch verb {
		case "PONG":
			now := time.Now()
			node.Status = "online"
			node.LastSeen = now
			if !node.PingSent.IsZero() {
				node.PingMs = now.Sub(node.PingSent).Milliseconds()
			}
			if from == "GOVERNOR" {
				m.requestGovernorEvents()
				m.requestGovernorDeadlines()
			}

		case "OK":
			topic := strings.ToUpper(msg.Noun)
			if topic == "UPTIME" && len(msg.Args) >= 1 {
				node.Uptime = parseUptime(msg.Args[0])
			}
		}
		return
	}
}

func parseUptime(raw string) string {
	if ms, err := strconv.ParseInt(raw, 10, 64); err == nil {
		return formatDuration(time.Duration(ms) * time.Millisecond)
	}
	if d, err := time.ParseDuration(raw); err == nil {
		return formatDuration(d)
	}
	return raw
}

func (m *Model) visibleLogLines() int {
	h := m.plainHeight() - 3 // logs header lines
	if h < 5 {
		return 5
	}
	return h
}

// scrollLogsUp moves viewport to older logs (increases offset).
func (m *Model) scrollLogsUp() {
	visibleLogLines := m.visibleLogLines()
	maxOffset := len(m.Logs) - visibleLogLines
	if maxOffset < 0 {
		maxOffset = 0
	}
	m.LogScrollOffset++
	if m.LogScrollOffset > maxOffset {
		m.LogScrollOffset = maxOffset
	}
}

// scrollLogsDown moves viewport to newer logs (decreases offset).
func (m *Model) scrollLogsDown() {
	m.LogScrollOffset--
	if m.LogScrollOffset < 0 {
		m.LogScrollOffset = 0
	}
}

func formatDuration(d time.Duration) string {
	days := int(d.Hours()) / 24
	hours := int(d.Hours()) % 24
	mins := int(d.Minutes()) % 60
	secs := int(d.Seconds()) % 60

	if days > 0 {
		return fmt.Sprintf("%dd %dh %dm", days, hours, mins)
	}
	if hours > 0 {
		return fmt.Sprintf("%dh %dm %ds", hours, mins, secs)
	}
	if mins > 0 {
		return fmt.Sprintf("%dm %ds", mins, secs)
	}
	return fmt.Sprintf("%ds", secs)
}
