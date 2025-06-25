package ui

import (
	"fmt"
	"slowlog-tui/types"
	"strings"

	"github.com/charmbracelet/bubbles/viewport"
	"github.com/charmbracelet/lipgloss"
)

// PreviewPanel handles the SQL preview/viewport logic
// Stateless; state is managed by the main Model
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
