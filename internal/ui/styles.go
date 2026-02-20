package ui

import "github.com/charmbracelet/lipgloss"

// Gruvbox dark palette
var (
	GruvBg     = lipgloss.Color("#282828")
	GruvBg1    = lipgloss.Color("#3c3836")
	GruvBg2    = lipgloss.Color("#504945")
	GruvFg     = lipgloss.Color("#ebdbb2")
	GruvFg0    = lipgloss.Color("#fbf1c7")
	GruvGray   = lipgloss.Color("#928374")
	GruvRed    = lipgloss.Color("#fb4934")
	GruvGreen  = lipgloss.Color("#b8bb26")
	GruvYellow = lipgloss.Color("#fabd2f")
	GruvBlue   = lipgloss.Color("#83a598")
	GruvPurple = lipgloss.Color("#d3869b")
	GruvAqua   = lipgloss.Color("#8ec07c")
	GruvOrange = lipgloss.Color("#fe8019")
)

// Text styles
var (
	TabActive = lipgloss.NewStyle().
			Foreground(GruvBg).
			Background(GruvYellow).
			Bold(true).
			Padding(0, 2)

	TabInactive = lipgloss.NewStyle().
			Foreground(GruvGray).
			Background(GruvBg1).
			Padding(0, 2)

	Label = lipgloss.NewStyle().
		Foreground(GruvGray)

	Value = lipgloss.NewStyle().
		Foreground(GruvFg).
		Bold(true)

	Title = lipgloss.NewStyle().
		Foreground(GruvYellow).
		Bold(true)

	Online = lipgloss.NewStyle().
		Foreground(GruvGreen).
		Bold(true)

	Offline = lipgloss.NewStyle().
		Foreground(GruvRed).
		Bold(true)

	Warning = lipgloss.NewStyle().
		Foreground(GruvYellow).
		Bold(true)

	Help = lipgloss.NewStyle().
		Foreground(GruvGray)

	Dim = lipgloss.NewStyle().
		Foreground(GruvBg2)

	Accent = lipgloss.NewStyle().
		Foreground(GruvAqua)

	Highlight = lipgloss.NewStyle().
			Foreground(GruvPurple).
			Bold(true)

	Selected = lipgloss.NewStyle().
			Background(GruvBg2).
			Foreground(GruvFg0)

	// NodeHeaderSelected highlights the active node (VERTEX or ACHTUNG) on Home
	NodeHeaderSelected = lipgloss.NewStyle().
				Foreground(GruvYellow).
				Bold(true).
				BorderStyle(lipgloss.NormalBorder()).
				BorderForeground(GruvYellow).
				BorderLeft(true).
				Padding(0, 1)
)
