package app

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"

	"monoview/internal/ui"
)

const synapseListWidth = 30

var synapseText = lipgloss.NewStyle().Foreground(ui.GruvFg)

func (m Model) renderSynapse() string {
	if !m.signedIn() {
		return ui.IndentLines(ui.Title.Render("▌SYNAPSE")+"\n\n"+
			ui.Dim.Render("SYNAPSE files messages under a person. Sign in at [5] PEOPLE to read and write."), "  ")
	}
	content := lipgloss.JoinHorizontal(lipgloss.Top, m.renderSynapseList(), "    ", m.renderSynapseConversation())
	return ui.IndentLines(content, "  ")
}

func (m Model) renderSynapseList() string {
	var b strings.Builder
	b.WriteString(ui.Title.Render("▌SYNAPSE") + "\n\n")
	peers := m.synapsePeers()
	if len(peers) == 0 {
		b.WriteString(ui.Dim.Render("Nobody to write to yet.") + "\n")
	}
	for i, peer := range peers {
		cursor, name := "  ", fmt.Sprintf("%-12s", peer)
		switch {
		case i == m.Synapse.selected:
			cursor, name = ui.Accent.Render("▸ "), ui.Accent.Render(name)
		case peer == m.Synapse.peer:
			name = ui.Highlight.Render(name)
		default:
			name = ui.Value.Render(name)
		}
		extra := ""
		if c := m.Synapse.chats[peer]; c.unread > 0 {
			extra = ui.Warning.Render(fmt.Sprintf("(%d)", c.unread))
		} else if !c.at.IsZero() {
			extra = ui.Dim.Render(synapseWhen(c.at, m.LastUpdate))
		}
		b.WriteString(cursor + name + " " + extra + "\n")
	}
	if m.Synapse.status != "" {
		b.WriteString("\n")
		for _, l := range wrapText(m.Synapse.status, synapseListWidth) {
			b.WriteString(ui.Offline.Render(l) + "\n")
		}
	}
	return lipgloss.NewStyle().Width(synapseListWidth).Render(b.String())
}

func (m Model) renderSynapseConversation() string {
	p := m.Synapse
	width := max(m.Width-synapseListWidth-10, 40)
	if p.peer == "" {
		body := "\n  " + ui.Dim.Render("Choose someone and press Enter.") + "\n"
		return ui.NewBox(width).WithTitle(" CONVERSATION ").WithDimTitle(true).Render(body)
	}
	inner := width - 4
	me := m.Session.User

	var lines []string
	for _, x := range p.msgs {
		who := ui.Accent.Render(x.from)
		receipt := ""
		if x.from == me {
			who = ui.Highlight.Render(x.from)
			receipt = ui.Dim.Render(" ·")
			if x.read {
				receipt = ui.Online.Render(" ✓")
			}
		}
		lines = append(lines, ui.Label.Render(synapseWhen(x.at, m.LastUpdate)+" ")+who+receipt)
		for _, l := range wrapText(x.text, inner-2) {
			lines = append(lines, "  "+synapseText.Render(l))
		}
	}
	if len(lines) == 0 {
		lines = append(lines, ui.Dim.Render("No messages yet."))
	}
	rows := max(m.plainHeight()-7, 4) // the box, the earlier-line, the input line
	if len(lines) > rows {
		lines = lines[len(lines)-rows:]
	}

	top := ""
	if p.older {
		top = ui.Dim.Render("[u] earlier")
	}
	input := ui.Dim.Render("[i] write")
	switch {
	case p.busy:
		input = ui.Dim.Render("sending…")
	case p.writing:
		input = ui.Accent.Render("> ") + synapseText.Render(tailFit(p.draft, inner-4)) + ui.Dim.Render("▌")
	}
	body := append([]string{top}, lines...)
	body = append(body, "", input)
	for i := range body {
		body[i] = " " + body[i]
	}
	return ui.NewBox(width).WithTitle(" " + p.peer + " ").Render(strings.Join(body, "\n"))
}

// synapseWhen is a message's time: the clock today, the date before.
func synapseWhen(t, now time.Time) string {
	t = t.Local()
	if y, d := t.Year(), t.YearDay(); y == now.Year() && d == now.YearDay() {
		return t.Format("15:04")
	}
	return t.Format("02 Jan 15:04")
}

// wrapText breaks s into lines at most width cells wide, at spaces where it
// can, and inside a word only when the word is longer than a line.
func wrapText(s string, width int) []string {
	width = max(width, 8)
	var lines []string
	line := ""
	for _, word := range strings.Fields(s) {
		for lipgloss.Width(word) > width {
			if line != "" {
				lines = append(lines, line)
				line = ""
			}
			head, rest := cutWidth(word, width)
			lines = append(lines, head)
			word = rest
		}
		switch {
		case word == "":
		case line == "":
			line = word
		case lipgloss.Width(line)+1+lipgloss.Width(word) <= width:
			line += " " + word
		default:
			lines = append(lines, line)
			line = word
		}
	}
	if line != "" {
		lines = append(lines, line)
	}
	return lines
}

func cutWidth(s string, width int) (string, string) {
	w := 0
	for i, r := range s {
		rw := lipgloss.Width(string(r))
		if w+rw > width {
			return s[:i], s[i:]
		}
		w += rw
	}
	return s, ""
}

// tailFit keeps the end of s, where the cursor is, within width cells.
func tailFit(s string, width int) string {
	if lipgloss.Width(s) <= width {
		return s
	}
	r := []rune(s)
	for len(r) > 0 && lipgloss.Width(string(r)) > width-1 {
		r = r[1:]
	}
	return "…" + string(r)
}
