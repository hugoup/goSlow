package ui

import (
	"fmt"
	"strings"

	"slowlog-tui/types"

	"github.com/charmbracelet/bubbles/viewport"
	"github.com/charmbracelet/lipgloss"
)

// PreviewPanel handles the SQL preview/viewport logic
func NewPreviewPanel(g types.GroupedQuery, highlightMode int, width, height int) viewport.Model {
	header := fmt.Sprintf("%s | %d queries | Avg: %.2fs, %.0f rows examined, %.0f sent\n\n",
		lipgloss.NewStyle().Bold(true).Render(g.QueryType),
		g.Count,
		g.AvgQueryTime,
		g.AvgRowsExamined,
		g.AvgRowsSent,
	)
	var allQueries strings.Builder
	for i, q := range g.Examples {
		allQueries.WriteString(q.SQLText)
		if i < len(g.Examples)-1 {
			allQueries.WriteString("\n---\n")
		}
	}
	var content string
	switch highlightMode {
	case 1: // HighlightSimple
		content = HighlightSQL(allQueries.String())
	case 0: // HighlightOff
		content = allQueries.String()
	}
	vp := viewport.New(width, height)
	vp.SetContent(header + content)
	return vp
}

// RenderZoomedPreviewView renders the zoomed preview panel
func RenderZoomedPreviewView(m Model) string {
	panelWidth := m.viewport.Width
	zoomBoxStyle := rightStyle.BorderForeground(activeBorder)
	m.viewport.Height = m.height - 3 // use all available rows minus help/status line
	zoomBox := zoomBoxStyle.Width(panelWidth).Height(m.viewport.Height).Render(m.viewport.View())
	// Help/status line
	helpText := "[z] Unzoom  [h] Highlight  [q] Quit"
	status := m.statusText
	statusColor := m.statusColor
	if status == "" {
		status = ""
		statusColor = ""
	}
	space := panelWidth - lipgloss.Width(helpText) - lipgloss.Width(status) - 4
	if space < 1 {
		space = 1
	}
	statusStyled := lipgloss.NewStyle().Foreground(statusColor).Render(status)
	helpLine := helpText + strings.Repeat(" ", space) + statusStyled
	helpBox := lipgloss.NewStyle().Border(lipgloss.NormalBorder()).Width(panelWidth).Height(1).Render(helpLine)
	return appStyle.Margin(0, 0).Render(zoomBox + "\n" + helpBox)
}

// RenderMainUIView renders the main UI (table, preview, help)
func RenderMainUIView(m Model) string {
	panelWidth := m.viewport.Width
	tableContent := m.table.View()
	var tableBoxStyle, sqlBoxStyle lipgloss.Style
	if m.focus == focusTable {
		tableBoxStyle = leftStyle.BorderForeground(activeBorder)
		sqlBoxStyle = rightStyle.BorderForeground(inactiveBorder)
	} else {
		tableBoxStyle = leftStyle.BorderForeground(inactiveBorder)
		sqlBoxStyle = rightStyle.BorderForeground(activeBorder)
	}
	tableBox := tableBoxStyle.Width(panelWidth).Render(tableContent)
	sqlBox := sqlBoxStyle.Width(panelWidth).Render(m.viewport.View())

	helpBox := RenderHelpPanel(int(m.highlightMode), panelWidth, m.statusText, m.statusColor)

	return appStyle.Margin(0, 0).Render(
		tableBox + "\n" + sqlBox + "\n" + helpBox,
	)
}
