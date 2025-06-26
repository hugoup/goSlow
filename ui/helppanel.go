package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

var HelpOptions = []struct{ Key, Desc string }{
	{"↑/↓", "Scroll"},
	{"↵", "Show Queries"},
	{"Tab", "Switch panel"},
	{"l", "Sort"},
	{"s", "Save queries"},
	{"z", "Zoom"},
	{"h", "HL-mode"},
	{"q", "Quit"},
}

func RenderHelpPanel(highlightMode int, panelWidth int, status string, statusColor lipgloss.Color) string {
	highlightStatus := "[h] Highlight: "
	switch highlightMode {
	case 1:
		highlightStatus += "ON"
	case 0:
		highlightStatus += "OFF"
	}

	helpParts := make([]string, len(HelpOptions))
	for i, opt := range HelpOptions {
		helpParts[i] = fmt.Sprintf("[%s] %s", opt.Key, opt.Desc)
	}
	helpText := strings.Join(helpParts, "  ") + "  " + highlightStatus
	if status == "" {
		status = ""
		statusColor = ""
	}
	space := panelWidth - lipgloss.Width(helpText) - lipgloss.Width(status) - 4 // 4 for border padding
	if space < 1 {
		space = 1
	}
	statusStyled := lipgloss.NewStyle().Foreground(statusColor).Render(status)
	helpLine := helpText + strings.Repeat(" ", space) + statusStyled
	return lipgloss.NewStyle().Border(lipgloss.NormalBorder()).Width(panelWidth).Height(1).Render(helpLine)
}
