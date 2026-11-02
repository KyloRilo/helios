package main

import "github.com/charmbracelet/lipgloss"

var (
	colorPrimary   = lipgloss.Color("#FF6F00")
	colorSecondary = lipgloss.Color("#FFA726")
	colorAccent    = lipgloss.Color("#FFCC02")
	colorSuccess   = lipgloss.Color("#66BB6A")
	colorDanger    = lipgloss.Color("#EF5350")
	colorWarning   = lipgloss.Color("#FFA726")
	colorMuted     = lipgloss.Color("#888888")
	colorBg        = lipgloss.Color("#1A1A2E")
	colorBgLight   = lipgloss.Color("#16213E")
	colorWhite     = lipgloss.Color("#EEEEEE")

	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(colorAccent).
			MarginBottom(1)

	subtitleStyle = lipgloss.NewStyle().
			Foreground(colorSecondary).
			Italic(true)

	menuItemStyle = lipgloss.NewStyle().
			PaddingLeft(2)

	selectedMenuStyle = lipgloss.NewStyle().
				PaddingLeft(1).
				Foreground(colorAccent).
				Bold(true)

	statusRunning = lipgloss.NewStyle().
			Foreground(colorSuccess).
			Bold(true)

	statusStopped = lipgloss.NewStyle().
			Foreground(colorMuted)

	statusError = lipgloss.NewStyle().
			Foreground(colorDanger).
			Bold(true)

	statusCreating = lipgloss.NewStyle().
			Foreground(colorWarning)

	boxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(colorPrimary).
			Padding(1, 2)

	headerStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(colorPrimary).
			BorderBottom(true).
			BorderStyle(lipgloss.NormalBorder()).
			BorderForeground(colorMuted).
			MarginBottom(1)

	helpStyle = lipgloss.NewStyle().
			Foreground(colorMuted).
			MarginTop(1)

	metricLabel = lipgloss.NewStyle().
			Foreground(colorMuted).
			Width(12)

	metricValue = lipgloss.NewStyle().
			Foreground(colorWhite).
			Bold(true)

	barFilled = lipgloss.NewStyle().
			Foreground(colorSuccess)

	barEmpty = lipgloss.NewStyle().
			Foreground(colorMuted)

	barWarning = lipgloss.NewStyle().
			Foreground(colorWarning)

	barDanger = lipgloss.NewStyle().
			Foreground(colorDanger)

	tableHeaderStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(colorAccent).
				BorderBottom(true).
				BorderStyle(lipgloss.NormalBorder()).
				BorderForeground(colorMuted)

	tableRowStyle = lipgloss.NewStyle().
			Foreground(colorWhite)

	selectedRowStyle = lipgloss.NewStyle().
				Foreground(colorAccent).
				Bold(true)

	dashboardCard = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(colorPrimary).
			Padding(0, 1).
			Width(36)

	inputStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(colorSecondary).
			Padding(0, 1)

	successMsg = lipgloss.NewStyle().
			Foreground(colorSuccess).
			Bold(true)

	errorMsg = lipgloss.NewStyle().
			Foreground(colorDanger).
			Bold(true)
)

func styledStatus(s ServiceStatus) string {
	switch s {
	case StatusRunning:
		return statusRunning.Render("● Running")
	case StatusStopped:
		return statusStopped.Render("○ Stopped")
	case StatusError:
		return statusError.Render("✗ Error")
	case StatusCreating:
		return statusCreating.Render("◌ Creating")
	default:
		return string(s)
	}
}
