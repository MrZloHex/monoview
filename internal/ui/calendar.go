package ui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
)

func (m Model) renderCalendar() string {
	// Left side: mini calendar + events + deadlines
	cal := m.renderMiniCalendar()
	events := m.renderEventList()
	deadlines := m.renderDeadlines()

	leftPanel := lipgloss.JoinVertical(lipgloss.Left, cal, "", events, "", deadlines)

	// Right side: university schedule
	schedule := m.renderSchedule()

	content := lipgloss.JoinHorizontal(lipgloss.Top, leftPanel, "   ", schedule)

	// Indent all lines, not just the first
	return indentLines(content, "  ")
}

func (m Model) renderMiniCalendar() string {
	width := 24

	var lines []string

	// Title
	titleText := m.SelectedDate.Format("January 2006")
	titlePadded := PadLine("  "+Title.Render(titleText), width-2)
	lines = append(lines, titlePadded)
	lines = append(lines, "")

	// Weekday headers
	days := []string{"Mo", "Tu", "We", "Th", "Fr", "Sa", "Su"}
	var header string
	for _, d := range days {
		header += Label.Render(d) + " "
	}
	lines = append(lines, PadLine(header, width-2))

	// Calendar grid
	firstOfMonth := time.Date(m.SelectedDate.Year(), m.SelectedDate.Month(), 1, 0, 0, 0, 0, m.SelectedDate.Location())
	lastOfMonth := firstOfMonth.AddDate(0, 1, -1)

	offset := int(firstOfMonth.Weekday())
	if offset == 0 {
		offset = 7
	}
	offset--

	today := time.Now()

	row := strings.Repeat("   ", offset)

	for day := 1; day <= lastOfMonth.Day(); day++ {
		currentDate := time.Date(m.SelectedDate.Year(), m.SelectedDate.Month(), day, 0, 0, 0, 0, m.SelectedDate.Location())
		dayStr := fmt.Sprintf("%2d", day)

		if currentDate.YearDay() == m.SelectedDate.YearDay() && currentDate.Year() == m.SelectedDate.Year() {
			row += lipgloss.NewStyle().Background(GruvYellow).Foreground(GruvBg).Render(dayStr)
		} else if currentDate.YearDay() == today.YearDay() && currentDate.Year() == today.Year() {
			row += Accent.Render(dayStr)
		} else if m.hasEvent(currentDate) {
			row += Highlight.Render(dayStr)
		} else {
			row += Value.Render(dayStr)
		}
		row += " "

		if (offset+day)%7 == 0 {
			lines = append(lines, PadLine(row, width-2))
			row = ""
		}
	}

	if row != "" {
		lines = append(lines, PadLine(row, width-2))
	}

	content := strings.Join(lines, "\n")
	return NewBox(width).Render(content)
}

func (m Model) hasEvent(date time.Time) bool {
	for _, e := range m.Events {
		if e.Date.YearDay() == date.YearDay() && e.Date.Year() == date.Year() {
			return true
		}
	}
	return false
}

func (m Model) renderEventList() string {
	width := 40

	var lines []string

	titleText := "EVENTS: " + m.SelectedDate.Format("02 Jan")
	lines = append(lines, PadLine(" "+Title.Render(titleText), width-2))
	lines = append(lines, "")

	found := false
	for _, e := range m.Events {
		if e.Date.YearDay() == m.SelectedDate.YearDay() && e.Date.Year() == m.SelectedDate.Year() {
			found = true
			timeStr := e.Date.Format("15:04")
			cat := getCategoryIcon(e.Category)
			line := fmt.Sprintf(" %s  %s  %s",
				Label.Render(timeStr),
				cat,
				Value.Render(e.Title))
			lines = append(lines, PadLine(line, width-2))
		}
	}

	if !found {
		lines = append(lines, PadLine(" "+Label.Render("No events scheduled"), width-2))
	}

	content := strings.Join(lines, "\n")
	return NewBox(width).Render(content)
}

func getCategoryIcon(cat string) string {
	switch cat {
	case "work":
		return lipgloss.NewStyle().Foreground(GruvBlue).Render("●")
	case "personal":
		return lipgloss.NewStyle().Foreground(GruvGreen).Render("●")
	case "deadline":
		return lipgloss.NewStyle().Foreground(GruvRed).Render("●")
	case "system":
		return lipgloss.NewStyle().Foreground(GruvPurple).Render("●")
	default:
		return lipgloss.NewStyle().Foreground(GruvGray).Render("●")
	}
}

