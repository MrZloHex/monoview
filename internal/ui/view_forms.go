package ui

import (
	"strings"
)

// Form rendering: add event, event details, timer/alarm.

func (m Model) renderEventAddForm() string {
	return m.renderEventAddFormInner(0)
}

// renderEventAddFormInner builds the form box. If minHeight > 0, pads inner content so the box spans minHeight lines (anchored right, full length).
func (m Model) renderEventAddFormInner(minHeight int) string {
	const width = 64
	labels := []string{"Title", "Date (YYYY-MM-DD)", "Time (HH:MM)", "Location", "Notes", "Visible from (opt)"}
	values := []string{m.EventAddTitle, m.EventAddDate, m.EventAddTime, m.EventAddLocation, m.EventAddNotes, m.EventAddVisibleFrom}
	focus := m.EventAddFocusField
	if focus < 0 || focus > 5 {
		focus = 0
	}
	var lines []string
	lines = append(lines, "")
	lines = append(lines, Title.Render("  New event")+" ")
	lines = append(lines, "")
	for i := 0; i < 6; i++ {
		line := Label.Render("  "+labels[i]+": ") + Value.Render(values[i])
		if i == focus {
			line += Dim.Render("▌")
		}
		lines = append(lines, line)
	}
	lines = append(lines, "")
	lines = append(lines, Label.Render("  [Esc] cancel  [Tab] next  [Enter] submit")+" ")
	lines = append(lines, "")
	inner := strings.Join(lines, "\n")
	if minHeight > 2 {
		innerLines := strings.Split(inner, "\n")
		needLines := minHeight - 2
		for len(innerLines) < needLines {
			innerLines = append(innerLines, "")
		}
		if len(innerLines) > needLines {
			innerLines = innerLines[:needLines]
		}
		inner = strings.Join(innerLines, "\n")
	}
	box := NewBox(width).WithBorderColor(GruvAqua).WithTitle(" ADD EVENT ")
	return box.Render(inner)
}

// renderEventDetailView shows the selected event's info in a right-side box (full height of plain).
func (m Model) renderEventDetailView(e Event, minHeight int) string {
	const width = 64
	var lines []string
	if e.ID == "" {
		lines = append(lines, "")
		lines = append(lines, Label.Render("  No event selected"))
		lines = append(lines, "")
		lines = append(lines, Dim.Render("  [Esc] close"))
	} else {
		lines = append(lines, "")
		lines = append(lines, Title.Render("  Event details")+" ")
		lines = append(lines, "")
		lines = append(lines, Label.Render("  Title: ")+Value.Render(e.Title))
		lines = append(lines, Label.Render("  Date: ")+Value.Render(e.Date.Format("2006-01-02")))
		lines = append(lines, Label.Render("  Time: ")+Value.Render(e.Date.Format("15:04")))
		lines = append(lines, Label.Render("  Category: ")+getCategoryIcon(e.Category)+" "+Value.Render(e.Category))
		lines = append(lines, Label.Render("  Location: ")+Value.Render(e.Location))
		lines = append(lines, Label.Render("  Notes: ")+Value.Render(e.Notes))
		lines = append(lines, "")
		lines = append(lines, Dim.Render("  [d] delete  [Esc] close"))
	}
	inner := strings.Join(lines, "\n")
	if minHeight > 2 {
		innerLines := strings.Split(inner, "\n")
		needLines := minHeight - 2
		for len(innerLines) < needLines {
			innerLines = append(innerLines, "")
		}
		if len(innerLines) > needLines {
			innerLines = innerLines[:needLines]
		}
		inner = strings.Join(innerLines, "\n")
	}
	box := NewBox(width).WithBorderColor(GruvAqua).WithTitle(" EVENT ")
	return box.Render(inner)
}

// renderAchtungFormBox renders the timer or alarm creation form (all fields at once, like events).
func (m Model) renderAchtungFormBox(minHeight int) string {
	const width = 64
	var lines []string
	lines = append(lines, "")
	if m.AchtungTimerMenu {
		lines = append(lines, Title.Render("  New timer")+" ")
		lines = append(lines, "")
		for i, label := range []string{"Duration (e.g. 5m, 1h)", "Name (optional)"} {
			val := m.AchtungTimerDuration
			if i == 1 {
				val = m.AchtungTimerName
			}
			line := Label.Render("  "+label+": ") + Value.Render(val)
			if i == m.AchtungTimerFocusField {
				line += Dim.Render("▌")
			}
			lines = append(lines, line)
		}
		lines = append(lines, "")
		lines = append(lines, Dim.Render("  [Tab] next  [Enter] submit  [Esc] cancel"))
	} else if m.AchtungAlarmMenu {
		lines = append(lines, Title.Render("  New alarm")+" ")
		lines = append(lines, "")
		for i, label := range []string{"Date (YYYY-MM-DD)", "Time (HH:MM)", "Name (optional)"} {
			var val string
			switch i {
			case 0:
				val = m.AchtungAlarmDate
			case 1:
				val = m.AchtungAlarmTime
			case 2:
				val = m.AchtungAlarmName
			}
			line := Label.Render("  "+label+": ") + Value.Render(val)
			if i == m.AchtungAlarmFocusField {
				line += Dim.Render("▌")
			}
			lines = append(lines, line)
		}
		lines = append(lines, "")
		lines = append(lines, Dim.Render("  [Tab] next  [Enter] submit  [Esc] cancel"))
	}
	inner := strings.Join(lines, "\n")
	if minHeight > 2 {
		innerLines := strings.Split(inner, "\n")
		needLines := minHeight - 2
		for len(innerLines) < needLines {
			innerLines = append(innerLines, "")
		}
		if len(innerLines) > needLines {
			innerLines = innerLines[:needLines]
		}
		inner = strings.Join(innerLines, "\n")
	}
	title := " ADD TIMER "
	if m.AchtungAlarmMenu {
		title = " ADD ALARM "
	}
	box := NewBox(width).WithBorderColor(GruvAqua).WithTitle(title)
	return box.Render(inner)
}

// renderAchtungJobDetailView shows the selected timer/alarm info in the right panel.
func (m Model) renderAchtungJobDetailView(j AchtungJob, minHeight int) string {
	const width = 64
	var lines []string
	if j.Name == "" {
		lines = append(lines, "")
		lines = append(lines, Label.Render("  No timer or alarm selected"))
		lines = append(lines, "")
		lines = append(lines, Dim.Render("  [Esc] close"))
	} else {
		lines = append(lines, "")
		lines = append(lines, Title.Render("  "+j.Kind) + " ")
		lines = append(lines, "")
		lines = append(lines, Label.Render("  Name: ")+Value.Render(j.Name))
		lines = append(lines, Label.Render("  Remaining: ")+Value.Render(j.Remaining))
		if j.Due != "" && j.Due != "—" {
			lines = append(lines, Label.Render("  Due: ")+Value.Render(j.Due))
		}
		lines = append(lines, "")
		lines = append(lines, Dim.Render("  [d] stop/delete  [Esc] close"))
	}
	inner := strings.Join(lines, "\n")
	if minHeight > 2 {
		innerLines := strings.Split(inner, "\n")
		needLines := minHeight - 2
		for len(innerLines) < needLines {
			innerLines = append(innerLines, "")
		}
		if len(innerLines) > needLines {
			innerLines = innerLines[:needLines]
		}
		inner = strings.Join(innerLines, "\n")
	}
	title := " JOB "
	if j.Kind != "" {
		title = " " + j.Kind + " "
	}
	box := NewBox(width).WithBorderColor(GruvAqua).WithTitle(title)
	return box.Render(inner)
}
