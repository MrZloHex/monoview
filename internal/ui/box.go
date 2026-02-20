package ui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// Box draws a proper box with consistent borders
type Box struct {
	Width       int
	BorderColor lipgloss.Color
	Title       string
}

func NewBox(width int) *Box {
	return &Box{
		Width:       width,
		BorderColor: GruvBg2,
	}
}

func (b *Box) WithBorderColor(c lipgloss.Color) *Box {
	b.BorderColor = c
	return b
}

func (b *Box) WithTitle(t string) *Box {
	b.Title = t
	return b
}

func (b *Box) Render(content string) string {
	borderStyle := lipgloss.NewStyle().Foreground(b.BorderColor)

	innerWidth := b.Width - 2 // account for left and right borders

	// Top border
	var top string
	if b.Title != "" {
		titleLen := lipgloss.Width(b.Title)
		if titleLen > innerWidth-2 {
			titleLen = innerWidth - 2
		}
		lineLen := innerWidth - titleLen // total dashes so that ┌ + dashes + title + ┐ = Width
		leftLine := 1
		rightLine := lineLen - leftLine
		if rightLine < 0 {
			rightLine = 0
		}
		top = borderStyle.Render("┌"+strings.Repeat("─", leftLine)) +
			Title.Render(b.Title) +
			borderStyle.Render(strings.Repeat("─", rightLine)+"┐")
	} else {
		top = borderStyle.Render("┌" + strings.Repeat("─", innerWidth) + "┐")
	}

	// Bottom border
	bottom := borderStyle.Render("└" + strings.Repeat("─", innerWidth) + "┘")

	// Content lines
	lines := strings.Split(content, "\n")
	var rendered []string
	rendered = append(rendered, top)

	for _, line := range lines {
		lineWidth := lipgloss.Width(line)
		padding := innerWidth - lineWidth
		if padding < 0 {
			// Truncate if too long
			line = truncateString(line, innerWidth)
			padding = 0
		}
		rendered = append(rendered,
			borderStyle.Render("│")+line+strings.Repeat(" ", padding)+borderStyle.Render("│"))
	}

	rendered = append(rendered, bottom)
	return strings.Join(rendered, "\n")
}

// truncateString truncates a string to fit within maxWidth (accounting for unicode)
func truncateString(s string, maxWidth int) string {
	if lipgloss.Width(s) <= maxWidth {
		return s
	}

	// Truncate rune by rune
	result := ""
	for _, r := range s {
		test := result + string(r)
		if lipgloss.Width(test) > maxWidth-1 {
			return result + "…"
		}
		result = test
	}
	return result
}

// PadLine pads a line to exact width
func PadLine(s string, width int) string {
	currentWidth := lipgloss.Width(s)
	if currentWidth >= width {
		return truncateString(s, width)
	}
	return s + strings.Repeat(" ", width-currentWidth)
}