func (m Model) renderDeadlines() string {
	width := 40

	var lines []string

	lines = append(lines, PadLine(" "+Title.Render("UPCOMING DEADLINES"), width-2))
	lines = append(lines, "")

	count := 0
	for _, e := range m.Events {
		if e.Category == "deadline" && e.Date.After(time.Now()) {
			days := int(e.Date.Sub(time.Now()).Hours() / 24)
			var daysStr string
			if days == 0 {
				daysStr = Warning.Render("TODAY")
			} else if days == 1 {
				daysStr = Warning.Render("  1d ")
			} else {
				daysStr = Label.Render(fmt.Sprintf("%3dd ", days))
			}

			line := fmt.Sprintf(" %s %s", daysStr, Value.Render(e.Title))
			lines = append(lines, PadLine(line, width-2))
			count++
			if count >= 3 {
				break
			}
		}
	}

	if count == 0 {
		lines = append(lines, PadLine(" "+Label.Render("No upcoming deadlines"), width-2))
	}

	content := strings.Join(lines, "\n")
	return NewBox(width).Render(content)
}

// ═══════════════════════════════════════════════════════════════════════════════
// SCHEDULE VIEW
// ═══════════════════════════════════════════════════════════════════════════════

func (m Model) renderSchedule() string {
	width := 44

	var lines []string

	// Header with weekday name
	weekdayName := m.SelectedDate.Weekday().String()
	header := fmt.Sprintf(" %s  %s",
		Title.Render("SCHEDULE"),
		Accent.Render(weekdayName))
	lines = append(lines, PadLine(header, width-2))
	lines = append(lines, PadLine(" "+Dim.Render(strings.Repeat("─", width-4)), width-2))

	// Get entries for selected weekday
	entries := m.getScheduleForDay(m.SelectedDate.Weekday())

	if len(entries) == 0 {
		lines = append(lines, "")
		lines = append(lines, PadLine(" "+Label.Render("No classes scheduled"), width-2))
		lines = append(lines, "")
	} else {
		now := m.LastUpdate
		for _, e := range entries {
			lines = append(lines, "")
			entryLines := m.renderScheduleEntry(e, width-2, now)
			lines = append(lines, entryLines...)
		}
		lines = append(lines, "")
	}

	content := strings.Join(lines, "\n")
	return NewBox(width).Render(content)
}

func (m Model) getScheduleForDay(weekday time.Weekday) []ScheduleEntry {
	var entries []ScheduleEntry
	for _, e := range m.Schedule {
		if e.Weekday == weekday {
			entries = append(entries, e)
		}
	}
	return entries
}

func (m Model) renderScheduleEntry(e ScheduleEntry, width int, now time.Time) []string {
	var lines []string

	// Check if current
	isCurrent := m.isCurrentClass(e, now)

	// Time range
	timeStr := fmt.Sprintf("%s-%s", e.Start, e.End)

	// Tags as colored badges
	var tagBadges []string
	for _, tag := range e.Tags {
		badge := renderTagBadge(tag)
		tagBadges = append(tagBadges, badge)
	}
	tagsStr := strings.Join(tagBadges, " ")

	// First line: indicator + time + tags
	indicator := " "
	if isCurrent {
		indicator = lipgloss.NewStyle().Foreground(GruvOrange).Bold(true).Render("▶")
	}

	line1 := fmt.Sprintf(" %s %s  %s", indicator, Label.Render(timeStr), tagsStr)
	lines = append(lines, PadLine(line1, width))

	// Second line: title
	titleStyle := Value
	if isCurrent {
		titleStyle = lipgloss.NewStyle().Foreground(GruvOrange).Bold(true)
	}
	line2 := fmt.Sprintf("   %s", titleStyle.Render(e.Title))
	lines = append(lines, PadLine(line2, width))

	// Third line: location
	line3 := fmt.Sprintf("   %s %s", Label.Render("@"), Accent.Render(e.Location))
	lines = append(lines, PadLine(line3, width))

	return lines
}

func (m Model) isCurrentClass(e ScheduleEntry, now time.Time) bool {
	// Only check if same weekday and same date
	if now.Weekday() != e.Weekday {
		return false
	}
	if now.YearDay() != m.SelectedDate.YearDay() || now.Year() != m.SelectedDate.Year() {
		return false
	}

	// Parse times
	startParts := strings.Split(e.Start, ":")
	endParts := strings.Split(e.End, ":")
	if len(startParts) != 2 || len(endParts) != 2 {
		return false
	}

	var startH, startM, endH, endM int
	fmt.Sscanf(startParts[0], "%d", &startH)
	fmt.Sscanf(startParts[1], "%d", &startM)
	fmt.Sscanf(endParts[0], "%d", &endH)
	fmt.Sscanf(endParts[1], "%d", &endM)

	nowMins := now.Hour()*60 + now.Minute()
	startMins := startH*60 + startM
	endMins := endH*60 + endM

	return nowMins >= startMins && nowMins <= endMins
}

func renderTagBadge(tag string) string {
	colorHex, ok := TagColors[tag]
	if !ok {
		return Label.Render("[" + tag + "]")
	}

	color := lipgloss.Color(colorHex)
	style := lipgloss.NewStyle().
		Foreground(GruvBg).
		Background(color).
		Bold(true)

	return style.Render(" " + tag + " ")
}
