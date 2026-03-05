package app

import (
	"strings"

	"monoview/internal/types"
	"monoview/internal/ui"
)

// Form rendering: add event, event details, timer/alarm.

func (m Model) renderEventAddForm() string {
	return m.renderEventAddFormInner(0)
}

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
	lines = append(lines, ui.Title.Render("  New event")+" ")
	lines = append(lines, "")
	for i := 0; i < 6; i++ {
		line := ui.Label.Render("  "+labels[i]+": ") + ui.Value.Render(values[i])
		if i == focus {
			line += ui.Dim.Render("▌")
		}
		lines = append(lines, line)
	}
	lines = append(lines, "")
	lines = append(lines, ui.Label.Render("  [Esc] cancel  [Tab] next  [Enter] submit")+" ")
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
	box := ui.NewBox(width).WithBorderColor(ui.GruvAqua).WithTitle(" ADD EVENT ")
	return box.Render(inner)
}

func (m Model) renderEventDetailView(e types.Event, minHeight int) string {
	const width = 64
	var lines []string
	if e.ID == "" {
		lines = append(lines, "")
		lines = append(lines, ui.Label.Render("  No event selected"))
		lines = append(lines, "")
		lines = append(lines, ui.Dim.Render("  [Esc] close"))
	} else {
		lines = append(lines, "")
		lines = append(lines, ui.Title.Render("  Event details")+" ")
		lines = append(lines, "")
		lines = append(lines, ui.Label.Render("  Title: ")+ui.Value.Render(e.Title))
		lines = append(lines, ui.Label.Render("  Date: ")+ui.Value.Render(e.Date.Format("2006-01-02")))
		lines = append(lines, ui.Label.Render("  Time: ")+ui.Value.Render(e.Date.Format("15:04")))
		lines = append(lines, ui.Label.Render("  Category: ")+getCategoryIcon(e.Category)+" "+ui.Value.Render(e.Category))
		lines = append(lines, ui.Label.Render("  Location: ")+ui.Value.Render(e.Location))
		lines = append(lines, ui.Label.Render("  Notes: ")+ui.Value.Render(e.Notes))
		lines = append(lines, "")
		lines = append(lines, ui.Dim.Render("  [d] delete  [Esc] close"))
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
	box := ui.NewBox(width).WithBorderColor(ui.GruvAqua).WithTitle(" EVENT ")
	return box.Render(inner)
}

func (m Model) renderAchtungFormBox(minHeight int) string {
	const width = 64
	var lines []string
	lines = append(lines, "")
	if m.AchtungTimerMenu {
		lines = append(lines, ui.Title.Render("  New timer")+" ")
		lines = append(lines, "")
		for i, label := range []string{"Duration (e.g. 5m, 1h)", "Name (optional)"} {
			val := m.AchtungTimerDuration
			if i == 1 {
				val = m.AchtungTimerName
			}
			line := ui.Label.Render("  "+label+": ") + ui.Value.Render(val)
			if i == m.AchtungTimerFocusField {
				line += ui.Dim.Render("▌")
			}
			lines = append(lines, line)
		}
		lines = append(lines, "")
		lines = append(lines, ui.Dim.Render("  [Tab] next  [Enter] submit  [Esc] cancel"))
	} else if m.AchtungAlarmMenu {
		lines = append(lines, ui.Title.Render("  New alarm")+" ")
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
			line := ui.Label.Render("  "+label+": ") + ui.Value.Render(val)
			if i == m.AchtungAlarmFocusField {
				line += ui.Dim.Render("▌")
			}
			lines = append(lines, line)
		}
		lines = append(lines, "")
		lines = append(lines, ui.Dim.Render("  [Tab] next  [Enter] submit  [Esc] cancel"))
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
	box := ui.NewBox(width).WithBorderColor(ui.GruvAqua).WithTitle(title)
	return box.Render(inner)
}

func (m Model) renderAchtungJobDetailView(j types.AchtungJob, minHeight int) string {
	const width = 64
	var lines []string
	if j.Name == "" {
		lines = append(lines, "")
		lines = append(lines, ui.Label.Render("  No timer or alarm selected"))
		lines = append(lines, "")
		lines = append(lines, ui.Dim.Render("  [Esc] close"))
	} else {
		lines = append(lines, "")
		lines = append(lines, ui.Title.Render("  "+j.Kind)+" ")
		lines = append(lines, "")
		lines = append(lines, ui.Label.Render("  Name: ")+ui.Value.Render(j.Name))
		lines = append(lines, ui.Label.Render("  Remaining: ")+ui.Value.Render(j.Remaining))
		if j.Due != "" && j.Due != "—" {
			lines = append(lines, ui.Label.Render("  Due: ")+ui.Value.Render(j.Due))
		}
		lines = append(lines, "")
		lines = append(lines, ui.Dim.Render("  [d] stop/delete  [Esc] close"))
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
	box := ui.NewBox(width).WithBorderColor(ui.GruvAqua).WithTitle(title)
	return box.Render(inner)
}
